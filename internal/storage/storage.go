package storage

import (
	"errors"
	"fmt"
	"github.com/gerasimovpavel/shortener.git/internal/config"
	"github.com/gerasimovpavel/shortener.git/internal/middleware"
	"strconv"
)

// мапа для хранения ссылок
var Pairs = make(map[string]string)

func Get(key string) (string, error) {
	switch config.Options.FileStoragePath {
	case "":
		{
			if Pairs[key] == "" {
				return "", errors.New("ссылка не найдена")
			}
			return Pairs[key], nil
		}
	default:
		{
			fw, err := NewFileWorker(config.Options.FileStoragePath)
			if err != nil {
				return "", err
			}
			defer fw.Close()
			items, err := fw.Read()
			if err != nil {
				return "", err
			}
			for _, item := range *items {
				if item.ShortURL == key {
					return item.OriginalURL, nil
				}
			}
		}
	}
	return "", errors.New("ссылка не найдена")
}

// Append in storage
func Append(key, value string) error {
	switch config.Options.FileStoragePath {
	case "":
		{
			Pairs[key] = value
		}
	default:
		{
			fw, err := NewFileWorker(config.Options.FileStoragePath)
			if err != nil {
				return err
			}
			defer fw.Close()
			_, ok := FindByValue(value)
			if !ok {

				item := &URLData{}

				items, err := fw.Read()
				if err != nil {
					return err
				}
				l := strconv.Itoa(len(*items) + 1)

				item.UUID = l
				item.ShortURL = key
				item.OriginalURL = value

				err = fw.WriteItem(item)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// FindByValue Поиск ключа по значению пары
func FindByValue(value string) (key string, ok bool) {
	switch config.Options.FileStoragePath {
	case "":
		{
			for k, v := range Pairs {
				if v == value {
					key = k
					ok = true
					return
				}
			}
		}
	default:
		{
			fw, err := NewFileWorker(config.Options.FileStoragePath)
			if err != nil {
				middleware.Sugar.Error(fmt.Errorf("failed to create consumer: %v", err))
			}
			defer fw.Close()
			items, err := fw.Read()
			if err != nil {
				middleware.Sugar.Error(fmt.Errorf("failed to read data: %v", err))
			}
			for _, item := range *items {
				if item.OriginalURL == value {
					key = item.ShortURL
					ok = true
					break
				}
			}

		}
	}
	return
}
