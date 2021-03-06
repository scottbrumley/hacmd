FROM golang:rc-alpine3.12

RUN apk add git
RUN cd / && git clone https://github.com/scottbrumley/hacmd.git

WORKDIR /hacmd
ENV GOPATH=/hacmd/
RUN go get "github.com/cskr/pubsub"
RUN go get "github.com/eclipse/paho.mqtt.golang"
RUN go build *.go

CMD /hacmd/hacmd /hacmd/config.json