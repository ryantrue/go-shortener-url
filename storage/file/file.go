package storage

import (
	"context"
	"encoding/json"
	"io"
	"os"

	log "github.com/RyanTrue/go-shortener-url/internal/app/logger"
	"github.com/RyanTrue/go-shortener-url/storage/memory"
	"github.com/RyanTrue/go-shortener-url/storage/model"
)

type FileStorage struct {
	Memory  storage.Memory
	file    *os.File
	encoder *json.Encoder
	decoder *json.Decoder
}

func New(filename string, logger log.Logger) (*FileStorage, error) {
	fileStorage := &FileStorage{}
	if err := os.MkdirAll("tmp", os.ModePerm); err != nil {
		return fileStorage, err
	}

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fileStorage, err
	}

	memory, err := storage.New(logger)
	if err != nil {
		return fileStorage, err
	}

	fileStorage.Memory = *memory
	fileStorage.file = file
	fileStorage.decoder = json.NewDecoder(file)
	fileStorage.encoder = json.NewEncoder(file)

	links, err := fileStorage.RecoverData(logger)
	if err != nil {
		return fileStorage, err
	}

	fileStorage.Memory.Store = links

	return fileStorage, nil
}

func (f *FileStorage) RecoverData(logger log.Logger) ([]model.Link, error) {
	logger.Sugar.Debug("RecoverData")

	links := []model.Link{}

	for {
		var link model.Link
		if err := f.decoder.Decode(&link); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	//logger.Sugar.Debug("links = ", links)

	return links, nil
}

func (f *FileStorage) Save(ctx context.Context, link model.Link, logger log.Logger) error {
	logger.Sugar.Debug("SaveDataToFile")

	logger.Sugar.Debugf("link: %#v\n", link)

	if err := f.Memory.Save(ctx, link, logger); err != nil {
		return err
	}

	return f.encoder.Encode(&link)
}

func (f *FileStorage) Get(ctx context.Context, short string, logger log.Logger) (string, error) {
	return f.Memory.Get(ctx, short, logger)
}

func (f *FileStorage) Close() error {
	return f.file.Close()
}
