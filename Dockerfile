FROM golang:1.14 as builder

WORKDIR /go-modules

COPY . ./

RUN COMMAND
