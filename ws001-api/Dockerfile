FROM golang:alpine
WORKDIR /app
ENV GO111MODULE=on

COPY . .

RUN apk update && apk upgrade && \
    apk add --no-cache bash curl

# grpcurl invoking test
RUN go get github.com/fullstorydev/grpcurl/...
RUN go install github.com/fullstorydev/grpcurl/cmd/grpcurl

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0  go build -o app

CMD ["./app"]

