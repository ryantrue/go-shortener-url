package main

import (
	"fmt"
	"net/http"

	internal "github.com/RyanTrue/go-shortener-url/internal/app"
)

func webhook(m internal.Model) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("webhook")

		if r.Method == http.MethodGet {
			fmt.Println("MethodGet")
			internal.GetURL(m, w, r)
			return
		} else if r.Method == http.MethodPost {
			fmt.Println("MethodPost")
			internal.ReceiveURL(m, w, r)
			return
		} else {
			fmt.Println("StatusBadRequest")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

}

func main() {
	if err := run(); err != nil {
		panic(err)
	}

}

func run() error {
	mux := http.NewServeMux()
	m := make(internal.Model)
	mux.HandleFunc(`/`, webhook(m))

	return http.ListenAndServe(`:8080`, mux)
}
