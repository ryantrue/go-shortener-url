package service

import (
	"context"

	"github.com/RyanTrue/go-shortener-url/internal/app/models"
	"github.com/RyanTrue/go-shortener-url/storage/model"
	"github.com/google/uuid"
)

type Service struct {
	Storage Storage
}

type Storage interface {
	Get(ctx context.Context, short string) (string, error)
	GetUserURLS(ctx context.Context, userID uuid.UUID) ([]models.UserLinks, error)
	Save(ctx context.Context, link model.Link) error
}

func New(storage Storage) *Service {
	return &Service{
		Storage: storage,
	}
}
