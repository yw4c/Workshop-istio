FROM golang:alpine
WORKDIR /app
ENV GO111MODULE=on

COPY . .

RUN apk update && apk upgrade && \
    apk add --no-cache bash curl
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0  go build -o app

CMD ["./app"]

EXPOSE 7003