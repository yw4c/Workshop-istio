package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"log"
	"net"
	"net/http"
	"os"
	pingpong "ws002/pb"
)
import "google.golang.org/grpc"
import "google.golang.org/grpc/reflection"

func main()  {

	const (
		grpcPort = ":7002"
		httpPort = ":8002"
	)

	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetOutput(os.Stdout)

	// ######### gRPC ############
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcSvc := grpc.NewServer()
	pingpong.RegisterPingPongServiceServer(grpcSvc, &PingPongSvc{})
	reflection.Register(grpcSvc)

	go func() {
		if err:= grpcSvc.Serve(lis); err!= nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()


	// ########## HTTP ##############
	r := gin.Default()

	// 需要受 auth 驗證的 endpoint , 請參考 README - envoy filter 章節
	r.Any("private/pingpong/auth-info", func(c *gin.Context) {
		logrus.Infof("%+v", c.Request.Header)
		xSecret := c.Request.Header.Get("x-secret") // Was Set in ws003
		if xSecret == "" {
			data := gin.H{
				"msg": "we didn't get x-secret",
			}
			c.JSONP(http.StatusUnauthorized, data)
			return
		}

		data := gin.H{
			"msg": "we got auth info "+ xSecret,
		}
		c.JSON(http.StatusOK, data)
		return
	})

	r.Run(httpPort)
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