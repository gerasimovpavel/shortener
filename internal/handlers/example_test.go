package handlers

import (
	"fmt"
	"github.com/brianvoe/gofakeit"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
)

func saveURL(URL string) (string, error) {
	storage.Stor.Post(&storage.URLData{OriginalURL: URL})
	return "", nil
}
func ExampleGetHandler() {
	var err error
	storage.Stor, err = storage.NewStorage()
	if err != nil {
		panic(err)
	}

	URL := gofakeit.URL()
	shortURL, err := saveURL(URL)
	if err != nil {
		panic(err)
	}
	r := httptest.NewRequest("GET", fmt.Sprintf("/%s", shortURL), nil)
	w := httptest.NewRecorder()

	var res *http.Response
	GetHandler(w, r)
	res = w.Result()

	b, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		panic(err)
	}

	if !assert.Equal(nil, URL, string(b)) {
		panic(fmt.Errorf("url expect %v actual %v", URL, string(b)))
	}
}
