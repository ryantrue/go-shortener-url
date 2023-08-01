package main

import (
	"fmt"
	"net/http"

	"github.com/RyanTrue/go-shortener-url/config"
	internal "github.com/RyanTrue/go-shortener-url/internal/app"
	"github.com/go-chi/chi"
)

func main() {
	config.ParseFlags()

	fmt.Println("Running server on", config.FlagRunAddr)

	http.ListenAndServe(config.FlagRunAddr, Run())
}

func Run() chi.Router {
	m := make(internal.Model)

	r := chi.NewRouter()
	r.Get("/{id}", func(rw http.ResponseWriter, r *http.Request) {
		internal.GetURL(m, rw, r)
	})
	r.Post("/", func(rw http.ResponseWriter, r *http.Request) {
		internal.ReceiveURL(m, rw, r, config.FlagBaseAddr)
	})

	return r
}
