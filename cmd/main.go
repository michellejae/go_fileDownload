package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

var URL = "https://storage.googleapis.com/covid19-open-data/v2/epidemiology.csv"

func main() {

	res, err := http.Head(URL)
	if err != nil {
		log.Fatalf("head request error: %v", err)
	}

	contentLength := res.ContentLength
	//etag := res.Header.Get("Etag")

	size := contentLength / 4

	downloadFile(URL, size)
	//fmt.Println(etag, size)

}

func downloadFile(url string, size int64) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatalf("new request error: %v", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("response error: %v", err)
	}
	defer res.Body.Close()

	//limitedReader := &io.LimitedReader{R: res.Body, 1_000_000}
	body, err := ioutil.ReadAll(io.LimitReader(res.Body, size))
	if err != nil {
		log.Fatalf("body limit error: %v", err)
	}

	text := string(body)

	fmt.Println(text)

}
