package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type deets struct {
	i      int
	offset int64
	limit  int64
}

var URL = "https://storage.googleapis.com/covid19-open-data/v2/epidemiology.csv"

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := initialPing(ctx)
	if err != nil {
		fmt.Printf("initial ping error %v", err)
	}

}

func initialPing(ctx context.Context) error {
	ch := make(chan deets)
	// just send header request to URL
	req, err := http.NewRequestWithContext(ctx, "HEAD", URL, nil)
	if err != nil {
		log.Fatalf("new request error: %v", err)
	}
	res, err := http.DefaultClient.Do(req)
	//res, err := http.Head(URL)
	if err != nil {
		log.Fatalf("head request error: %v", err)
	}
	// get content length and eTag

	contentLength := res.ContentLength

	//contentLength = int64(105)
	//	etag := res.Header.Get("Etag")
	// length of every file chunk we will want to download
	sectionLength := contentLength / 3

	if res.StatusCode != http.StatusOK {
		log.Fatal(res.Status)
	}
	// need an index to send through
	i := 0
	// loop through the content length; but jump by every sectoin length
	// ie if contentlength is 105, and our section length is 26 (100/4) we would += 26 each time
	for offset := int64(0); offset < contentLength; offset += sectionLength {
		fmt.Println("here")
		i++
		offset := offset
		// limit will be our 'high' each time cause it's our min
		limit := offset + sectionLength
		// because we do not have an even sectionLengt, eventually our limit will go over contentlength
		// so if limits are 0, 26, 52, 78, 104 .. we still have no hit 105 our conentlegnth so we would loop
		// through once more ...but now we would sit limit to 120 and that's too high so limit is not just 105
		if limit > contentLength {
			limit = contentLength
		}
		go func(ctx context.Context, url string, offset, limit int64, i int) {
			downloadFile(ctx, URL, offset, limit, i)
			ch <- test{i, limit}

		}(ctx, URL, offset, limit, i)

	}
	// for j := 0; j < i; j++ {
	// 	result := <-ch
	// 	fmt.Println("RESULT", result)
	// }
	for off := int64(0); off < contentLength; off += sectionLength {
		result := <-ch
		fmt.Println("RESULT", result)
	}

	return nil

}

func downloadFile(ctx context.Context, url string, offset int64, limit int64, i int) {

	req, err := http.NewRequestWithContext(ctx, "GET", URL, nil)
	if err != nil {
		log.Fatalf("new request error: %v", err)
	}

	// need to set our range header so we are only downloading a portion of the file
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", offset, limit))
	// fire off request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("response error: %v", err)
	}
	// have to defer body.close so not memeory leak
	defer res.Body.Close()

	if res.StatusCode != http.StatusPartialContent {
		log.Fatalf(res.Status)
	}

	//limitedReader := &io.LimitedReader{R: res.Body, 1_000_000}
	//	body, err := ioutil.ReadAll(io.LimitReader(res.Body, size))
	// if err != nil {
	// 	log.Fatalf("body limit error: %v", err)
	// }

	body, err := ioutil.ReadAll(res.Body)

	fmt.Println(body)

	createEmptyFile("./test.txt", offset, body)

	// fmt.Println(text)

}

func createEmptyFile(path string, offset int64, body []byte) error {
	fmt.Println("hello poppit")
	file, err := os.OpenFile(path, os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer file.Close()
	file.Seek(offset, 0)

	os.WriteFile(path, body, 0666)
	return nil
}

func createEmptyFile(path string, size int64) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	file.Seek(size-1, os.SEEK_SET)
	file.Write([]byte{0})
	return nil
}
