package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/RyanTrue/go-shortener-url/config"
	"github.com/RyanTrue/go-shortener-url/storage"
	"github.com/RyanTrue/go-shortener-url/util"
	"github.com/go-chi/chi"
)

func ReceiveURL(memory *storage.LinkStorage, w http.ResponseWriter, r *http.Request, conf config.Config, db *storage.Database) {
	fmt.Println("ReceiveUrl")

	// сократить ссылку
	// записать в базу

	j, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	short := util.Shorten(string(j))

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	memory.SaveLink(ctx, "", short, string(j), conf.FlagSaveToFile, conf.FlagSaveToDB, db)

	path, err := util.MakeURL(conf.FlagBaseAddr, short)
	if err != nil {
		fmt.Println("err: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setHeader(w, "Content-Type", "text/plain", http.StatusCreated)
	w.Write([]byte(path))
}

func GetURL(memory *storage.LinkStorage, w http.ResponseWriter, r *http.Request, conf config.Config, db *storage.Database) {
	fmt.Println("GetUrl")

	// проверить наличие ссылки в базе
	// выдать ссылку

	id := chi.URLParam(r, "id")

	fmt.Println("url = ", id)

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	val, err := memory.GetLinkByID(ctx, id, conf.FlagSaveToFile, conf.FlagSaveToDB, db)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	setHeader(w, "Location", val, http.StatusTemporaryRedirect)
}

func Ping(w http.ResponseWriter, r *http.Request, db *storage.Database, flagDB bool) {
	// ping

	if flagDB {
		err := db.Ping()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusForbidden)
	}

}

func setHeader(w http.ResponseWriter, header string, val string, statusCode int) {
	w.Header().Set(header, val)
	w.WriteHeader(statusCode)
}
