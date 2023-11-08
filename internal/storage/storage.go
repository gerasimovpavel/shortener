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
}

var Stor Storage

type URLData struct {
	UUID        string `json:"uuid,omitempty" db:"uuid"`
	CorrID      string `json:"correlation_id,omitempty"`
	ShortURL    string `json:"short_url,omitempty" db:"shortURL"`
	OriginalURL string `json:"original_url,omitempty" db:"originalURL"`
}

func NewStorage() error {
	switch config.Options.DatabaseDSN {
	case "":
		{
			switch config.Options.FileStoragePath {
			case "":
				{
					mStor, err := NewMapStorage()
					if err != nil {
						return err
					}
					err = mStor.Ping()
					if err != nil {
						return err
					}
					Stor = mStor
				}
			default:
				{
					fStor, err := NewFileWorker(config.Options.FileStoragePath)
					if err != nil {
						return err
					}
					Stor = fStor
					err = Stor.Ping()
					if err != nil {
						return err
					}
				}
			}
		}
	default:
		{
			pgStor, err := NewPgStorage(config.Options.DatabaseDSN)
			if err != nil {
				return err
			}
			Stor = pgStor
			err = Stor.Ping()
			if err != nil {
				return err
			}
		}

	}
	return nil
}
