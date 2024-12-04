FROM golang:1.20 as builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o checkpoint_container

FROM amazon/aws-cli

RUN yum update -y && \
    yum install -y containerd buildah podman && \
    yum clean all

COPY --from=builder /app/checkpoint_container /usr/local/bin/checkpoint_container

EXPOSE 8080

ENTRYPOINT ["checkpoint_container"]