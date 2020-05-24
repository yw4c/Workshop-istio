FROM golang:alpine
WORKDIR /app
ENV GO111MODULE=on

COPY . .
RUN ls

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh tar gzip curl
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0  go build -o app

CMD ["./app"]

