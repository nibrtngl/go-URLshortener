package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
)

var urlMapping map[string]string

func main() {
	urlMapping = make(map[string]string)

	http.HandleFunc("/", shortenURLHandler)
	http.ListenAndServe(":8080", nil)
}

func shortenURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusTemporaryRedirect)
			return
		}

		originalURL := string(body)
		id := generateShortID()
		urlMapping[id] = originalURL

		shortURL := fmt.Sprintf("http://localhost:8080/%s", id)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, shortURL)
	} else if r.Method == http.MethodGet {
		id := r.URL.Path[1:]
		originalURL, exists := urlMapping[id]
		if !exists {
			http.Error(w, "Not Found", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, originalURL)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func generateShortID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	idLength := 8
	b := make([]byte, idLength)

	// Генерация уникального идентификатора
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}
