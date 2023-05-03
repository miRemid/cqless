package main

import "net/http"

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("hello!"))
	})

	s := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	s.ListenAndServe()

}
