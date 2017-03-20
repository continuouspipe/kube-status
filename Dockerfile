FROM golang:1.8

ARG KUBE_STATUS_LISTEN_ADDRESS=

# Add the application
ADD . /go/src/github.com/continuouspipe/kube-status
WORKDIR /go/src/github.com/continuouspipe/kube-status

### For now vendor/ is included in the repository since we have very few dependencies, if the number of dependencies
### grows significantly then vendor/ can be removed and the dependencies installed using glide
## Add glide to the image
# RUN wget https://github.com/Masterminds/glide/releases/download/v0.12.3/glide-v0.12.3-linux-amd64.tar.gz && \
#     tar -xzvf glide-v0.12.3-linux-amd64.tar.gz -C /usr/local/bin --strip-components=1 linux-amd64/glide && \
#     rm glide-v0.12.3-linux-amd64.tar.gz
## Install dependencies
# RUN glide install --strip-vendor

# Run build
RUN go install

# Run the kube-status when the container starts.
ENTRYPOINT ["/go/bin/kube-status", "-logtostderr"]