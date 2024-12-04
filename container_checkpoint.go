package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/exec"
    "strings"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/ecr"
    "github.com/containerd/containerd"
    "github.com/containerd/containerd/namespaces"
)

func init() {
    log.SetOutput(os.Stdout)
    log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
}

func main() {
    if len(os.Args) < 4 {
        log.Fatal("Usage: checkpoint_container <pod_identifier> <ecr_repo> <aws_region>")
    }

    podID := os.Args[1]
    ecrRepo := os.Args[2]
    awsRegion := os.Args[3]

    log.Printf("Starting checkpoint process for pod %s", podID)

    containerID, err := getContainerIDFromPod(podID)
    if err != nil {
        log.Fatalf("Error getting container ID: %v", err)
    }

    err = processContainerCheckpoint(containerID, ecrRepo, awsRegion)
    if err != nil {
        log.Fatalf("Error processing container checkpoint: %v", err)
    }

    log.Printf("Successfully checkpointed container %s and pushed to ECR", containerID)
}

func getContainerIDFromPod(podID string) (string, error) {
    log.Printf("Searching for container ID for pod %s", podID)
    client, err := containerd.New("/run/containerd/containerd.sock")
    if err != nil {
        return "", fmt.Errorf("failed to connect to containerd: %v", err)
    }
    defer client.Close()

    ctx := namespaces.WithNamespace(context.Background(), "k8s.io")

    containers, err := client.Containers(ctx)
    if err != nil {
        return "", fmt.Errorf("failed to list containers: %v", err)
    }

    for _, container := range containers {
        info, err := container.Info(ctx)
        if err != nil {
            continue
        }
        if strings.Contains(info.Labels["io.kubernetes.pod.uid"], podID) {
            log.Printf("Found container ID %s for pod %s", container.ID(), podID)
            return container.ID(), nil
        }
    }

    return "", fmt.Errorf("container not found for pod %s", podID)
}

func processContainerCheckpoint(containerID, ecrRepo, region string) error {
    log.Printf("Processing checkpoint for container %s", containerID)
    checkpointPath, err := createCheckpoint(containerID)
    if err != nil {
        return err
    }
    defer os.RemoveAll(checkpointPath)

    imageName, err := convertCheckpointToImage(checkpointPath, ecrRepo, containerID)
    if err != nil {
        return err
    }

    err = pushImageToECR(imageName, region)
    if err != nil {
        return err
    }

    return nil
}

func createCheckpoint(containerID string) (string, error) {
    log.Printf("Creating checkpoint for container %s", containerID)
    checkpointPath := "/tmp/checkpoint-" + containerID
    cmd := exec.Command("ctr", "-n", "k8s.io", "tasks", "checkpoint", containerID, "--checkpoint-path", checkpointPath)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("checkpoint command failed: %v, output: %s", err, output)
    }
    log.Printf("Checkpoint created at: %s", checkpointPath)
    return checkpointPath, nil
}

func convertCheckpointToImage(checkpointPath, ecrRepo, containerID string) (string, error) {
    log.Printf("Converting checkpoint to image for container %s", containerID)
    imageName := ecrRepo + ":checkpoint-" + containerID

    cmd := exec.Command("buildah", "from", "scratch")
    containerId, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("failed to create container: %v", err)
    }

    cmd = exec.Command("buildah", "copy", string(containerId), checkpointPath, "/")
    err = cmd.Run()
    if err != nil {
        return "", fmt.Errorf("failed to copy checkpoint: %v", err)
    }

    cmd = exec.Command("buildah", "commit", string(containerId), imageName)
    err = cmd.Run()
    if err != nil {
        return "", fmt.Errorf("failed to commit image: %v", err)
    }

    log.Printf("Created image: %s", imageName)
    return imageName, nil
}

func pushImageToECR(imageName, region string) error {
    log.Printf("Pushing image %s to ECR in region %s", imageName, region)
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String(region),
    })
    if err != nil {
        return fmt.Errorf("failed to create AWS session: %v", err)
    }

    svc := ecr.New(sess)

    authToken, registryURL, err := getECRAuthorizationToken(svc)
    if err != nil {
        return err
    }

    err = loginToECR(authToken, registryURL)
    if err != nil {
        return err
    }

    cmd := exec.Command("podman", "push", imageName)
    err = cmd.Run()
    if err != nil {
        return fmt.Errorf("failed to push image to ECR: %v", err)
    }

    log.Printf("Successfully pushed checkpoint image to ECR: %s", imageName)
    return nil
}

func getECRAuthorizationToken(svc *ecr.ECR) (string, string, error) {
    log.Print("Getting ECR authorization token")
    output, err := svc.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
    if err != nil {
        return "", "", fmt.Errorf("failed to get ECR authorization token: %v", err)
    }

    authData := output.AuthorizationData[0]
    log.Print("Successfully retrieved ECR authorization token")
    return *authData.AuthorizationToken, *authData.ProxyEndpoint, nil
}

func loginToECR(authToken, registryURL string) error {
    log.Printf("Logging in to ECR at %s", registryURL)
    cmd := exec.Command("podman", "login", "--username", "AWS", "--password", authToken, registryURL)
    err := cmd.Run()
    if err != nil {
        return fmt.Errorf("failed to login to ECR: %v", err)
    }
    log.Print("Successfully logged in to ECR")
    return nil
}