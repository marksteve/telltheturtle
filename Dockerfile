FROM golang:1.3
RUN go get github.com/tools/godep
WORKDIR /go/src/github.com/marksteve/telltheturtle

