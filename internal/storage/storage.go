package storage

import (
	"github.com/gerasimovpavel/shortener.git/internal/config"
)

type Storage interface {
	Get(shortURL string) (*URLData, error)
	Post(data *URLData) error
	PostBatch(data []*URLData) error
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
	IsConflict  bool   `json:"-"`
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

func Get(shortURL string) (*URLData, error) {
	data, err := Stor.Get(shortURL)
	if err != nil {
		return &URLData{}, err
	}
	return data, nil
}

func Post(data *URLData) error {
	err := Stor.Post(data)
	if err != nil {
		return err
	}
	return nil
}

func PostBatch(data []*URLData) error {
	err := Stor.PostBatch(data)
	if err != nil {
		return err
	}
	return nil
}

func FindByOriginalURL(originalURL string) (*URLData, error) {
	return Stor.FindByOriginalURL(originalURL)
}

func Ping() error {
	return Stor.Ping()
}
