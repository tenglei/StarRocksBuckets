package tools

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

func Post(method, u string, body io.Reader) []byte {
	request, err := http.NewRequest(method, u, body)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	request.Header.Set("Content-Type", "application/json;charset=utf-8")
	client := &http.Client{
		Timeout:   time.Second * 30,
		Transport: &http.Transport{},
	}
	respone, err := client.Do(request)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	defer respone.Body.Close()
	b, err := ioutil.ReadAll(respone.Body)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return b
}
