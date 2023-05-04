package main

import (
	"net/http"
	"strings"
)

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("hello!"))
	})

	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		param := r.URL.Query()["param"]
		returnMessage := strings.Join(param, " ")
		w.WriteHeader(200)
		w.Write([]byte(returnMessage))
	})

	mux.HandleFunc("/echo/echo", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		param := r.URL.Query()["param"]
		returnMessage := strings.Join(param, " ")
		w.WriteHeader(200)
		w.Write([]byte(returnMessage))
	})

	s := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	s.ListenAndServe()

}
