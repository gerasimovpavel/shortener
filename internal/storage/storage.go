package storage

import (
	"errors"
	"github.com/gerasimovpavel/shortener.git/internal/config"
)

type Storage interface {
	Get(shortURL string) (*URLData, error)
	Post(data *URLData) error
	FindByOriginalURL(originalURL string) (*URLData, error)
	Ping() error
	Close() error
}

var Stor Storage

type URLData struct {
	UUID        string `json:"uuid" db:"uuid"`
	ShortURL    string `json:"short_url" db:"shortURL"`
	OriginalURL string `json:"original_url" db:"originalURL"`
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
	data := &URLData{}
	if Stor == nil {
		return data, errors.New("Stor не определен")
	}
	data, err := Stor.Get(shortURL)
	if err != nil {
		return &URLData{}, err
	}
	return data, nil
}

func Post(data *URLData) error {
	if Stor == nil {
		errors.New("Stor не определен")
	}
	err := Stor.Post(data)
	if err != nil {
		return err
	}
	return nil
}

func FindByOriginalURL(originalURL string) (*URLData, error) {
	if Stor == nil {
		return &URLData{}, errors.New("Stor не определен")
	}
	return Stor.FindByOriginalURL(originalURL)
}

func Ping() error {
	if Stor == nil {
		errors.New("Stor не определен")
	}
	return Stor.Ping()
}
