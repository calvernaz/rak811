FROM golang:1.12 as builder

WORKDIR /go-modules

COPY . ./

RUN COMMAND
