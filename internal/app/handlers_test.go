package app

import (
	"github.com/RyanTrue/go-shortener-url/config"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	store "github.com/RyanTrue/go-shortener-url/storage"
	"github.com/RyanTrue/go-shortener-url/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetURL(t *testing.T) {
	tests := []struct {
		name       string
		request    string
		store      store.LinkStorage
		statusCode int
	}{
		{
			name:    "positive test #1",
			request: "/YjhkNDY",
			store: store.LinkStorage{
				Store: []store.Link{
					{
						ID:          uuid.New(),
						ShortURL:    "YjhkNDY",
						OriginalURL: "https://practicum.yandex.ru/",
					},
				},
			},
			statusCode: http.StatusTemporaryRedirect,
		},
		{
			name:    "positive test #2",
			request: "/" + util.Shorten("Y2NlMzI"),
			store: store.LinkStorage{
				Store: []store.Link{
					{
						ID:          uuid.New(),
						ShortURL:    util.Shorten("Y2NlMzI"),
						OriginalURL: "Y2NlMzI",
					},
				},
			},
			statusCode: http.StatusTemporaryRedirect,
		},
		{
			name:    "not found",
			request: "/" + util.Shorten("asdasda"),
			store: store.LinkStorage{
				Store: []store.Link{},
			},
			statusCode: http.StatusNotFound,
		},
	}
	conf := config.Config{
		FlagSaveToFile: false,
		FlagSaveToDB:   false,
		FlagBaseAddr:   "http://localhost:8000/",
	}

	for _, v := range tests {
		ts := httptest.NewServer(runTestServer(&v.store, conf, nil))
		defer ts.Close()

		resp := testRequest(t, ts, "GET", v.request, nil)
		defer resp.Body.Close()

		assert.Equal(t, v.statusCode, resp.StatusCode)

		if v.statusCode != http.StatusNotFound {
			assert.Equal(t, v.store.Store[0].OriginalURL, resp.Header.Get("Location"))
		}
	}
}

func TestReceiveURL(t *testing.T) {
	tests := []struct {
		name         string
		request      string
		store        store.LinkStorage
		statusCode   int
		body         []byte
		expectedBody string
	}{
		{
			name:    "positive test #1",
			request: "/",
			store: store.LinkStorage{
				Store: []store.Link{},
			},
			statusCode:   http.StatusCreated,
			body:         []byte("https://practicum.yandex.ru/"),
			expectedBody: "http://localhost:8000/MGRkMTk",
		},
		{
			name:    "positive test #2",
			request: "/",
			store: store.LinkStorage{
				Store: []store.Link{},
			},
			statusCode:   http.StatusCreated,
			body:         []byte("EwHXdJfB"),
			expectedBody: "http://localhost:8000/ODczZGQ",
		},
		{
			name:    "negative test",
			request: "/",
			store: store.LinkStorage{
				Store: []store.Link{},
			},
			statusCode:   http.StatusCreated,
			body:         []byte(""),
			expectedBody: "http://localhost:8000/ZDQxZDh",
		},
	}

	conf := config.Config{
		FlagSaveToFile: false,
		FlagSaveToDB:   false,
		FlagBaseAddr:   "http://localhost:8000/",
	}

	for _, v := range tests {
		ts := httptest.NewServer(runTestServer(&v.store, conf, nil))
		defer ts.Close()

		body := strings.NewReader(string(v.body))
		resp := testRequest(t, ts, "POST", v.request, body)
		defer resp.Body.Close()

		assert.Equal(t, v.statusCode, resp.StatusCode)

		resBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, v.expectedBody, string(resBody))

	}
}
