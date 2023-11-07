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
	file    *os.File
	encoder *json.Encoder
	decoder *json.Decoder
}

func NewFileWorker(filename string) (*FileWorker, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &FileWorker{file: file,
		encoder: json.NewEncoder(file),
		decoder: json.NewDecoder(file)}, nil
}

func (fw *FileWorker) rowsCount() (int, error) {
	var cnt int
	scanner := bufio.NewScanner(fw.file)

	for scanner.Scan() {
		cnt++
	}

	if err := scanner.Err(); err != nil {
		return -1, err
	}
	return cnt, nil
}

func (fw *FileWorker) Post(data *URLData) error {
	item, err := fw.Get(data.ShortURL)
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
	for {
		err := fw.decoder.Decode(&item)
		if err != nil {
			return nil, err
		}
		if item.ShortURL == shortURL || err == io.EOF {
			break
		}
	}
	return item, nil
}

func (fw *FileWorker) FindByOriginalURL(originalURL string) (*URLData, error) {
	data := &URLData{}
	items, err := fw.GetAll()
	if err != nil {
		return data, err
	}
	for _, item := range *items {
		if item.OriginalURL == originalURL {
			data = &item
			break
		}
	}
	return data, nil
}

func (fw *FileWorker) GetAll() (*[]URLData, error) {
	items := []URLData{}
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
	return &items, nil
}

func (fw *FileWorker) Ping() error {
	return nil
}

func (fw *FileWorker) Close() error {
	return fw.file.Close()
}
