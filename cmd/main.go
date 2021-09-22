package main

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/michellejae/filedownload/downloadfile"
)

type deets struct {
	offset int64
	body   []byte
}

var URL = "https://storage.googleapis.com/covid19-open-data/v2/epidemiology.csv"

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := initialPing(ctx)
	if err != nil {
		fmt.Printf("initial ping error %v", err)
	}

}

func initialPing(ctx context.Context) error {

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

	//contentLength = int64(1105)
	etag := res.Header.Get("Etag")
	fmt.Printf("etag %v \n", etag)
	// length of every file chunk we will want to download
	sectionLength := contentLength / 3

	if res.StatusCode != http.StatusOK {
		log.Fatal(res.Status)
	}

	sum, err := createEmptyFile(ctx, "./test2.txt", contentLength, sectionLength)
	if err != nil {
		fmt.Println("wherejfdsjl", err)
	}

	fmt.Printf("%x", sum)
	return nil

}

func createEmptyFile(ctx context.Context, path string, contentLength, sectionLength int64) ([]byte, error) {
	ch := make(chan deets)
	file, err := os.Create(path)
	if err != nil {
		fmt.Printf("error opening file %v", err)
	}
	defer file.Close()

	// loop through the content length; but jump by every sectoin length
	// ie if contentlength is 105, and our section length is 26 (105/4) we would += 26 each time
	for offset := int64(0); offset < contentLength; offset += sectionLength {
		//offset := offset
		// limit will be our 'high' each time
		limit := offset + sectionLength
		// because we do not have an even sectionLengt, eventually our limit will go over contentlength
		// so if limits are 0, 26, 52, 78, 104 .. we still have no hit 105 our conentlegnth so we would loop
		// through once more ...but now we would sit limit to 120 and that's too high so limit is not just 105
		if limit > contentLength {
			limit = contentLength
		}
		// make call to download the file bits at a time in parrellel
		go func(ctx context.Context, url string, offset, limit int64) {
			body := downloadfile.Download(ctx, URL, offset, limit)
			ch <- deets{offset, body}

		}(ctx, URL, offset, limit)

	}
	// loop back through the same stuff so i can receive channel
	for off := int64(0); off < contentLength; off += sectionLength {
		result := <-ch
		// find correct spot in file
		file.Seek(result.offset, 0)
		// obviously write to the file
		file.Write(result.body)
	}
	h := md5.New()
	if _, err := io.Copy(h, file); err != nil {
		log.Fatal(err)
	}
	sum := h.Sum(nil)

	return sum, nil
}
