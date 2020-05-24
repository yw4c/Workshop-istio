package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"os"
	"reflect"
)

func main()  {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)


	r:= gin.Default()
	// health check for k8s service
	r.GET("/",  func(c *gin.Context) {

		zipkinHeaders := []string{
			"X-Request-Id",
			"X-B3-Traceid",
			"X-B3-Spanid",
			"X-B3-Parentspanid",
			"X-B3-Sampled",
			"X-B3-Flags",
			"X-Ot-Span-Context",
		}

		logrus.Info("received request")
		logrus.WithField("request_header", c.Request.Header).Debug()

		for k := range c.Request.Header {
			isZipkinHeader, _ :=  inArray(k, zipkinHeaders)
			if isZipkinHeader{
				c.Writer.Header().Add(k, c.Request.Header.Get(k))
			}
		}
		logrus.Debug("response header spanId", c.Writer.Header().Get("X-B3-Spanid"))

		c.String(200, "alive")
	})

	r.Run(":7003")
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
