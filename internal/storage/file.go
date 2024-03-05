// Package storage реализует чтение и сохранение данных в текстовом файле
package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	urlgen "github.com/gerasimovpavel/shortener.git/internal/urlgenerator"
	"io"
	"os"
	"strconv"
)

// FileWorker Структура для работы с файловым хранилищем
type FileWorker struct {
	decoder  *json.Decoder
	encoder  *json.Encoder
	file     *os.File
	filename string
}

// NewFileWorker Создание нового хранилища
func NewFileWorker(filename string) (*FileWorker, error) {

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &FileWorker{
		filename: filename,
		file:     file,
		encoder:  json.NewEncoder(file),
		decoder:  json.NewDecoder(file)}, nil
}

func (fw *FileWorker) refresh() error {
	var err error
	err = fw.Close()
	if err != nil {
		return err
	}
	fw.file, err = os.OpenFile(fw.filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	fw.decoder = json.NewDecoder(fw.file)
	fw.encoder = json.NewEncoder(fw.file)
	return nil
}

func (fw *FileWorker) GetStat() (*StatData, error) {
	stat := &StatData{}
	var UserID string
	items, err := fw.GetAll()
	if err != nil {
		return stat, err
	}
	for _, item := range items {
		if item.UserID == UserID {
			stat.Users++
			UserID = item.UserID
		}
		stat.URLS++
	}
	return stat, nil
}

func (fw *FileWorker) rowsCount() (int, error) {
	var cnt int
	err := fw.refresh()
	if err != nil {
		return -1, err
	}
	scanner := bufio.NewScanner(fw.file)

	for scanner.Scan() {
		cnt++
	}

	if err = scanner.Err(); err != nil {
		return -1, err
	}
	return cnt, nil
}

// PostBatch Пакетная запись ссылок
func (fw *FileWorker) PostBatch(data []*URLData) error {
	var errConf error
	for _, u := range data {
		err := fw.Post(u)
		if err != nil && !errors.Is(err, ErrDataConflict) {
			return err
		}
		if err != nil {
			errConf = errors.Join(err, errConf)
		}
	}
	return errors.Join(errConf, nil)
}

// Post Запись ссылки
func (fw *FileWorker) Post(data *URLData) error {
	var errConf error
	if data.ShortURL == "" {
		data.ShortURL = urlgen.GenShortOptimized()
	}
	item, err := fw.FindByOriginalURL(data.OriginalURL)
	if err != nil {
		return err
	}

	if item.ShortURL != "" {
		errConf = errors.Join(errConf, ErrDataConflict)
	}

	item, err = fw.Get(data.ShortURL)
	if err != nil {
		return err
	}
	if item.ShortURL != "" {
		errConf = errors.Join(errConf, ErrDataConflict)
	}
	uuid, err := fw.rowsCount()
	if err != nil {
		return err
	}

	data.UUID = strconv.Itoa(uuid + 1)
	return errors.Join(fw.encoder.Encode(&data), errConf)
}

// Get Чтение оргинальной ссылки по значению короткой ссылки
func (fw *FileWorker) Get(shortURL string) (*URLData, error) {

	item := &URLData{}
	err := fw.refresh()
	if err != nil {
		return item, err
	}
	for {
		err = fw.decoder.Decode(&item)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if err == io.EOF {
			break
		}
		if item.ShortURL == shortURL {
			return item, nil
		}
	}
	return &URLData{}, nil
}

// FindByOriginalURL поиск по оригинальной ссылки
func (fw *FileWorker) FindByOriginalURL(originalURL string) (*URLData, error) {
	data := &URLData{}
	items, err := fw.GetAll()
	if err != nil {
		return data, err
	}
	for _, item := range items {
		if item.OriginalURL == originalURL {
			data = &item
			break
		}
	}
	return data, nil
}

// GetAll Чтение все ссылок в хранилище
func (fw *FileWorker) GetAll() ([]URLData, error) {
	items := []URLData{}
	err := fw.refresh()
	if err != nil {
		return items, err
	}
	for {
		item := URLData{}
		err := fw.decoder.Decode(&item)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// Ping Проверка доступности файлового хранилища
func (fw *FileWorker) Ping() error {
	return fw.file.Sync()
}

// Close Закрытие хранилища
func (fw *FileWorker) Close() error {
	return fw.file.Close()
}

// GetUserURL Чтение ссылок определенного пользователя
func (fw *FileWorker) GetUserURL(userID string) ([]*URLData, error) {
	urls := []*URLData{}
	err := fw.refresh()
	if err != nil {
		return urls, err
	}
	item := &URLData{}
	for {
		err = fw.decoder.Decode(&item)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if err == io.EOF {
			break
		}
		if item.UserID == userID {
			urls = append(urls, item)
		}
	}
	return urls, nil
}

// DeleteUserURL Удаление ссылок определенного пользователя
func (fw *FileWorker) DeleteUserURL(urls []*URLData) error {
	return nil
}
