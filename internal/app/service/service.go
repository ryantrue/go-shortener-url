package service

import (
	"context"

	"github.com/RyanTrue/go-shortener-url/internal/app/logger"
	"github.com/RyanTrue/go-shortener-url/storage/model"
)

type Service struct {
	Storage Storage
}

type Storage interface {
	Get(ctx context.Context, short string, logger logger.Logger) (string, error)
	Save(ctx context.Context, link model.Link, logger logger.Logger) error
}

func New(storage Storage) *Service {
	return &Service{
		Storage: storage,
	}
}
