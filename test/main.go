package main

import (
	"fmt"
	"net/http"
)
const baseURL = "http://34.68.16.210"

func main() {
	client := &http.Client{}
	req, err := http.NewRequest("GET", baseURL+"/api/httpsvc", nil)
	if err != nil {
		panic(err.Error())
	}

	go func(r *http.Request) {
		for i:=0;i<100;i++{
			if _, err := client.Do(r);err != nil {
				fmt.Println(err.Error())
			}
		}
	}(req)

	go func(r *http.Request) {
		for i:=0;i<100;i++{
			if _, err := client.Do(r);err != nil {
				fmt.Println(err.Error())
			}
		}
	}(req)

	select {

	}
}

