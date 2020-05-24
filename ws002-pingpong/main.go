package main

import (
	"context"
	"github.com/sirupsen/logrus"
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
	pingpong.RegisterPingPongServiceServer(grpcSvc, &SvcA{})
	reflection.Register(grpcSvc)

	if err:= grpcSvc.Serve(lis); err!= nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

type SvcA struct{}
func (*SvcA) PingPongEndpoint(ctx context.Context, req *pingpong.PingPong) ( resp *pingpong.PingPong, err error) {
	logrus.Info("received ping")
	return &pingpong.PingPong{
		Pong:                 1,
	}, nil
}