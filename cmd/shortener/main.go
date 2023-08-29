package main

import (
	"net/http"

	"github.com/RyanTrue/go-shortener-url/config"
	internal "github.com/RyanTrue/go-shortener-url/internal/app"
	"github.com/RyanTrue/go-shortener-url/internal/app/compress"
	log "github.com/RyanTrue/go-shortener-url/internal/app/logger"
	"github.com/RyanTrue/go-shortener-url/internal/app/service"
	"github.com/RyanTrue/go-shortener-url/internal/app/session"
	storage "github.com/RyanTrue/go-shortener-url/storage/db"
	file "github.com/RyanTrue/go-shortener-url/storage/file"
	memory "github.com/RyanTrue/go-shortener-url/storage/memory"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

func main() {
	conf := config.ParseConfigAndFlags()

	logger := log.Logger{}

	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		zapLogger.Fatal("error while creating sugar: ", zap.Error(err))
	}
	defer zapLogger.Sync()

	sugar := *zapLogger.Sugar()

	logger.Sugar = sugar

	logger.Sugar.Infow(
		"Starting server",
		"addr", conf.FlagRunAddr,
	)

	var srv *service.Service
	var db *storage.URLStorage

	if conf.FlagSaveToDB {
		conn, err := storage.Connect(conf.FlagDatabaseAddress)
		if err != nil {
			logger.Sugar.Fatal("error while creating db connection: ", zap.Error(err))
		}

		db, err = storage.New(conn, logger)
		if err != nil {
			logger.Sugar.Fatal("error while creating db: ", zap.Error(err))
		}

		srv = service.New(db)
	} else if conf.FlagSaveToFile {
		storage, err := file.New(conf.FlagPathToFile, logger)
		if err != nil {
			logger.Sugar.Fatal("error while creating file storage: ", zap.Error(err))
		}

		srv = service.New(storage)
	} else {
		storage, err := memory.New(logger)
		if err != nil {
			logger.Sugar.Fatal("error while creating memory storage: ", zap.Error(err))
		}
		srv = service.New(storage)
	}

	handler := internal.Handler{
		Service:      srv,
		Logger:       logger,
		FlagBaseAddr: conf.FlagBaseAddr,
	}

	if err := http.ListenAndServe(conf.FlagRunAddr, Run(handler, db)); err != nil {
		logger.Sugar.Fatal("error while executing server: ", zap.Error(err))
	}
}

func Run(handler internal.Handler, db *storage.URLStorage) chi.Router {
	r := chi.NewRouter()

	r.Use(session.CookieMiddleware)
	r.Use(compress.UnpackData)
	r.Use(handler.Logger.WithLogging)

	r.Use(middleware.Compress(5, "application/javascript",
		"application/json",
		"text/css",
		"text/html",
		"text/plain",
		"text/xml"))

	r.Get("/{id}", func(rw http.ResponseWriter, r *http.Request) {
		internal.GetURL(handler, rw, r)
	})

	r.Post("/", func(rw http.ResponseWriter, r *http.Request) {
		internal.ReceiveURL(handler, rw, r)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.AllowContentType("application/json"))
		r.Route("/api", func(r chi.Router) {
			r.Post("/shorten", func(rw http.ResponseWriter, r *http.Request) {
				internal.ReceiveURLAPI(handler, rw, r)
			})

			r.Post("/shorten/batch", func(rw http.ResponseWriter, r *http.Request) {
				internal.ReceiveManyURLAPI(handler, rw, r)
			})

			r.Get("/user/urls", func(w http.ResponseWriter, r *http.Request) {
				internal.GetUserURLS(handler, w, r)
			})
		})
	})

	r.Get("/ping", func(rw http.ResponseWriter, r *http.Request) {
		internal.Ping(rw, r, db)
	})

	return r
}
