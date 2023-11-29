package storage

import (
	"errors"
	urlgen "github.com/gerasimovpavel/shortener.git/internal/urlgenerator"
)

type MapStorage []URLData

func NewMemWorker() (*MapStorage, error) {
	return &MapStorage{}, nil
}

func (m *MapStorage) Get(shortURL string) (*URLData, error) {
	for _, data := range *m {
		if data.ShortURL == shortURL {
			return &data, nil
		}
	}
	return &URLData{}, nil
}

func (m *MapStorage) FindByOriginalURL(originalURL string) (*URLData, error) {
	for _, data := range *m {
		if data.OriginalURL == originalURL {
			return &data, nil
		}
	}
	return &URLData{}, nil
}

func (m *MapStorage) PostBatch(data []*URLData) error {
	var errConf error
	for _, u := range data {
		err := m.Post(u)
		if err != nil && !errors.Is(err, ErrDataConflict) {
			return err
		}
		if err != nil {
			errConf = errors.Join(err, errConf)
		}
	}
	return errors.Join(errConf, nil)
}

func (m *MapStorage) Post(data *URLData) error {
	var errConf error
	if data.ShortURL == "" {
		data.ShortURL = urlgen.GenShort()
	}
	item, err := m.FindByOriginalURL(data.OriginalURL)
	if err != nil {
		return err
	}
	if item.ShortURL != "" {
		errConf = errors.Join(errConf, ErrDataConflict)
	}
	item, err = m.Get(data.ShortURL)
	if err != nil {
		return err
	}
	if item.ShortURL != "" {
		errConf = errors.Join(errConf, ErrDataConflict)
	}
	*m = append(*m, *data)
	return errors.Join(nil, errConf)
}

func (m *MapStorage) Ping() error {
	return nil
}

func (m *MapStorage) Close() error {
	m = nil
	return nil
}

func (m *MapStorage) GetUserURL(userID string) ([]*URLData, error) {
	urls := []*URLData{}
	for _, data := range *m {
		if data.UserID == userID {
			urls = append(urls, &data)
		}
	}
	return urls, nil
}

func (m *MapStorage) DeleteUserURL(urls []*URLData) error {
	for _, deldata := range urls {
		for _, data := range *m {
			if data.UserID == deldata.UserID && data.ShortURL == deldata.ShortURL && !data.DeletedFlag {
				data.DeletedFlag = true
			}
		}
	}
	return nil
}
