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

In deployment.yaml update the following line.

*image: `<your-docker-repo>`/checkpoint-container:v1*

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

**EKS Specific**

**Add the EKS chart repo to Helm**

```
helm repo add eks https://aws.github.io/eks-charts
```

**Install the AWS Load Balancer Controller**

```
helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
  -n kube-system \
  --set clusterName=<your-cluster-name> \
  --set serviceAccount.create=false \
  --set serviceAccount.name=aws-load-balancer-controller
```

*Replace <your-cluster-name> with your EKS cluster name.*

*Note: Ensure that you have the necessary IAM permissions set up for the AWS Load Balancer Controller. You can find the detailed IAM policy in the AWS documentation.*

```
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

```
kubectl apply -f ingress.yaml
```

**Get the ALB DNS name**

```   
kubectl get ingress checkpoint-ingress
```

**Test the API**

```
curl -X POST http://<ALB-DNS-NAME>/checkpoint \
     -H "Content-Type: application/json" \
     -d '{"podId": "your-pod-id", "ecrRepo": "your-ecr-repo", "awsRegion": "your-aws-region"}'
```
