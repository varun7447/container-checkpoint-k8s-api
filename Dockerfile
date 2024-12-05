# Build stage
FROM golang:1.23 as builder

WORKDIR /app
COPY . .
RUN GOPROXY=direct CGO_ENABLED=0 GOOS=linux go build -o checkpoint_container

# Final stage
FROM amazonlinux:2

# Install necessary tools
RUN yum update -y && \
    amazon-linux-extras install -y docker && \
    yum install -y awscli containerd skopeo && \
    yum clean all

# Copy the built Go binary
COPY --from=builder /app/checkpoint_container /usr/local/bin/checkpoint_container

EXPOSE 8080

ENTRYPOINT ["checkpoint_container"]
