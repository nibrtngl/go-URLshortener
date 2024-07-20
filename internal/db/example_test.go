package db_test

import (
	"context"
	"fiber-apis/internal/db"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Пример использования функции NewDatabaseStorage.
func ExampleNewDatabaseStorage() {
	pool, _ := pgxpool.Connect(context.Background(), "your_connection_string")
	storage := db.NewDatabaseStorage(pool)
	fmt.Println(storage)
	// Output: <ваш ожидаемый вывод>
}

// Пример использования функции GetURL.
func ExampleDatabaseStorage_GetURL() {
	pool, _ := pgxpool.Connect(context.Background(), "your_connection_string")
	storage := db.NewDatabaseStorage(pool)
	url, err := storage.GetURL("shortURL", "userID")
	fmt.Println(url, err)
	// Output: <ваш ожидаемый вывод>
}

// Пример использования функции SetURL.
func ExampleDatabaseStorage_SetURL() {
	pool, _ := pgxpool.Connect(context.Background(), "your_connection_string")
	storage := db.NewDatabaseStorage(pool)
	id, err := storage.SetURL("id", "url", "userID")
	fmt.Println(id, err)
	// Output: <ваш ожидаемый вывод>
}

// Пример использования функции SetURLsAsDeleted.
func ExampleDatabaseStorage_SetURLsAsDeleted() {
	pool, _ := pgxpool.Connect(context.Background(), "your_connection_string")
	storage := db.NewDatabaseStorage(pool)
	err := storage.SetURLsAsDeleted([]string{"id1", "id2"}, "userID")
	fmt.Println(err)
	// Output: <ваш ожидаемый вывод>
}

// Пример использования функции GetUserURLs.
func ExampleDatabaseStorage_GetUserURLs() {
	pool, _ := pgxpool.Connect(context.Background(), "your_connection_string")
	storage := db.NewDatabaseStorage(pool)
	urls, err := storage.GetUserURLs("userID")
	fmt.Println(urls, err)
	// Output: <ваш ожидаемый вывод>
}

// Пример использования функции Ping.
func ExampleDatabaseStorage_Ping() {
	pool, _ := pgxpool.Connect(context.Background(), "your_connection_string")
	storage := db.NewDatabaseStorage(pool)
	err := storage.Ping()
	fmt.Println(err)
	// Output: <ваш ожидаемый вывод>
}
