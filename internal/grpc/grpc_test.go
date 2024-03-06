package grpc

import (
	"context"
	"github.com/gerasimovpavel/shortener.git/internal/config"
	pb "github.com/gerasimovpavel/shortener.git/internal/proto"
	service "github.com/gerasimovpavel/shortener.git/internal/service"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"os"
	"testing"
)

func TestCreateShortURL(t *testing.T) {
	Logger, _ := zap.NewDevelopment()
	testCases := []struct {
		name        string
		input       *pb.CreateShortURLRequest
		Resp        string
		Err         error
		expectError bool
	}{
		{
			name: "Successful URL shortening",
			input: &pb.CreateShortURLRequest{
				UserId: "1",
				Url:    "http://example.com",
			},
			expectError: false,
		},
		{
			name:        "Error in URL shortening",
			input:       &pb.CreateShortURLRequest{},
			expectError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage, err := storage.NewFileWorker("./test.json")
			if err != nil {
				t.Errorf("failed to initialize a new storage: %v", err)
				return
			}
			defer os.Remove("./test.json")

			service := service.NewService(&config.Cfg, storage, zap.L())
			grpcService := &GRPCService{logger: Logger.Sugar(), service: service}

			_, err = grpcService.CreateShortURL(context.Background(), tc.input)
			hasError := err != nil
			assert.Equal(t, hasError, tc.expectError)
		})
	}
}

func TestBatchCreateShortURL(t *testing.T) {
	Logger, _ := zap.NewDevelopment()
	testCases := []struct {
		name         string
		request      *pb.BatchCreateShortURLRequest
		Response     []*storage.URLData
		expectedCode codes.Code
	}{
		{
			name: "Successful batch creation",
			request: &pb.BatchCreateShortURLRequest{
				Records: []*pb.BatchCreateShortURLRequestData{
					{
						OriginalUrl:   "http://ya.ru",
						CorrelationId: "cor1",
					},
				},
				UserId: "1",
			},
			Response: []*storage.URLData{{
				UUID:        "",
				CorrID:      "cor1",
				ShortURL:    "",
				OriginalURL: "http://ya.ru",
				UserID:      "1",
				DeletedFlag: false,
			}},
			expectedCode: codes.OK,
		},
		{
			name: "Error from core logic",
			request: &pb.BatchCreateShortURLRequest{
				Records: []*pb.BatchCreateShortURLRequestData{
					{
						OriginalUrl:   "htp://ya.ru",
						CorrelationId: "cor1",
					},
				},
				UserId: "1"},
			expectedCode: codes.Internal,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage, err := storage.NewFileWorker("./test.json")
			if err != nil {
				t.Errorf("failed to initialize a new storage: %v", err)
				return
			}
			defer os.Remove("./test.json")

			service := service.NewService(&config.Cfg, storage, zap.L())
			GRPCservice := &GRPCService{logger: Logger.Sugar(), service: service}

			ctx := context.Background()

			resp, err := GRPCservice.BatchCreateShortURL(ctx, tc.request)

			if tc.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.Len(t, resp.Records, len(tc.Response))
				assert.Equal(t, tc.Response[0].CorrID, resp.Records[0].CorrelationId)
			} else {
				assert.Error(t, err)
				st, _ := status.FromError(err)
				assert.Equal(t, tc.expectedCode, st.Code())
			}
		})
	}
}

func TestGetByShort(t *testing.T) {
	Logger, _ := zap.NewDevelopment()
	testCases := []struct {
		name         string
		request      *pb.GetOriginalURLRequest
		Response     string
		Err          error
		expectedCode codes.Code
	}{
		{
			name:         "Successful get by short",
			request:      &pb.GetOriginalURLRequest{UserId: "1"},
			Response:     "http://example.com",
			expectedCode: codes.OK,
		},
		{
			name:         "Error get by short",
			request:      &pb.GetOriginalURLRequest{Url: "http://bit.ly/test", UserId: "1"},
			Response:     "http://example.com",
			expectedCode: codes.Internal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage, err := storage.NewFileWorker("./test.json")
			if err != nil {
				t.Errorf("failed to initialize a new storage: %v", err)
				return
			}
			defer os.Remove("./test.json")
			service := service.NewService(&config.Cfg, storage, zap.L())
			GRPCservice := &GRPCService{logger: Logger.Sugar(), service: service}
			ctx := context.Background()

			respShort, err := GRPCservice.CreateShortURL(ctx, &pb.CreateShortURLRequest{
				UserId: "1",
				Url:    tc.Response,
			})
			if err != nil {
				panic(err)
			}

			if tc.expectedCode == codes.OK {
				tc.request.Url = respShort.GetResult()
			}

			resp, err := GRPCservice.GetOriginalURL(ctx, tc.request)

			if tc.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.Equal(t, tc.Response, resp.OriginalUrl)
			} else {
				assert.Error(t, err)
				st, _ := status.FromError(err)
				assert.Equal(t, tc.expectedCode, st.Code())
			}
		})
	}
}

func TestGetUserURLs(t *testing.T) {
	Logger, _ := zap.NewDevelopment()
	testCases := []struct {
		name         string
		request      *pb.GetUserURLsRequest
		Response     []storage.URLData
		expectedCode codes.Code
	}{
		{
			name:         "Successful get URL urls",
			request:      &pb.GetUserURLsRequest{UserId: "1"},
			expectedCode: codes.OK,
		},
		{
			name:         "Error get URL urls",
			request:      &pb.GetUserURLsRequest{UserId: "1"},
			expectedCode: codes.Internal,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage, err := storage.NewFileWorker("./test.json")
			if err != nil {
				t.Errorf("failed to initialize a new storage: %v", err)
				return
			}
			defer os.Remove("./test.json")
			service := service.NewService(&config.Cfg, storage, zap.L())
			GRPCservice := &GRPCService{logger: Logger.Sugar(), service: service}
			ctx := context.Background()

			if tc.expectedCode == codes.OK {
				GRPCservice.CreateShortURL(ctx, &pb.CreateShortURLRequest{
					UserId: "1",
					Url:    "http://ya.ru",
				})
			}

			_, err = GRPCservice.GetUserURLs(ctx, tc.request)
			if tc.expectedCode == codes.OK {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				st, _ := status.FromError(err)
				assert.Equal(t, tc.expectedCode, st.Code())
			}
		})
	}
}

func TestGRPCService_GetStats(t *testing.T) {
	Logger, _ := zap.NewDevelopment()

	testCases := []struct {
		name         string
		expectedCode codes.Code
	}{
		{
			name:         "Successful get stats",
			expectedCode: codes.OK,
		},
		{
			name:         "Error get stats",
			expectedCode: codes.Internal,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := storage.NewFileWorker("./test.json")
			defer os.Remove("./test.json")

			var storage *storage.FileWorker
			if tc.expectedCode == codes.OK {
				storage = s
			}
			service := service.NewService(&config.Cfg, storage, zap.L())
			GRPCservice := &GRPCService{logger: Logger.Sugar(), service: service}
			ctx := context.Background()

			_, err = GRPCservice.GetStats(ctx, &pb.ServiceStatsRequest{})
			if tc.expectedCode == codes.OK {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				st, _ := status.FromError(err)
				assert.Equal(t, tc.expectedCode, st.Code())
			}
		})
	}
}
