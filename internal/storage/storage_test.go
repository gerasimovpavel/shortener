package storage

import (
	"errors"
	"fmt"
	"github.com/brianvoe/gofakeit"
	"github.com/gerasimovpavel/shortener.git/internal/config"
	"os"
	"reflect"
	"testing"
)

// includeDatabase пришлось добавить так как не проходит автотест 2 инкремента,
// потому что в нем нет подключения к СУБД
const includeDatabase bool = false

var urls = []*struct {
	CorrelationID string `json:"correlation_id,omitempty"`
	OriginalURL   string `json:"original_url,omitempty"`
	ShortURL      string `json:"short_url,omitempty"`
}{
	{gofakeit.UUID(),
		gofakeit.URL(),
		""},
	{gofakeit.UUID(),
		gofakeit.URL(),
		""},
	{gofakeit.UUID(),
		gofakeit.URL(),
		""},
	{gofakeit.UUID(),
		gofakeit.URL(),
		""},
}

func getShortURL(Store Storage, originalURL string) string {
	data, err := Store.FindByOriginalURL(originalURL)
	if err != nil {
		panic(err)
	}
	return data.ShortURL
}

func getURLData() *URLData {
	return &URLData{"", urls[0].CorrelationID, urls[0].ShortURL, urls[0].OriginalURL, gofakeit.UUID(), false}
}

func getURLDataBatch() []*URLData {
	batch := []*URLData{}
	for _, url := range urls {
		batch = append(batch, &URLData{"", url.CorrelationID, url.ShortURL, url.OriginalURL, gofakeit.UUID(), false})
	}
	return batch
}

func Test_Storage(t *testing.T) {

	tests := []struct {
		name   string
		method string
		in     []reflect.Value
	}{
		{
			"ping storage",
			"Ping",
			[]reflect.Value{},
		},
		{
			"post storage",
			"Post",
			[]reflect.Value{reflect.ValueOf(getURLData())},
		},
		{
			"post batch storage",
			"PostBatch",
			[]reflect.Value{reflect.ValueOf(getURLDataBatch())},
		},
		{
			"get storage",
			"Get",
			[]reflect.Value{},
		},
		{
			"user urls storage",
			"GetUserURL",
			[]reflect.Value{reflect.ValueOf(gofakeit.UUID())},
		},
		{
			"close storage",
			"Close",
			[]reflect.Value{},
		},
	}
	var i int
	var err error
	var storname string
	for i = 0; i < 3; i++ {
		switch i {
		case 0:
			{
				storname = "map"
				Stor, err = NewMemWorker()
			}
		case 1:
			{
				storname = "file"
				Stor, err = NewFileWorker("/tmp/short-url-db.json")
			}
		case 2:
			{
				if !includeDatabase {
					t.Skip()
				}
				storname = "postgres"
				Stor, err = NewPostgreWorker("host=localhost port=6513  user=postgres password=a766657h dbname=shortener sslmode=disable")

			}
		default:
			panic("something wrong")
		}
		if err != nil {
			panic(fmt.Errorf("storage: %s. failed to create store: %w", storname, err))
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				storType := reflect.TypeOf(&MapStorage{})
				if _, ok := storType.MethodByName(tt.method); !ok {
					panic(fmt.Errorf("storage: %s method: %s. method not found", storname, tt.method))
				}
				if tt.method == "Get" {
					tt.in = []reflect.Value{reflect.ValueOf(getShortURL(Stor, urls[0].OriginalURL))}
				}
				res := reflect.ValueOf(Stor).MethodByName(tt.method).Call(tt.in)
				var i int
				for i = 0; i < len(res); i++ {
					if res[i].Type().Name() == "error" && res[i].Interface() != nil {
						err = res[i].Interface().(error)
						if err != nil && !errors.Is(err, ErrDataConflict) {
							panic(fmt.Errorf("storage: %s method:  %s. failed to method call: %w", storname, tt.method, err))
						}
					}

				}

			})
		}
	}
}

func TestNewStorage(t *testing.T) {
	tests := []struct {
		name    string
		storage string
	}{
		{
			"new map storage",
			"map",
		},
		{
			"new file storage",
			"file",
		},
		{
			"new postgres storage",
			"pgx",
		},
	}
	config.ParseEnvFlags()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error

			switch tt.storage {
			case "pgx":
				{
					config.Cfg.DatabaseDSN = "host=localhost port=6513  user=postgres password=a766657h dbname=shortener sslmode=disable"
					config.Cfg.FileStoragePath = ""
				}
			case "file":
				{
					config.Cfg.DatabaseDSN = "host=localhost port=6513  user=postgres password=a766657h dbname=shortener sslmode=disable"
					config.Cfg.FileStoragePath = "db.json"
					defer os.Remove(config.Cfg.FileStoragePath)
				}
			default:
				{
					config.Cfg.DatabaseDSN = ""
					config.Cfg.FileStoragePath = ""
				}
			}
			Stor, err = NewStorage()
			if err != nil {
				panic(err)
			}
		})
	}
}

func TestPgxDeleteUserURL(t *testing.T) {
	var err error
	config.ParseEnvFlags()
	Stor, err = NewStorage()
	if err != nil {
		panic(err)
	}
	data := []*URLData{}
	data = append(data, &URLData{
		UserID:   "",
		ShortURL: "",
	},
	)
	err = Stor.DeleteUserURL(data)
	if err != nil {
		panic(err)
	}
}
