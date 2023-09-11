package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/RyanTrue/go-shortener-url/internal/app/logger"
	"github.com/RyanTrue/go-shortener-url/internal/app/models"
	"github.com/RyanTrue/go-shortener-url/internal/app/service"
	store "github.com/RyanTrue/go-shortener-url/storage/memory"
	"github.com/RyanTrue/go-shortener-url/storage/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestReceiveURLAPI(t *testing.T) {
	testCases := []struct {
		name         string
		method       string
		body         models.Request
		store        store.Memory
		request      string
		expectedCode int
		expectedBody models.Response
	}{
		{
			name:   "positive test",
			method: http.MethodPost,
			body:   models.Request{URL: "https://practicum.yandex.ru"},
			store: store.Memory{
				Store: []model.Link{},
			},
			request:      "/api/shorten",
			expectedCode: http.StatusCreated,
			expectedBody: models.Response{
				Result: `http://localhost:8000/NmJkYjV`,
			},
		},
	}

	h := Handler{
		FlagBaseAddr: "http://localhost:8000/",
	}

	for _, v := range testCases {
		logger, err := makeLogger()
		require.NoError(t, err)

		memory := &v.store
		memory.Logger = logger

		srv := service.New(memory)
		h.Service = srv

		r, err := runTestServer(h)
		require.NoError(t, err)

		ts := httptest.NewServer(r)
		defer ts.Close()

		bodyJSON, err := json.Marshal(v.body)
		require.NoError(t, err)

		resp := testRequest(t, ts, v.method, v.request, bytes.NewReader(bodyJSON))
		defer resp.Body.Close()

		assert.Equal(t, v.expectedCode, resp.StatusCode)

		var result models.Response
		dec := json.NewDecoder(resp.Body)
		err = dec.Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, v.expectedBody, result)
	}
}

func TestReceiveManyURLAPI(t *testing.T) {
	type args struct {
		method       string
		request      string
		expectedCode int
		body         []models.RequestAPI
		expectedBody []models.ResponseAPI
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "positive test without DB",
			args: args{
				method:       http.MethodPost,
				request:      "/api/shorten/batch",
				expectedCode: http.StatusCreated,
				body: []models.RequestAPI{
					{
						ID:  "e169d217-d3c8-493a-930f-7432368139c7",
						URL: "mail2.ru",
					},
					{
						ID:  "c82b937d-c303-40e1-a655-ab085002dfa0",
						URL: "https://practicum.yandex.ru",
					},
					{
						ID:  "cd53c344-fb57-42cf-b576-823476f90918",
						URL: "EwHXdJfB",
					}},

				expectedBody: []models.ResponseAPI{
					{
						ID:       "e169d217-d3c8-493a-930f-7432368139c7",
						ShortURL: "http://localhost:8000/NjYyNjB",
					},
					{
						ID:       "c82b937d-c303-40e1-a655-ab085002dfa0",
						ShortURL: "http://localhost:8000/NmJkYjV",
					},
					{
						ID:       "cd53c344-fb57-42cf-b576-823476f90918",
						ShortURL: "http://localhost:8000/ODczZGQ",
					}},
			},
		},
	}

	h := Handler{
		FlagBaseAddr: "http://localhost:8000/",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.Logger{}

			zapLogger, err := zap.NewDevelopment()
			require.NoError(t, err)

			defer zapLogger.Sync()

			sugar := *zapLogger.Sugar()

			logger.Sugar = sugar
			h.Logger = logger

			memory, err := store.New(h.Logger)
			require.NoError(t, err)

			srv := service.New(memory)
			h.Service = srv

			h.Service = srv

			r, err := runTestServer(h)
			require.NoError(t, err)

			ts := httptest.NewServer(r)
			defer ts.Close()

			bodyJSON, err := json.Marshal(tt.args.body)
			require.NoError(t, err)

			resp := testRequest(t, ts, tt.args.method, tt.args.request, bytes.NewReader(bodyJSON))
			defer resp.Body.Close()

			assert.Equal(t, tt.args.expectedCode, resp.StatusCode)

			var result []models.ResponseAPI

			dec := json.NewDecoder(resp.Body)
			err = dec.Decode(&result)
			require.NoError(t, err)

			assert.Equal(t, tt.args.expectedBody, result)
		})
	}
}
