# container-checkpoint-k8s-api

Container checkpoint in Kubernetes using `crt`

**Prerequisites**
* Kubernetes cluster
* Install ctr commandline tool. if you are able to run ctr commands on the kubelet/worker node, if not install/adjust AMI to contain the ctr. https://github.com/containerd/containerd/tree/main/cmd/ctr
* kubectl configured to communicate with your cluster
* Docker installed locally
* Access to a container registry (e.g., Docker Hub, ECR)
* Helm (for installing Nginx Ingress Controller)

**Build and Publish Docker Image**

```
docker build -t <your-docker-repo>/checkpoint-container:v1 .
docker push <your-docker-repo>/checkpoint-container:v1
```

*Replace ```<your-docker-repo>``` with your actual Docker repository.*

**Apply the RBAC resources**

```
kubectl apply -f rbac.yaml
```

**Deployment**

```
kubectl apply -f deployment.yaml
```
In deployment.yaml update the following

*image: <your-docker-repo>/checkpoint-container:v1*

**Kubernetes Service**

```
kubectl apply -f service.yaml
```

**Install Ngnix Ingress Contoller**

```
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
helm install ingress-nginx ingress-nginx/ingress-nginx
```

**Ingress**

```
kubectl apply -f ingress.yaml
```

**Test the API**

```
kubectl get services ingress-ngnix-contoller -n ingress-ngnix
```

```
curl -X POST http://<EXTERNAL-IP>/checkpoint \
 -H "Content-Type: application/json" \
 -d '{"podId": "your-pod-id", "ecrRepo": "your-ecr-repo", "awsRegion": "your-aws-region"}'
```

*Replace ```<EXTERNAL-IP>``` with the actual external IP.*
