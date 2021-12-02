FROM ubuntu:20.04

RUN export DEBIAN_FRONTEND=noninteractive
RUN export GOBIN=$GOPATH/bin
RUN ln -fs /usr/share/zoneinfo/Europe/Paris /etc/localtime
RUN apt-get update && apt-get install -y golang git ca-certificates
RUN mkdir /build
COPY go.sum  /build/
COPY go.mod  /build/
WORKDIR /build/
RUN go mod download
COPY . /build/
WORKDIR /build/killmailsGetter
RUN go build 
WORKDIR /build/killmailsServer
RUN go build
WORKDIR /build/
CMD /bin/bash
