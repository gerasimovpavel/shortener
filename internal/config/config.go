package config

import (
	flag "github.com/spf13/pflag"
	"os"
)

var Options struct {
	Host            string
	ShortURLHost    string
	FileStoragePath string
}

func ParseEnvFlags() {
	var ok bool
	// ищем переменную SERVER_ADDRESS
	Options.Host, ok = os.LookupEnv(`SERVER_ADDRESS`)
	if !ok {
		// если не нашли, обрабатываем командную строку
		flag.StringVarP(&Options.Host, "a", "a", ":8080", "Адрес HTTP-сервера")
	}
	// ищем переменную BASE_URL
	Options.ShortURLHost, ok = os.LookupEnv(`BASE_URL`)
	if !ok {
		// если не нашли, обрабатываем командную строку
		flag.StringVarP(&Options.ShortURLHost, "b", "b", "http://localhost:8080", "URL короткой ссылки")
	}
	Options.FileStoragePath, ok = os.LookupEnv("FILE_STORAGE_PATH")
	if !ok {
		flag.StringVarP(&Options.FileStoragePath, "f", "f", "/tmp/short-url-db.json", "Путь к файлу для сохраненных ссылок")
	}
	// если хотя бы одну переменную ищем в командной строке
	if !ok {
		// парсим аргументы
		flag.Parse()
	}
}
