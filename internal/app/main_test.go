package app

import (
	"github.com/RyanTrue/go-shortener-url/config"
	store "github.com/RyanTrue/go-shortener-url/storage"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string, body io.Reader) *http.Response {

	req, err := http.NewRequest(method, ts.URL+path, body)
	req.Close = true
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("User-Agent", "PostmanRuntime/7.32.3")
	require.NoError(t, err)

	ts.Client()

	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	return resp
}

func runTestServer(storage *store.LinkStorage, conf config.Config, db *store.Database) chi.Router {
	router := chi.NewRouter()

	router.Get("/{id}", func(rw http.ResponseWriter, r *http.Request) {
		GetURL(storage, rw, r, conf, db)
	})
	router.Post("/", func(rw http.ResponseWriter, r *http.Request) {
		ReceiveURL(storage, rw, r, conf, db)
	})
	router.Route("/api", func(r chi.Router) {
		r.Post("/shorten", func(rw http.ResponseWriter, r *http.Request) {
			ReceiveURLAPI(storage, rw, r, conf, db)
		})
	})

	return router
}
