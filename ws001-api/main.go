package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"os"
	"reflect"
	pingpong "ws001/pb"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
)

func main() {

	// log set up
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)

	const (
		WS002Addr = "ws002-pingpong:7002"
		WS003Addr = "ws003-httpsvc:7003"
	)

	r:= gin.Default()

	// health check for k8s service
	r.GET("/",  func(c *gin.Context) {
		c.String(200, "alive")
		return
	})

	// Example ping-pong via gRPC
	{
		var opts []grpc.DialOption
		opts = append(opts, grpc.WithInsecure())
		opts = append(opts, grpc.WithStreamInterceptor(
			grpc_opentracing.StreamClientInterceptor(
				grpc_opentracing.WithTracer(opentracing.GlobalTracer()))))
		opts = append(opts, grpc.WithUnaryInterceptor(
			grpc_opentracing.UnaryClientInterceptor(
				grpc_opentracing.WithTracer(opentracing.GlobalTracer()))))

		conn, err := grpc.Dial(WS002Addr, opts...)
		if err != nil {
			log.Panicf("did not connect: %v", err)
		}
		r.GET("/api/pingpong", func(c *gin.Context) {
			if err := doPingPong(conn); err != nil {
				c.String(500, err.Error())
				return
			}
			c.String(200, "pong !")
			return
		})
	}

	// Example invoke http service
	{
		zipkinHeaders := []string{
			"X-Request-Id",
			"X-B3-Traceid",
			"X-B3-Spanid",
			"X-B3-Parentspanid",
			"X-B3-Sampled",
			"X-B3-Flags",
			"X-Ot-Span-Context",
		}

		r.GET("/api/httpsvc", func(c *gin.Context) {

			logrus.Info("/api/httpsvc", "trace_id", c.Request.Header.Get("x-b3-traceid"))

			client := &http.Client{}
			req, err := http.NewRequest("GET", "http://"+WS003Addr, nil)
			if err != nil {
				c.String(500, err.Error())
				return
			}

			// getForwardHeaders
			for k := range c.Request.Header {
				isZipkinHeader, _ :=  inArray(k, zipkinHeaders)
				if isZipkinHeader{
					req.Header.Add(k, c.Request.Header.Get(k))
				}
			}

			logrus.Debug("ForwardHeaders ", req.Header)

			resp, err := client.Do(req)
			if err != nil {
				c.String(500, err.Error())
				return
			}

			if resp.StatusCode >= 500 {
				c.String(500, "httpsvc response code: "+resp.Status)
				return
			}
			c.String(200, "success !")
		})
	}

	r.Run(":7001")
}

func doPingPong(conn *grpc.ClientConn) error{
	client := pingpong.NewPingPongServiceClient(conn)

	resp, err := client.PingPongEndpoint(context.Background(), &pingpong.PingPong{Ping:1})
	if err != nil {
		return err
	}
	if resp.GetPong() != 1 {
		return fmt.Errorf("It is not pong !")
	}
	fmt.Println("received pong")
	return nil
}

func inArray(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}

	return
}