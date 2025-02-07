// Create a Golang server that listens 8081, returns a JSON like
// `{"experimental_flag_api_rest":{"key":"on"}}`
// and logs the input request like this
// `2025/02/07 14:50:52 Method: GET, Path: /v1/vardata, Body: {}, Headers: map[Accept-Encoding:[gzip] User-Agent:[KrakenD Version 2.9.1] X-Appversion:[84040] X-Deviceos:[0] X-Forwarded-For:[192.168.215.1] X-Forwarded-Host:[localhost:8080]], Query: map[device_id:[2a56cb88-51c5-43ac-b8df-57ca891d231f] flag_keys:[experimental_flag_api_rest,another]]`

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/v1/vardata", func(w http.ResponseWriter, r *http.Request) {
		// Log the request
		fmt.Printf("%s Method: %s, Path: %s, Body: %s, Headers: %v, Query: %v\n",
			time.Now().Format("2006/01/02 15:04:05"),
			r.Method,
			r.URL.Path,
			"{}", // r.Body is a ReadCloser, we can't read it twice
			r.Header,
			r.URL.Query(),
		)

		// Return the response
		resp := map[string]interface{}{
			"experimental_flag_api_rest": map[string]string{
				"key": "on",
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	fmt.Println("Server listening on :8081")
	http.ListenAndServe(":8081", nil)
}
