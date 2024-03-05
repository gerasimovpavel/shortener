// Package config реализует управоение настройками через чтение переменных окружения
// или параметро командной строки
package config

import (
	"encoding/json"
	"github.com/caarlos0/env/v10"
	"github.com/gerasimovpavel/shortener.git/pkg/logger"
	flag "github.com/spf13/pflag"
	"go.uber.org/zap"
	"net"
	"os"
)

// Options Опции для запуска сервера
type Options struct {
	// Адрес сервера
	Host string `json:"server_address" env:"SERVER_ADDRESS" envDefault:":8080"`
	// Адрес хоста при формировании короткой ссылки
	ShortURLHost string `json:"base_url" env:"BASE_URL" envDefault:"http://localhost:8080"`
	// Путь к файловому хранилищу
	FileStoragePath string `json:"file_storage_path" env:"FILE_STORAGE_PATH" envDefault:"/tmp/short-url-db.json"`
	// Строка подключения к базе данных
	DatabaseDSN string `json:"database_dsn" env:"DATABASE_DSN"`
	// Секретный ключ для формирования UserID
	PassphraseKey string `env:"PASSPHRASE_KEY"`
	// Настройка использования SSL
	SSLEnabled bool `json:"enable_https" env:"ENABLE_HTTPS"`
	// Ключ SSL
	SSLKey string `env:"KEY_FILE" envDefault:"./shortener/certs/key.pem"`
	// Сертифкат SSL
	SSLCert string `env:"CERT_FILE" envDefault:"./shortener/certs/cert.pem"`
	// Путь ка файлу конфигурации
	JSONConfig string `env:"CONFIG"`
	//Доверенная подсеть
	TrustedSubNet string `env:"TRUSTED_SUBNET"`
}

// Переменная, содержащая настройки программы
var Cfg Options

// ParseEnvFlags Обработка окружения и флагов для формирования конфигурации
func ParseEnvFlags() {
	if err := env.Parse(&Cfg); err != nil {
		panic("can't parse environment variables")
	}
	if !flag.CommandLine.Parsed() {
		flag.BoolVarP(&Cfg.SSLEnabled, "s", "s", Cfg.SSLEnabled, "Использовать HTTPS")
		flag.StringVarP(&Cfg.PassphraseKey, "k", "k", Cfg.PassphraseKey, "Пароль для ключа")
		flag.StringVarP(&Cfg.DatabaseDSN, "d", "d", Cfg.DatabaseDSN, "Строка подключения к БД")
		flag.StringVarP(&Cfg.FileStoragePath, "f", "f", Cfg.FileStoragePath, "Путь к файлу для сохраненных ссылок")
		flag.StringVarP(&Cfg.Host, "a", "a", Cfg.Host, "Адрес HTTP-сервера")
		flag.StringVarP(&Cfg.ShortURLHost, "b", "b", Cfg.ShortURLHost, "URL короткой ссылки")
		flag.StringVarP(&Cfg.ShortURLHost, "t", "t", Cfg.TrustedSubNet, "URL короткой ссылки")
		flag.StringVarP(&Cfg.JSONConfig, "config", "c", Cfg.JSONConfig, "Файл конфигурации")
		flag.Parse()
	}
	if Cfg.JSONConfig != "" {
		if err := parseJSONConfig(Cfg.JSONConfig); err != nil {
			logger.Logger.Error("can't parse json config", zap.Error(err))
		}
	}
	if Cfg.TrustedSubNet != "" {
		_, _, err := net.ParseCIDR(Cfg.TrustedSubNet)
		if err != nil {
			logger.Logger.Error("wrong trusted subnet mask")
		}
	}
}

func parseJSONConfig(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var cfg Options
	if err := json.Unmarshal(b, &cfg); err != nil {
		return err
	}

	if Cfg.FileStoragePath == "" {
		Cfg.FileStoragePath = cfg.FileStoragePath
	}

	if Cfg.Host == "" {
		Cfg.Host = cfg.Host
	}

	if Cfg.DatabaseDSN == "" {
		Cfg.DatabaseDSN = cfg.DatabaseDSN
	}

	if Cfg.ShortURLHost == "" {
		Cfg.ShortURLHost = cfg.ShortURLHost
	}

	if !Cfg.SSLEnabled {
		Cfg.SSLEnabled = cfg.SSLEnabled
	}

	if Cfg.SSLCert == "" {
		Cfg.SSLCert = cfg.SSLCert
	}

	if Cfg.SSLKey == "" {
		Cfg.SSLKey = cfg.SSLKey
	}

	if Cfg.TrustedSubNet == "" {
		Cfg.TrustedSubNet = cfg.TrustedSubNet
	}
	return nil
}
