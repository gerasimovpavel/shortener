// Package config реализует управоение настройками через чтение переменных окружения
// или параметро командной строки
package config

import (
	"github.com/caarlos0/env/v10"
	flag "github.com/spf13/pflag"
)

// Options Опции для запуска сервера
var Options struct {
	// Адрес сервера
	Host string `env:"SERVER_ADDRESS" envDefault:":8080"`
	// Адрес хоста при формировании короткой ссылки
	ShortURLHost string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	// Путь к файловому хранилищу
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"/tmp/short-url-db.json"`
	// Строка подключения к базе данных
	DatabaseDSN string `env:"DATABASE_DSN"`
	// Секретный ключ для формирования UserID
	PassphraseKey string `env:"PASSPHRASE_KEY"`
	// Настройка использования SSL
	SSL struct {
		Enabled bool   `env:"ENABLE_HTTPS"`
		Key     string `env:"KEY_FILE" envDefault:"./shortener/certs/key.pem"`
		Cert    string `env:"CERT_FILE" envDefault:"./shortener/certs/cert.pem"`
	}
}

// ParseEnvFlags Обработка окружения и флагов для формирования конфигурации
func ParseEnvFlags() {
	if err := env.Parse(&Options); err != nil {
		panic("can't parse environment variables")
	}
	flag.BoolVarP(&Options.SSL.Enabled, "s", "s", Options.SSL.Enabled, "Использовать HTTPS")
	flag.StringVarP(&Options.PassphraseKey, "k", "k", Options.PassphraseKey, "Пароль для ключа")
	flag.StringVarP(&Options.DatabaseDSN, "d", "d", Options.DatabaseDSN, "Строка подключения к БД")
	flag.StringVarP(&Options.FileStoragePath, "f", "f", Options.FileStoragePath, "Путь к файлу для сохраненных ссылок")
	flag.StringVarP(&Options.Host, "a", "a", Options.Host, "Адрес HTTP-сервера")
	flag.StringVarP(&Options.ShortURLHost, "b", "b", Options.ShortURLHost, "URL короткой ссылки")
	flag.Parse()
}
