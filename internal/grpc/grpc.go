// package grpc реализует сервис для grpc сервера
package grpc

import (
	"context"
	pb "github.com/gerasimovpavel/shortener.git/internal/proto"
	"github.com/gerasimovpavel/shortener.git/internal/service"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCService Структура grpc сервиса
type GRPCService struct {
	pb.UnimplementedShortenerServer
	logger  *zap.SugaredLogger
	service *service.Service
}

// NewService создание нового grpc сервиса
func NewGRPCService(logger *zap.SugaredLogger, service *service.Service) *GRPCService {
	return &GRPCService{logger: logger, service: service}
}

// CreateShortURL создание короткой ссылки
func (g *GRPCService) CreateShortURL(ctx context.Context, req *pb.CreateShortURLRequest,
) (*pb.CreateShortURLResponse, error) {
	data := &storage.URLData{
		UserID:      req.GetUserId(),
		OriginalURL: req.GetUrl(),
	}
	url, err := g.service.Post(ctx, data)
	if err != nil {
		g.logger.Error("post url service err", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &pb.CreateShortURLResponse{Result: url}, nil
}

// BatchCreateShortURL массовое создание коротких ссылок
func (gh *GRPCService) BatchCreateShortURL(
	ctx context.Context,
	req *pb.BatchCreateShortURLRequest,
) (*pb.BatchCreateShortURLResponse, error) {
	records := req.GetRecords()

	items := make([]*storage.URLData, len(records))

	for idx, item := range records {
		items[idx] = &storage.URLData{OriginalURL: item.OriginalUrl, CorrID: item.CorrelationId, UserID: req.GetUserId()}
	}
	res, err := gh.service.PostBatch(items)
	if err != nil {
		gh.logger.Error("post batch service err", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	result := make([]*pb.BatchCreateShortURLResponseData, len(res))
	for idx, item := range res {
		result[idx] =
			&pb.BatchCreateShortURLResponseData{
				ShortUrl:      item.ShortURL,
				CorrelationId: item.CorrID,
			}
	}
	return &pb.BatchCreateShortURLResponse{Records: result}, nil
}

// GetOriginalURL  получение оригинальной ссылки по короткой
func (g *GRPCService) GetOriginalURL(
	ctx context.Context,
	req *pb.GetOriginalURLRequest,
) (*pb.GetOriginalURLResponse, error) {
	originalURL, err := g.service.GetOriginalURL(ctx, req.GetUrl())
	if err != nil {
		g.logger.Error("redirectToOriginal service err", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &pb.GetOriginalURLResponse{OriginalUrl: originalURL}, nil
}

// GetUserURLs получение всех ссылок, созданных пользователем
func (g *GRPCService) GetUserURLs(
	ctx context.Context,
	req *pb.GetUserURLsRequest,
) (*pb.GetUserURLsResponse, error) {
	data, err := g.service.GetUserURL(req.GetUserId())
	if err != nil {
		g.logger.Error("GetUserURLs service err", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	result := make([]*pb.ShortenData, len(data))
	for idx, item := range data {
		result[idx] = &pb.ShortenData{ShortUrl: item.ShortURL, OriginalUrl: item.OriginalURL}
	}
	return &pb.GetUserURLsResponse{Records: result}, nil
}

// DeleteUserURLsBatch удаление ссылок, созданных пользователем
func (g *GRPCService) DeleteUserURLsBatch(
	ctx context.Context,
	req *pb.DeleteUserURLsBatchRequest,
) (*pb.DeleteUserURLsBatchResponse, error) {
	urls := req.GetUrls()
	data := make([]*storage.URLData, len(urls))
	for _, item := range data {
		item.UserID = req.GetUserId()
	}
	if err := g.service.DeleteUserURL(data); err != nil {
		g.logger.Error("DeleteUserURL service err", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &pb.DeleteUserURLsBatchResponse{}, nil
}

// GetStats статистика сервиса коротких ссылок
func (g *GRPCService) GetStats(
	ctx context.Context,
	req *pb.ServiceStatsRequest,
) (*pb.ServiceStatsResponse, error) {
	stats, err := g.service.GetStat(ctx)
	if err != nil {
		g.logger.Error("getStats service error", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &pb.ServiceStatsResponse{Urls: stats.URLS, Users: stats.Users}, nil
}
