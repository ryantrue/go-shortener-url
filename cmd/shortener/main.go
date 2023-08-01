package main

import (
	"net/http"

	"github.com/RyanTrue/go-shortener-url/config"
	internal "github.com/RyanTrue/go-shortener-url/internal/app"
	log "github.com/RyanTrue/go-shortener-url/internal/app/logger"
	"github.com/RyanTrue/go-shortener-url/storage"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func main() {
	conf := config.ParseConfigAndFlags()

	// создаём предустановленный регистратор zap
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Sugar.Fatal("error while creating sugar: ", zap.Error(err))
	}
	defer logger.Sync()

	// делаем регистратор SugaredLogger
	log.Sugar = *logger.Sugar()

	// записываем в лог, что сервер запускается
	log.Sugar.Infow(
		"Starting server",
		"addr", conf.FlagRunAddr,
	)

	if err := http.ListenAndServe(conf.FlagRunAddr, Run(conf)); err != nil {
		log.Sugar.Fatal("error while executing server: ", zap.Error(err))
	}
}

func Run(conf config.Config) chi.Router {
	storage := storage.New()

	r := chi.NewRouter()
	r.Use(log.WithLogging)

	r.Get("/{id}", func(rw http.ResponseWriter, r *http.Request) {
		internal.GetURL(storage, rw, r)
	})
	r.Post("/", func(rw http.ResponseWriter, r *http.Request) {
		internal.ReceiveURL(storage, rw, r, conf.FlagBaseAddr)
	})

	r.Route("/api", func(r chi.Router) {
		r.Post("/shorten", func(rw http.ResponseWriter, r *http.Request) {
			internal.ReceiveURLAPI(storage, rw, r, conf.FlagBaseAddr)
		})
	})

	return r
}
