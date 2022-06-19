package main

import (
	"fmt"
	"github.com/bloominlabs/baseplate-go/http"
)

func main() {
	c := http.NewClient()

	resp, err := c.Get("https://tip.service.consul")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	fmt.Println(resp.Body)
}
