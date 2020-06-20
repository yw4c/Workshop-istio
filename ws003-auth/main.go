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


