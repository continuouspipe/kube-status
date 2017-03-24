FROM golang:1.8

ARG KUBE_STATUS_LISTEN_ADDRESS=

# Add the application
ADD . /go/src/github.com/continuouspipe/kube-status
WORKDIR /go/src/github.com/continuouspipe/kube-status

# Add glide to the image
RUN wget https://github.com/Masterminds/glide/releases/download/v0.12.3/glide-v0.12.3-linux-amd64.tar.gz && \
    tar -xzvf glide-v0.12.3-linux-amd64.tar.gz -C /usr/local/bin --strip-components=1 linux-amd64/glide && \
    rm glide-v0.12.3-linux-amd64.tar.gz && \

# Install dependencies
    glide install --strip-vendor && \

# Run build
    go install && \

# Extract the build binary
    mv /go/bin/kube-status /usr/bin/kube-status && \

# Clean up the image
    rm -rf /go

# Run the kube-status when the container starts.
ENTRYPOINT ["/usr/bin/kube-status", "-logtostderr", "-v", "5"]
