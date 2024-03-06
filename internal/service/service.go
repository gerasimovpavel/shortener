// package service реализует бизнес логику сервиса коротких ссылок
package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/gerasimovpavel/shortener.git/internal/config"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
	"go.uber.org/zap"
	"net/url"
	"reflect"
)

const (
	ErrorJoinURL = "URL cannot be joined: %v"
)

var (
	ErrNoContent = errors.New("no content")
	ErrNotFound  = errors.New("not found")
	ErrIsDeleted = errors.New("deleted")
)

// Service структура сервиса бизнес логики
type Service struct {
	cfg    *config.Options
	logger *zap.Logger
	store  storage.Storage
}

// NewService создание нового сервиса бизнес логики
func NewService(config *config.Options, store storage.Storage, logger *zap.Logger) *Service {
	return &Service{
		cfg:    config,
		logger: logger,
		store:  store,
	}
}

// DeleteUserURL удаление ссылок пользователя
func (s *Service) DeleteUserURL(urls []*storage.URLData) error {
	if err := s.store.DeleteUserURL(urls); err != nil {
		err = fmt.Errorf("error deleting: %w", err)
		s.logger.Sugar().Error(err)
		return err
	}
	return nil
}

// GetUserURL получение ссылок пользователя
func (s *Service) GetUserURL(userID string) ([]*storage.URLData, error) {
	urls, err := s.store.GetUserURL(userID)
	if err != nil {
		err = fmt.Errorf("error getting all user urls: %w", err)
		s.logger.Sugar().Error(err)
		return nil, err
	}

	if len(urls) == 0 {
		return nil, ErrNoContent
	}

	for idx, urlObj := range urls {
		resultURL, err := url.JoinPath(s.cfg.ShortURLHost, urlObj.ShortURL)
		if err != nil {
			err = fmt.Errorf(ErrorJoinURL, err)
			s.logger.Sugar().Error(err)
			return nil, err
		}
		urls[idx].ShortURL = resultURL
	}

	return urls, nil
}

// GetOriginalURL получение оригинальной ссылки по короткой
func (s *Service) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	urlObj, err := s.store.Get(shortURL)
	if err != nil {
		if errors.Is(err, storage.ErrURLDeleted) {
			return "", ErrIsDeleted
		}

		err = fmt.Errorf("error getting original URL: %w", err)
		s.logger.Sugar().Error(err)
		return "", err
	}

	if urlObj.OriginalURL == "" {
		return "", ErrNotFound
	}

	return urlObj.OriginalURL, nil
}

// PostBatch массовое сохранение ссылок
func (s *Service) PostBatch(data []*storage.URLData) ([]*storage.URLData, error) {
	err := s.store.PostBatch(data)
	if err != nil {
		err := fmt.Errorf("cant put batch: %w", err)
		s.logger.Sugar().Error(err)
		return nil, err
	}

	return data, nil
}

// Post сохранение короткой ссылки
func (s *Service) Post(ctx context.Context, data *storage.URLData) (string, error) {

	err := s.store.Post(data)

	if err != nil {
		if errors.Is(err, storage.ErrDataConflict) {
			return "", storage.ErrDataConflict
		}
		err = fmt.Errorf("error saving data: %w", err)
		s.logger.Sugar().Error(err)
		return "", err
	}

	resultURL := data.ShortURL

	return resultURL, nil
}

// Ping проверка работоспособности хранилища
func (s *Service) Ping(ctx context.Context) error {
	if err := s.store.Ping(); err != nil {
		err := fmt.Errorf("error opening connection to DB: %w", err)
		s.logger.Sugar().Error(err)
		return err
	}
	return nil
}

// GetStat получение статистики сервиса коротких ссылок
func (s *Service) GetStat(ctx context.Context) (*storage.StatData, error) {
	if reflect.ValueOf(s.store).IsNil() {
		return nil, errors.New("storage not defined")
	}
	stats, err := s.store.GetStat()
	if err != nil {
		err := fmt.Errorf("GetStat service error: %w", err)
		s.logger.Sugar().Error(err)
		return nil, err
	}
	return stats, nil
}
