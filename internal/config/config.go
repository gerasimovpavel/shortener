// Package config реализует управоение настройками через чтение переменных окружения
// или параметро командной строки
package config

import (
	flag "github.com/spf13/pflag"
	"os"
)

// Options Опции для запуска сервера
var Options struct {
	// Адрес сервера
	Host string
	// Адрес хоста при формировании короткой ссылки
	ShortURLHost string
	// Путь к файловому хранилищу
	FileStoragePath string
	// Строка подключения к базе данных
	DatabaseDSN string
	// Секретный ключ для формирования UserID
	PassphraseKey string
}

// ParseEnvFlags Обработка окружения и флагов для формирования конфигурации
func ParseEnvFlags() {
	var ok bool
	Options.PassphraseKey, ok = os.LookupEnv("PASSPHRASE_KEY")
	if !ok {
		flag.StringVarP(&Options.PassphraseKey, "k", "k", "", "Пароль для ключа")
	}
	Options.DatabaseDSN, ok = os.LookupEnv("DATABASE_DSN")
	if !ok {
		flag.StringVarP(&Options.DatabaseDSN, "d", "d", "", "Строка подключения к БД")
	}
	Options.FileStoragePath, ok = os.LookupEnv("FILE_STORAGE_PATH")
	if !ok {
		flag.StringVarP(&Options.FileStoragePath, "f", "f", "/tmp/short-url-db.json", "Путь к файлу для сохраненных ссылок")
	}
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
	// если хотя бы одну переменную ищем в командной строке
	if !ok {
		// парсим аргументы
		flag.Parse()
	}
}
