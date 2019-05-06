package web

import (
	"bytes"
	"log"
	"net/http"
	"time"
)

func Hook(url string, method string, body *string) (int, error) {
	var httpClient = &http.Client{Timeout: 20 * time.Second}

	log.Printf("\n\n\n\n%s\n\n\n\n", *body)
	req, err := http.NewRequest(method, url, bytes.NewBufferString(*body))
	if err != nil {
		return -1, err
	}
	req.Header.Set("Content-Type", "application/json")

	r, err := httpClient.Do(req)

	if err != nil {
		return -1, err
	}

	defer r.Body.Close()

	return r.StatusCode, nil
}
