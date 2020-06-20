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

	//gw "ws001/pb"
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
		pingpong, err := ws002Cli.PingPongEndpoint(c.Request.Context(), &pingpong.PingPong{Ping:1})
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
	r.GET("private/api/auth-info", func(c *gin.Context) {
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


	// gRPC Gateway
	//annotators := []annotator{injectHeadersIntoMetadata}
	//mux := runtime.NewServeMux(
	//	runtime.WithMetadata(chainGrpcAnnotators(annotators...)),
	//)
	//opts := []grpc.DialOption{grpc.WithInsecure()}
	//err := gw.RegisterPingPongServiceHandlerFromEndpoint(ctx, mux, WS002Addr, opts)
	//if err != nil {
	//	logrus.Fatal(err.Error())
	//}

	r.Run(httpPort)


}



// 註釋者: http 轉 grpc metadata
type annotator func(context.Context, *http.Request) metadata.MD

// 實現註釋者：轉換追蹤表頭
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
				logrus.Debug("merging otHeader:" , h)
				v := req.Header.Get(k)
				pairs = append(pairs, h, v)
			}
		}
	}

	return metadata.Pairs(pairs...)
}

// 將所有註釋者 metadata 組合
func chainGrpcAnnotators(annotators ...annotator) annotator {
	return func(c context.Context, r *http.Request) metadata.MD {
		var mds []metadata.MD
		for _, a := range annotators {
			mds = append(mds, a(c, r))
		}
		return metadata.Join(mds...)
	}
}