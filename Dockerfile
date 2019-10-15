FROM golang:latest 

LABEL version="1.0"

RUN mkdir /go/src/app

ENV DB_HOST redis
ENV DB_PORT 6379
ENV DB_NAME 0
ENV APP_PORT 8080
ENV APP_HOST 0.0.0.0

COPY . /go/src/app

WORKDIR /go/src/app 

RUN go build

CMD ["./app"]
