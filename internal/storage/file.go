package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strconv"
)

type FileWorker struct {
	filename string
	file     *os.File
	encoder  *json.Encoder
	decoder  *json.Decoder
}

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
func (fw *FileWorker) PostBatch(data []*URLData) error {
	for _, u := range data {
		err := Post(u)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fw *FileWorker) Post(data *URLData) error {
	item, err := fw.FindByOriginalURL(data.OriginalURL)
	if err != nil {
		return err
	}
	if item.ShortURL != "" {
		return errors.New("ссылка уже существует")
	}
	item, err = fw.Get(data.ShortURL)
	if err != nil {
		return err
	}
	if item.ShortURL != "" {
		return errors.New("ссылка уже существует")
	}
	uuid, err := fw.rowsCount()
	if err != nil {
		return err
	}
	data.UUID = strconv.Itoa(uuid + 1)
	return fw.encoder.Encode(&data)
}

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

func (fw *FileWorker) Ping() error {
	return nil
}

func (fw *FileWorker) Close() error {
	return fw.file.Close()
}
