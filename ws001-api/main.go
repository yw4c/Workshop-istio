package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net/http"
	"os"
	"strings"
	pingpong "ws001/pb"

)

func main() {

	// logger set up
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)

	// Server Addrs
	const (
		httpPort = ":7001"
		WS002Addr = "ws002-pingpong:7002"
	)

	// init context
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// prepare gRPC connection
	ws002Conn, err := grpc.Dial(WS002Addr, grpc.WithInsecure())
	if err != nil {
		panic(err.Error())
	}
	ws002Cli := pingpong.NewPingPongServiceClient(ws002Conn)


	r := gin.Default()

	// 透過 gRPC 請求 pingpong
	r.GET("/public/api/pingpong", func(c *gin.Context) {
		ctx := c.Request.Context()

		// 注入來自 http 的追蹤表頭到 metadata
		md := injectHeadersIntoMetadata(ctx, c.Request)
		ctx = metadata.NewOutgoingContext(ctx, md)


		// 請求 grpc to ws002
		pingpong, err := ws002Cli.PingPongEndpoint(ctx, &pingpong.PingPong{Ping:1})
		if err != nil {
			c.AbortWithError(404, err)
			return
		}
		data := gin.H{
			"msg": pingpong,
		}
		c.JSONP(http.StatusOK, data)
		return
	})


	// 需要受 auth 驗證的 endpoint , 請參考 README - envoy filter 章節
	r.Any("private/api/auth-info", func(c *gin.Context) {
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

// injectHeadersIntoMetadata：轉換追蹤表頭
func injectHeadersIntoMetadata(ctx context.Context, req *http.Request) metadata.MD {
	//https://aspenmesh.io/2018/04/tracing-grpc-with-istio/
	var (
		otHeaders = []string{
			"x-request-id",
			"x-b3-traceid",
			"x-b3-spanid",
			"x-b3-parentspanid",
			"x-b3-sampled",
			"x-b3-flags",
			"x-ot-span-context"}
	)
	var pairs []string

	for k := range req.Header {
		for _, h := range otHeaders {
			if strings.ToLower(k) == h {
				//logrus.Debug("merging otHeader:" , h)
				v := req.Header.Get(k)
				pairs = append(pairs, h, v)
			}
		}
	}

	md := metadata.Pairs(pairs...)

	logrus.WithField("requestID", md.Get("x-request-id")).
		WithField("traceID", md.Get("x-b3-traceid")).
		WithField("spanID", md.Get("x-b3-spanid")).Info("Tracing Info")

	return md
}
