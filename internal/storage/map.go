package storage

import "errors"

type MapStorage struct {
	pairs map[string]string
}

func NewMapStorage() (*MapStorage, error) {
	pairs := make(map[string]string)
	return &MapStorage{pairs: pairs}, nil
}

func (m *MapStorage) Get(shortURL string) (*URLData, error) {
	data := &URLData{}
	if m.pairs[shortURL] == "" {
		return data, errors.New("ссылка не найдена")
	}
	data.ShortURL = shortURL
	data.OriginalURL = m.pairs[shortURL]
	return data, nil
}

func (m *MapStorage) FindByOriginalURL(originalURL string) (*URLData, error) {
	data := &URLData{}
	for k, v := range m.pairs {
		if v == originalURL {
			data.OriginalURL = v
			data.ShortURL = k
			break
		}
	}
	return data, nil
}

func (m *MapStorage) Post(data *URLData) error {
	if m.pairs[data.ShortURL] != "" {
		return errors.New("ссылка уже существует")
	}
	m.pairs[data.ShortURL] = data.OriginalURL
	return nil
}

func (m *MapStorage) PostBatch(data []*URLData) error {
	for _, u := range data {
		err := Post(u)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MapStorage) Ping() error {
	return nil
}

func (m *MapStorage) Close() error {
	clear(m.pairs)
	return nil
}
