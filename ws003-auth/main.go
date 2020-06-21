package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {

	// logger set up
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)

	// Server Addrs
	r := gin.Default()

	// Log tracing info
	r.Use(func(context *gin.Context) {
		logrus.WithField("x-request-id", context.Request.Header.Get("x-request-id")).
			WithField("x-b3-traceid", context.Request.Header.Get("x-b3-traceid")).
			WithField("URI", context.Request.Method + context.Request.RequestURI).
			Info("Tracing Info")
	})

	r.GET("/validate", func(c *gin.Context) {
		logrus.Info("auth", c.Request.Header.Get("Authorization"))
		if c.Request.Header.Get("Authorization") != "1234" {
			c.Status(403)
			return
		}
		c.Writer.Header().Set("x-secret","{user-id:1}")
		c.Status(200)
		return
	})

	r.Run(":7003")

}


