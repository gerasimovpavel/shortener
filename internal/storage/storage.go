package storage

import (
	"errors"
	"fmt"
	"github.com/gerasimovpavel/shortener.git/internal/config"
	"github.com/gerasimovpavel/shortener.git/internal/log"
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
			c, err := NewConsumer(config.Options.FileStoragePath)
			if err != nil {
				return "", err
			}
			items, err := c.Read()
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
			p, err := NewProducer(config.Options.FileStoragePath)
			if err != nil {
				return err
			}
			_, ok := FindByValue(value)
			if !ok {

				item := &FileStruct{}

				c, err := NewConsumer(config.Options.FileStoragePath)
				if err != nil {
					return err
				}
				items, err := c.Read()
				if err != nil {
					return err
				}
				l := strconv.Itoa(len(*items) + 1)

				item.Uuid = l
				item.ShortURL = key
				item.OriginalURL = value

				err = p.WriteItem(item)
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
			c, err := NewConsumer(config.Options.FileStoragePath)
			if err != nil {
				log.Sugar.Error(fmt.Errorf("failed to create consumer: %v", err))
			}
			defer c.Close()
			items, err := c.Read()
			if err != nil {
				log.Sugar.Error(fmt.Errorf("failed to read data: %v", err))
			}
			for _, item := range *items {
				if item.OriginalURL == value {
					key = item.ShortURL
					ok = true
					return
					break
				}
			}

		}
	}
	return
}
