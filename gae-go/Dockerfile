FROM google/cloud-sdk:alpine
LABEL maintainer "supertaihei02"

ARG GOLANG_VERSION=1.8.5

# install go
RUN curl -Lso go${GOLANG_VERSION}.linux-amd64.tar.gz https://storage.googleapis.com/golang/go${GOLANG_VERSION}.linux-amd64.tar.gz \
	&& tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz \
	&& rm go${GOLANG_VERSION}.linux-amd64.tar.gz

# install GAE for Go SDK
RUN gcloud components install app-engine-go

# install git
RUN apk add --update --no-cache \
		sudo \
		git

# for saving gcloud config
RUN gcloud config set core/disable_usage_reporting true && \
    gcloud config set component_manager/disable_update_check true && \
    gcloud config set metrics/environment github_docker_image

ENV GOPATH /go
ENV PATH /go/bin:/usr/local/go/bin:$PATH
