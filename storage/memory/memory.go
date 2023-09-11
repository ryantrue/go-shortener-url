package memory

import (
	"context"

	"github.com/RyanTrue/go-shortener-url/internal/app/logger"
	"github.com/RyanTrue/go-shortener-url/internal/app/models"
	errs "github.com/RyanTrue/go-shortener-url/storage/errors"
	"github.com/RyanTrue/go-shortener-url/storage/model"
	"github.com/google/uuid"
)

type Memory struct {
	Store  []model.Link
	Logger logger.Logger
}

func New(logger logger.Logger) (*Memory, error) {
	memory := &Memory{}
	memory.Store = []model.Link{}
	memory.Logger = logger

	return memory, nil
}

func (s *Memory) Get(ctx context.Context, short string) (string, bool, error) {
	s.Logger.Sugar.Debug("GetLinkByID")

	s.Logger.Sugar.Debug("shortURL = ", short)
	s.Logger.Sugar.Debug("s.Store = ", s.Store)

	for _, val := range s.Store {
		if val.ShortURL == short {
			return val.OriginalURL, false, nil
		}
	}

	return "", false, errs.ErrNotFound
}

func (s *Memory) Save(ctx context.Context, link model.Link) error {
	s.Logger.Sugar.Debug("SaveLink")

	s.Logger.Sugar.Debug("shortURL = ", link.ShortURL, "original URL = ", link.OriginalURL)

	s.Store = append(s.Store, link)

	return nil
}

func (s *Memory) GetUserURLS(ctx context.Context, userID uuid.UUID) ([]models.UserLinks, error) {
	s.Logger.Sugar.Debug("(s *Memory) GetUserURLS")
	res := []models.UserLinks{}

	for _, val := range s.Store {
		if val.UserID == userID {
			link := models.UserLinks{
				ShortURL:    val.ShortURL,
				OriginalURL: val.OriginalURL,
			}
			res = append(res, link)
		}
	}

	if len(res) == 0 {
		return res, errs.ErrNotFound
	}

	return res, nil
}
