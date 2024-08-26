package config

import (
	"encoding/json"
	"fiber-apis/internal/models"
	"io"
	"os"
)

// LoadConfigFromFile загружает конфигурацию из JSON-файла.
func LoadConfigFromFile(filePath string) (*models.Config, error) {
	var cfg models.Config

	// Открываем файл конфигурации
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Читаем содержимое файл
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Парсим JSON в структуру Config
	err = json.Unmarshal(bytes, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
