package app

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	log "github.com/RyanTrue/go-shortener-url/internal/app/logger"
	"github.com/RyanTrue/go-shortener-url/internal/app/models"
	"github.com/RyanTrue/go-shortener-url/internal/app/session"
	errs "github.com/RyanTrue/go-shortener-url/storage/errors"
	"github.com/RyanTrue/go-shortener-url/storage/model"
	"github.com/RyanTrue/go-shortener-url/util"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const uniqueViolation = `ERROR: duplicate key value violates unique constraint "urls_original_url_idx" (SQLSTATE 23505)`

func ReceiveURLAPI(handler Handler, w http.ResponseWriter, r *http.Request) {
	handler.Logger.Sugar.Debug("ReceiveURLAPI")

	var req models.Request

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		handler.Logger.Sugar.Debug("ReceiveURLAPI cannot decode request JSON body; err = ", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	shortURL := util.Shorten(req.URL)

	var userID uuid.UUID
	var ok bool

	cookie, err := r.Cookie("token")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			userID = ctx.Value(session.UserIDKey).(uuid.UUID)
			handler.Logger.Sugar.Debug("ReceiveURLAPI userID = ", userID)
		} else {
			handler.Logger.Sugar.Debug("ReceiveURLAPI Cookie err = ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		userID, ok = session.GetUserID(cookie.Value)
		if !ok {
			handler.Logger.Sugar.Debug("ReceiveURLAPI GetUserID userID not ok")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	linkModel, err := model.MakeLinkModel("", userID, shortURL, req.URL)
	if err != nil {
		handler.Logger.Sugar.Debug("ReceiveURLAPI MakeLinkModel err = ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = handler.Service.Storage.Save(ctx, linkModel)
	if err != nil {
		if err.Error() == uniqueViolation {
			sendJSONRespSingleURL(w, handler.FlagBaseAddr, shortURL, http.StatusConflict, handler.Logger)
			return
		}
		handler.Logger.Sugar.Debug("ReceiveURLAPI handler.Service.Storage.Save err = ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sendJSONRespSingleURL(w, handler.FlagBaseAddr, shortURL, http.StatusCreated, handler.Logger)

}

func sendJSONRespSingleURL(w http.ResponseWriter, flagBaseAddr, short string, statusCode int, logger log.Logger) error {
	resp := models.Response{
		Result: "",
	}

	path, err := util.MakeURL(flagBaseAddr, short)
	if err != nil {
		logger.Sugar.Debug("cannot make path", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	resp.Result = path

	setHeader(w, "Content-Type", "application/json", statusCode)

	respJSON, err := json.Marshal(resp)
	if err != nil {
		logger.Sugar.Debug("cannot Marshal resp: ", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	_, err = w.Write(respJSON)
	if err != nil {
		logger.Sugar.Debug("cannot Write resp: ", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	logger.Sugar.Debug("respJSON: ", string(respJSON))

	return nil
}

func ReceiveManyURLAPI(handler Handler, w http.ResponseWriter, r *http.Request) {
	handler.Logger.Sugar.Debug("ReceiveManyURLAPI")

	var requestArr []models.RequestAPI
	var responseArr []models.ResponseAPI

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&requestArr); err != nil {
		handler.Logger.Sugar.Debug("ReceiveManyURLAPI cannot decode request JSON body: ", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	statusCode := http.StatusCreated
	var path string

	var userID uuid.UUID
	var ok bool

	cookie, err := r.Cookie("token")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			userID = ctx.Value(session.UserIDKey).(uuid.UUID)
			handler.Logger.Sugar.Debug("ReceiveManyURLAPI userID = ", userID)
		} else {
			handler.Logger.Sugar.Debug("ReceiveManyURLAPI Cookie err = ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		userID, ok = session.GetUserID(cookie.Value)
		if !ok {
			handler.Logger.Sugar.Debug("ReceiveManyURLAPI GetUserID userID not ok")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	for _, val := range requestArr {
		resp := models.ResponseAPI{ID: val.ID}
		shortURL := util.Shorten(val.URL)

		linkModel, err := model.MakeLinkModel("", userID, shortURL, val.URL)
		if err != nil {
			handler.Logger.Sugar.Debug("ReceiveManyURLAPI MakeLinkModel err = ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = handler.Service.Storage.Save(ctx, linkModel)
		if err != nil {
			if err.Error() == uniqueViolation {
				statusCode = http.StatusConflict
			} else { // if error is not unique violation
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		}

		path, err = util.MakeURL(handler.FlagBaseAddr, shortURL)
		if err != nil {
			handler.Logger.Sugar.Debug("ReceiveManyURLAPI cannot make path: ", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp.ShortURL = path

		responseArr = append(responseArr, resp)

	}

	setHeader(w, "Content-Type", "application/json", statusCode)

	respJSON, err := json.Marshal(responseArr)
	if err != nil {
		handler.Logger.Sugar.Debug("ReceiveManyURLAPI cannot Marshal resp: ", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(respJSON)
	if err != nil {
		handler.Logger.Sugar.Debug("ReceiveManyURLAPI cannot Write resp: ", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	handler.Logger.Sugar.Debug("ReceiveManyURLAPI respJSON Many URL: ", string(respJSON))

}

func GetUserURLS(handler Handler, w http.ResponseWriter, r *http.Request) {
	handler.Logger.Sugar.Debug("GetUserURLS")

	ctx := r.Context()

	var userID uuid.UUID
	cookie, err := r.Cookie("token")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			userID = ctx.Value(session.UserIDKey).(uuid.UUID)
			handler.Logger.Sugar.Debug("GetUserURLS userID = ", userID)
		} else {
			handler.Logger.Sugar.Debug("GetUserURLS Cookie err = ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		userID, _ = session.GetUserID(cookie.Value)
	}

	links, err := handler.Service.Storage.GetUserURLS(ctx, userID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			handler.Logger.Sugar.Debug("GetUserURLS  ErrNotFound: ", zap.Error(err))
			w.WriteHeader(http.StatusNoContent)
			return
		}
		handler.Logger.Sugar.Debug("Storage.GetUserURLS err: ", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := []models.UserLinks{}

	for _, link := range links {
		path, err := util.MakeURL(handler.FlagBaseAddr, link.ShortURL)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		link.ShortURL = path

		resp = append(resp, link)
	}

	respJSON, err := json.Marshal(resp)
	if err != nil {
		handler.Logger.Sugar.Debug("GetUserURLS cannot Marshal links: ", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	setHeader(w, "Content-Type", "application/json", http.StatusOK)
	_, err = w.Write(respJSON)
	if err != nil {
		handler.Logger.Sugar.Debug("GetUserURLS cannot Write resp: ", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func DeleteURL(handler Handler, w http.ResponseWriter, r *http.Request) {
	handler.Logger.Sugar.Debug("DeleteURL")

	ctx := r.Context()

	var userID uuid.UUID
	cookie, err := r.Cookie("token")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			userID = ctx.Value(session.UserIDKey).(uuid.UUID)
			handler.Logger.Sugar.Debug("DeleteURL userID = ", userID)
		} else {
			handler.Logger.Sugar.Debug("DeleteURL Cookie err = ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		userID, _ = session.GetUserID(cookie.Value)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		handler.Logger.Sugar.Debug("DeleteURL io.ReadAll err = ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	handler.Logger.Sugar.Debug("DeleteURL body: ", string(body))
	var reqArr []string

	if err := json.Unmarshal(body, &reqArr); err != nil {
		handler.Logger.Sugar.Debug("DeleteURL json.Unmarshal err = ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	handler.Logger.Sugar.Debug("DeleteURL res: ", reqArr)

	for _, val := range reqArr {
		req := model.DeleteLink{
			UserID:   userID,
			ShortURL: val,
		}

		handler.LinksChan <- req
	}

	w.WriteHeader(http.StatusAccepted)
}
