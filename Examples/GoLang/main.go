package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	url := "http://localhost:8080/cache/Hello"
	http.Post(url, "", bytes.NewBufferString("World"))
	response, err := http.Get(url)
	if err == nil {
		value, err := ioutil.ReadAll(response.Body)
		if err == nil {
			fmt.Println("Hello =", string(value))
		}
	}
}
