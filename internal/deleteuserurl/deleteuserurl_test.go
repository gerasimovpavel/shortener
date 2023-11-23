package deleteuserurl

import (
	"fmt"
	"github.com/brianvoe/gofakeit"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
	"testing"
)

func Test_DeleteUserURL(t *testing.T) {
	gofakeit.Seed(0)

	var err error
	storage.Stor, err = storage.NewPostgreWorker("host=localhost user=shortener password=shortener dbname=shortener sslmode=disable")
	if err != nil {
		panic(fmt.Errorf("failed to create storage: %v", err))
	}
	URLDel = NewURLDeleter()

	userID := gofakeit.UUID()
	urls := []*storage.URLData{}
	for i := 0; i < 20; i++ {
		urls = append(urls, &storage.URLData{CorrID: gofakeit.UUID(), OriginalURL: gofakeit.URL(), UserID: userID})
	}

	err = storage.Stor.PostBatch(urls)
	if err != nil {
		panic(fmt.Errorf("failed to post to storage: %v", err))
	}

	urls, err = storage.Stor.GetUserURL(userID)
	var list []string
	for _, url := range urls {
		list = append(list, url.ShortURL)
	}
	URLDel.AddURL(&list)
}
