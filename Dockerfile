## Build the user interface
FROM quay.io/continuouspipe/nodejs7:stable as userinterface

# Install prerequisites build tools
RUN apt-get update \
  && apt-get install -y ruby ruby-dev build-essential git \
  && gem install --no-rdoc --no-ri sass -v 3.4.22 \
  && gem install --no-rdoc --no-ri compass \
  && npm install -g grunt-cli bower

# Build the application
RUN mkdir /app
WORKDIR /app/ui

# Install node dependencies
ADD ./ui/package.json /app/ui/package.json
RUN npm install

# Install bower dependencies
ADD ./.bowerrc /app/.bowerrc
ADD ./bower.json /app/bower.json
RUN cd /app && bower install --config.interactive=false --allow-root

# Build the code
COPY ./ui/ /app/ui
RUN grunt build

## Build the API
FROM golang:1.8 as api

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
    go install

## Run the API & UI
FROM golang:1.8

ARG KUBE_STATUS_LISTEN_ADDRESS=

WORKDIR /app

COPY --from=userinterface /app/dist /app/var/static
COPY --from=api /go/bin/kube-status /usr/bin/kube-status

ADD ./ui/docker/run.sh /app/prepare-ui.sh
ADD ./docker/run.sh /app/run.sh

CMD ["/app/run.sh"]
