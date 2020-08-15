package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	url := "http://localhost:8080/cache/Hello"
	client := &http.Client{}
	request, _ := http.NewRequest(http.MethodPut, url, bytes.NewBufferString("World"))
	client.Do(request)

	response, err := http.Get(url)
	if err == nil {
		value, err := ioutil.ReadAll(response.Body)
		if err == nil {
			fmt.Println("Hello =", string(value))
		}
	}
}
