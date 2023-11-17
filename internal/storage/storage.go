package storage

import (
	"errors"
	"github.com/gerasimovpavel/shortener.git/internal/config"
)

var ErrDataConflict = errors.New("дубликат данных")

type Storage interface {
	Get(shortURL string) (*URLData, error)
	Post(data *URLData) error
	PostBatch(urls []*URLData) error
	FindByOriginalURL(originalURL string) (*URLData, error)
	Ping() error
	Close() error
	GetUserURL(userID string) ([]*URLData, error)
}

var Stor Storage

type URLData struct {
	UUID        string `json:"uuid,omitempty" db:"uuid"`
	CorrID      string `json:"correlation_id,omitempty"`
	ShortURL    string `json:"short_url,omitempty" db:"shortURL"`
	OriginalURL string `json:"original_url,omitempty" db:"originalURL"`
	UserID      string `json:"user_id,omitempty" db:"userID"`
}

func NewStorage() (Storage, error) {
	if config.Options.DatabaseDSN != "" {
		return NewPostgreWorker(config.Options.DatabaseDSN)
	}
	if config.Options.FileStoragePath != "" {
		return NewFileWorker(config.Options.FileStoragePath)
	}
	return NewMemWorker()
}
