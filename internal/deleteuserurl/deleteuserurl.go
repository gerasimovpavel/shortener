package deleteuserurl

import (
	"errors"
	"github.com/gerasimovpavel/shortener.git/internal/middleware"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
	"sync"
)

func deleteURL(userID string, urls []string, results chan<- error) {
	urlsdata := []*storage.URLData{}
	for _, u := range urls {
		url := &storage.URLData{ShortURL: u,
			UserID: userID}
		urlsdata = append(urlsdata, url)
	}
	err := storage.Stor.DeleteUserURL(urlsdata)
	if err != nil {
		results <- err
	}
}

func DeleteUserURL(userID string, urls []string) {
	var err error

	results := make(chan error)
	jobs := make(chan []string)

	go func() {
		jobs <- urls
		close(jobs)
	}()

	workers := 4

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case urls, ok := <-jobs:
					if !ok {
						return
					}
					deleteURL(userID, urls, results)
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		err = errors.Join(err, result)
	}
	if err != nil {
		middleware.Sugar.Warn("failed to delete user urls: %v", err)
	}
}
