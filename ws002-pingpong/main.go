package main

import (
	"context"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"log"
	"net"
	"os"
	pingpong "ws002/pb"
)
import "google.golang.org/grpc"
import "google.golang.org/grpc/reflection"

func main()  {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetOutput(os.Stdout)

	lis, err := net.Listen("tcp", ":7002")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcSvc := grpc.NewServer()
	pingpong.RegisterPingPongServiceServer(grpcSvc, &PingPongSvc{})
	reflection.Register(grpcSvc)

	if err:= grpcSvc.Serve(lis); err!= nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

type PingPongSvc struct{}
func (*PingPongSvc) PingPongEndpoint(ctx context.Context, req *pingpong.PingPong) ( resp *pingpong.PingPong, err error) {

	// 取得 trace 表頭
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logrus.Error("get metadata error")
	}
	logrus.WithField("requestID", md.Get("x-request-id")).
		WithField("traceID", md.Get("x-b3-traceid")).
		WithField("spanID", md.Get("x-b3-spanid")).Info("Tracing Info")

	logrus.Info("received ping")
	return &pingpong.PingPong{
		Pong:                 1,
	}, nil
}