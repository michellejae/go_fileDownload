package downloadfile

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var URL = "https://storage.googleapis.com/covid19-open-data/v2/epidemiology.csv"

func Download(ctx context.Context, url string, offset int64, limit int64) []byte {

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
	// reads res.body and converts in bytes
	body, err := ioutil.ReadAll(res.Body)

	return body

}
