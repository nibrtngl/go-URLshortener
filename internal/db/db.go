package db

import (
	"flag"
	"github.com/jackc/pgx/v4"
	"os"
)

var db *pgx.Conn

func GetDBConnectionParams() (string, string, string, string, string) {
	// Параметры подключения к базе данных
	var (
		host     string
		port     string
		user     string
		password string
		dbName   string
	)

	// Проверяем флаги
	flag.StringVar(&host, "dbhost", "", "database host")
	flag.StringVar(&port, "dbport", "", "database port")
	flag.StringVar(&user, "dbuser", "", "database user")
	flag.StringVar(&password, "dbpassword", "", "database password")
	flag.StringVar(&dbName, "dbname", "", "database name")

	flag.Parse()

	// Если флаги не установлены, проверяем переменные окружения
	if host == "" {
		host = os.Getenv("DB_HOST")
	}
	if port == "" {
		port = os.Getenv("DB_PORT")
	}
	if user == "" {
		user = os.Getenv("DB_USER")
	}
	if password == "" {
		password = os.Getenv("DB_PASSWORD")
	}
	if dbName == "" {
		dbName = os.Getenv("DB_NAME")
	}

	// Установка значений по умолчанию, если они не были установлены через флаги или переменные окружения
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	if user == "" {
		user = "postgres"
	}
	if dbName == "" {
		dbName = "url-db"
	}

	return host, port, user, password, dbName
}
