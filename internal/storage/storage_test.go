package storage

import (
	"errors"
	"fmt"
	"github.com/brianvoe/gofakeit"
	"reflect"
	"testing"
)

// includeDatabase пришлось доави ть так как не проходит автотест 2 инкремента,
// потому что в нем нет подключения  СУБД
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
	return &URLData{"", urls[0].CorrelationID, urls[0].ShortURL, urls[0].OriginalURL}
}

func getURLDataBatch() []*URLData {
	batch := []*URLData{}
	for _, url := range urls {
		batch = append(batch, &URLData{"", url.CorrelationID, url.ShortURL, url.OriginalURL})
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
			"ping map storage",
			"Ping",
			[]reflect.Value{},
		},
		{
			"post map storage",
			"Post",
			[]reflect.Value{reflect.ValueOf(getURLData())},
		},
		{
			"post batch map storage",
			"PostBatch",
			[]reflect.Value{reflect.ValueOf(getURLDataBatch())},
		},
		{
			"get map storage",
			"Get",
			[]reflect.Value{},
		},
		{
			"close map storage",
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
				Stor, err = NewMapStorage()
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
					storname = "postgres"
					Stor, err = NewPgStorage("host=localhost user=shortener password=shortener dbname=shortener sslmode=disable")

				}
			}
		default:
			panic("something wrong")
		}
		if err != nil {
			panic(fmt.Errorf("storage: %s. failed to create store: %v", storname, err))
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
							panic(fmt.Errorf("storage: %s method:  %s. failed to method call: %v", storname, tt.method, err))
						}
					}

				}

			})
		}
	}
}
