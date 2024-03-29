// Package deleteuserurl реализует пакетное удаление коротких ссылок
// с помощью паттерна FanIn
package deleteuserurl

import (
	"github.com/gerasimovpavel/shortener.git/internal/middleware"
	"github.com/gerasimovpavel/shortener.git/internal/storage"
	"sync"
)

// URLDel Глобальная переменная для структуры для удаления ссылок
var URLDel *URLDeleter

// URLDeleter Структура для удаления ссылок
type URLDeleter struct {
	wg   *sync.WaitGroup
	done chan struct{}
}

// NewURLDeleter Создание удалятора ссылок
func NewURLDeleter() *URLDeleter {
	urldel := &URLDeleter{
		wg:   &sync.WaitGroup{},
		done: make(chan struct{})}
	return urldel
}

func (ud *URLDeleter) generator(urls *[]string) chan []string {
	ch := make(chan []string)
	ud.wg.Add(1)
	go func() {
		defer ud.wg.Done()

		select {
		case <-ud.done:
			close(ch)
			return
		default:
			{
				ch <- *urls
			}
		}

	}()
	return ch
}

// AddURL Добавление короткой ссылки на удаление
func (ud *URLDeleter) AddURL(urls *[]string) {
	g := ud.generator(urls)
	out := ud.merge(g)
	go func() {
		for s := range out {
			ud.deleteURL(s)
		}
	}()
}

func (ud *URLDeleter) merge(in ...<-chan []string) <-chan []string {
	out := make(chan []string)

	output := func(c <-chan []string) {
		for s := range c {
			out <- s
		}
		ud.wg.Done()
	}
	ud.wg.Add(len(in))
	for _, c := range in {
		go output(c)
	}
	go func() {
		ud.wg.Wait()
	}()
	return out
}

func (ud *URLDeleter) deleteURL(list []string) {
	urls := []*storage.URLData{}
	for _, url := range list {
		data := storage.URLData{}

		data.UserID = middleware.UserID
		data.ShortURL = url
		urls = append(urls, &data)
	}

	storage.Stor.DeleteUserURL(urls)

}
