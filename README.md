# Container Checkpoint API

Container checkpoint in Kubernetes using `ctr`

#### Prerequisites
* Kubernetes cluster
* Install ctr commandline tool. (if you are able to run ctr commands on the kubelet/worker node, if not install/adjust AMI to contain the ctr. https://github.com/containerd/containerd/tree/main/cmd/ctr)
* kubectl configured to communicate with your cluster
* Docker installed locally
* Access to a container registry (e.g., ECR)
* Helm (for installing Nginx/ALB Ingress Controller)

#### Initialize the go module
```sh
go mod init checkpoint_container
```
##### Modify the go.mod file
```go
module checkpoint_container

go 1.23

require (
	github.com/aws/aws-sdk-go v1.44.298
	github.com/containerd/containerd v1.7.2
)

require (
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc2.0.20221005185240-3a7f492d3f1b // indirect
	github.com/pkg/errors v0.9.1 // indirect
	google.golang.org/genproto v0.0.0-20230306155012-7f2fa6fef1f4 // indirect
	google.golang.org/grpc v1.53.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
)
```
*Run:*

```sh
go mod tidy
```

#### Build and Publish Docker Image**

```sh
docker build -t <your-docker-repo>/checkpoint-container:v1 .
docker push <your-docker-repo>/checkpoint-container:v1
```

*Replace ```<your-docker-repo>``` with your actual Docker repository.*

#### Apply the RBAC resources

```sh
kubectl apply -f rbac.yaml
```

#### Deployment

```sh
kubectl apply -f deployment.yaml
```
In deployment.yaml update the following

*image: `<your-docker-repo>`/checkpoint-container:v1*

#### Kubernetes Service

```sh
kubectl apply -f service.yaml
```

#### Install Ngnix Ingress Contoller

```sh
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
helm install ingress-nginx ingress-nginx/ingress-nginx
kubectl apply -f ingress.yaml
```

#### Test the API

```sh
kubectl get services ingress-ngnix-contoller -n ingress-ngnix
```

```sh
curl -X POST http://<EXTERNAL-IP>/checkpoint \
 -H "Content-Type: application/json" \
 -d '{"podId": "your-pod-id", "ecrRepo": "your-ecr-repo", "awsRegion": "your-aws-region"}'
```

*Replace ```<EXTERNAL-IP>``` with the actual external IP.*


### EKS Specific

**1. Add the EKS chart repo to Helm:**

```sh
helm repo add eks https://aws.github.io/eks-charts
```

**2. Install the AWS Load Balancer Controller:**

```sh
helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
  -n kube-system \
  --set clusterName=<your-cluster-name> \
  --set serviceAccount.create=false \
  --set serviceAccount.name=aws-load-balancer-controller
```

*Note: Replace `<your-cluster-name>` with your EKS cluster name.*

*Note: Ensure that you have the necessary IAM permissions set up for the AWS Load Balancer Controller. You can find the detailed IAM policy in the AWS documentation.*

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: checkpoint-ingress
  annotations:
    kubernetes.io/ingress.class: alb
    alb.ingress.kubernetes.io/scheme: internet-facing
    alb.ingress.kubernetes.io/target-type: ip
spec:
  rules:
  - http:
      paths:
      - path: /checkpoint
        pathType: Prefix
        backend:
          service:
            name: checkpoint-service
            port: 
              number: 80
```

```sh
kubectl apply -f ingress.yaml
```

**3. Get the ALB DNS name:**

```sh
kubectl get ingress checkpoint-ingress
```

**4. Test the API**
```sh
curl -X POST http://<ALB-DNS-NAME>/checkpoint \
     -H "Content-Type: application/json" \
     -d '{"podId": "your-pod-id", "ecrRepo": "your-ecr-repo", "awsRegion": "your-aws-region"}'
```
