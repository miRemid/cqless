package main

import (
	"fmt"
	"io"
	"net/http"
)

func copyHeaders(destination http.Header, source *http.Header) {
	for k, v := range *source {
		vClone := make([]string, len(v))
		copy(vClone, v)
		destination[k] = vClone
	}
}

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("hello!"))
	})

	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		data, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		copyHeaders(w.Header(), &r.Header)
		w.WriteHeader(200)
		_, err = w.Write(data)
		if err != nil {
			fmt.Println(err)
		}
	})

	s := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	s.ListenAndServe()

}
