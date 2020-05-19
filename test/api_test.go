package test

import (
	"net/http"
	"testing"
	"time"
)
const baseURL = "http://localhost:7001"

func TestPingPong(t *testing.T) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", baseURL+"/api/pingpong", nil)
	if err != nil {
		panic(err.Error())
	}

	for {
		client.Do(req)
		time.Sleep(5 * time.Second)
	}
}