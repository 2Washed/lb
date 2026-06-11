package main

import (
	"io"
	"log"
	"net/http"
)

func main() {
	var handler = func(w http.ResponseWriter, req *http.Request) {
		//Clone request
		outReq := req.Clone(req.Context())
		outReq.URL.Scheme = "http"
		outReq.URL.Host = "127.0.0.1:9000"
		outReq.RequestURI = ""

		//Forward the request
		res, forwardErr := http.DefaultClient.Do(outReq)
		if forwardErr != nil {
			log.Printf("[ERROR] Forwarding request failed, Error: %v\n", forwardErr)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(nil)
			return
		}
		defer res.Body.Close()

		//Write headers
		for headerName, headerValues := range res.Header {
			for _, v := range headerValues {
				w.Header().Add(headerName, v)
			}
		}
		w.WriteHeader(res.StatusCode)

		_, copyErr := io.Copy(w, res.Body)
		if copyErr != nil {
			log.Printf("[ERROR] Copying request to client failed, Error: %v\n", copyErr)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(nil)
			return
		}
	}

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
