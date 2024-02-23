package deleteuserurl

import (
	"fmt"
	"github.com/brianvoe/gofakeit"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
	"testing"
)

// includeDatabase пришлось добавить так как не проходит автотест 2 инкремента,
// потому что в нем нет подключения к СУБД
const includeDatabase bool = true

func Test_DeleteUserURL(t *testing.T) {

	if !includeDatabase {
		t.Skip()
	}
	gofakeit.Seed(0)

	var err error
	storage.Stor, err = storage.NewPostgreWorker("host=localhost port=6513  user=postgres password=a766657h dbname=shortener sslmode=disable")
	if err != nil {
		panic(fmt.Errorf("failed to create storage: %w", err))
	}
	URLDel = NewURLDeleter()

	userID := gofakeit.UUID()
	urls := []*storage.URLData{}
	for i := 0; i < 20; i++ {
		urls = append(urls, &storage.URLData{CorrID: gofakeit.UUID(), OriginalURL: gofakeit.URL(), UserID: userID})
	}

	err = storage.Stor.PostBatch(urls)
	if err != nil {
		panic(fmt.Errorf("failed to post to storage: %w", err))
	}

	urls, err = storage.Stor.GetUserURL(userID)
	if err != nil {
		panic(fmt.Errorf("failed to get from storage: %w", err))
	}
	var list []string
	for _, url := range urls {
		list = append(list, url.ShortURL)
	}
	URLDel.AddURL(&list)
}
