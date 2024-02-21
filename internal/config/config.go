package config

import (
	flag "github.com/spf13/pflag"
	"os"
	"path/filepath"
	"strings"
)

// Опции для запуска сервера
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
	// Настройка использования SSL
	SSL struct {
		Enabled bool
		Key     string
		Cert    string
	}
}

// ParseEnvFlags Обработка окружения и флагов для формирования конфигурации
func ParseEnvFlags() {
	var ok bool
	EnableHTTPS, ok := os.LookupEnv("ENABLE_HTTPS")
	Options.SSL.Enabled = strings.EqualFold(EnableHTTPS, "true")
	if !ok {
		flag.BoolVarP(&Options.SSL.Enabled, "s", "s", false, "Использовать HTTPS")
		Options.SSL.Enabled = flag.CommandLine.ShorthandLookup("s") != nil
	}
	wd, _ := filepath.Abs(".")

	Options.SSL.Key, ok = os.LookupEnv("KEY_FILE")
	if !ok {
		Options.SSL.Key = wd + "/shortener/certs/key.pem"
	}
	Options.SSL.Cert, ok = os.LookupEnv("CERT_FILE")
	if !ok {
		Options.SSL.Cert = wd + "/shortener/certs/cert.pem"
	}

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
	flag.CommandLine.ShorthandLookup("s")
}
