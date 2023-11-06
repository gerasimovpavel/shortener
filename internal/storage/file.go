package storage

import (
	"encoding/json"
	"io"
	"os"
)

type URLData struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

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

func (fw *FileWorker) WriteItem(item *URLData) error {
	return fw.encoder.Encode(&item)
}

func (fw *FileWorker) Read() (*[]URLData, error) {
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

func (fw *FileWorker) Close() error {
	return fw.file.Close()
}
