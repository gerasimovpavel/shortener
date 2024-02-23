package config

import (
	"os"
	"strconv"
	"testing"
)

func TestParseJSONConfig(t *testing.T) {
	tests := []struct {
		json      string
		wantError bool
	}{
		{
			json: `{
						"server_address": "localhost:8080", 
						"base_url": "http://localhost", 
						"file_storage_path": "/path/to/file.db", 
						"database_dsn": "", 
						"enable_https": true 
					}`,
			wantError: false,
		},
		{
			json: `{
						"server_address": localhost:8080", 
						"base_url": "http://localhost",
						"file_storage_path": "/path/to/file.db", 
						"database_dsn": "", 
						"enable_https": true 
					}`,
			wantError: true,
		},
	}
	for idx, tt := range tests {
		t.Run("test parser json config "+strconv.Itoa(idx), func(t *testing.T) {
			d := []byte(tt.json)
			f, err := os.Create("config.json")
			if err != nil {
				panic(err)
			}
			defer f.Close()
			defer os.Remove("config.json")
			_, err = f.Write(d)
			if err != nil {
				panic(err)
			}

			err = parseJSONConfig("config.json")
			if err != nil && !tt.wantError {
				panic(err)
			}
		})
	}
}
