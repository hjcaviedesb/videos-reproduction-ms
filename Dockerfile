FROM golang:latest

WORKDIR $GOPATH/src/video-reproduction-ms
COPY . .
RUN apt-get update
RUN apt-get install vim -y

RUN go get -d -v ./...
RUN go build

CMD ["./video-reproduction-ms"] 