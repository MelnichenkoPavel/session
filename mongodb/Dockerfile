FROM golang:alpine

WORKDIR /go/src/app
COPY ./src/app .

RUN set -xe; apk add --no-cache git

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["app"]