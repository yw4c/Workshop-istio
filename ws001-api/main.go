package main

import (
	"context"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net/http"
	"os"
	"strings"
	gw "ws001/pb"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
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


	annotators := []annotator{injectHeadersIntoMetadata}
	mux := runtime.NewServeMux(
		runtime.WithMetadata(chainGrpcAnnotators(annotators...)),
	)
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := gw.RegisterPingPongServiceHandlerFromEndpoint(ctx, mux, WS002Addr, opts)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	http.ListenAndServe(httpPort, mux)


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