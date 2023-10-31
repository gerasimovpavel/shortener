package storage

import (
	"encoding/json"
	"io"
	"os"
)

type FileStruct struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{file: file,
		encoder: json.NewEncoder(file)}, nil
}

func (p *Producer) WriteItem(item *FileStruct) error {
	return p.encoder.Encode(&item)
}

func (p *Producer) Close() error {
	return p.file.Close()
}

type Consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &Consumer{file: file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *Consumer) Read() (*[]FileStruct, error) {
	items := []FileStruct{}
	for {
		item := FileStruct{}
		err := c.decoder.Decode(&item)
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

func (c *Consumer) ReadItem() (*FileStruct, error) {
	item := &FileStruct{}
	if err := c.decoder.Decode(&item); err != nil {
		return nil, err
	}
	return item, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}
