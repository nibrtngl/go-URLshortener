package server_test

import (
	"fiber-apis/internal/models"
	"fiber-apis/internal/server"
	"log"
)

func ExampleServer_PingHandler() {
	cfg := models.Config{
		Address: "localhost:8080",
		BaseURL: "http://localhost:8080",
	}

	s := server.NewServer(cfg, nil, nil)
	s.App.Get("/ping", s.PingHandler)

	log.Fatal(s.App.Listen(s.Cfg.Address))
}

func ExampleServer_shortenURLHandler() {
	cfg := models.Config{
		Address: "localhost:8080",
		BaseURL: "http://localhost:8080",
	}

	s := server.NewServer(cfg, nil, nil)
	s.App.Post("/", s.ShortenURLHandler)

	log.Fatal(s.App.Listen(s.Cfg.Address))
}

func ExampleServer_redirectToOriginalURL() {
	cfg := models.Config{
		Address: "localhost:8080",
		BaseURL: "http://localhost:8080",
	}

	s := server.NewServer(cfg, nil, nil)
	s.App.Get("/:id", s.RedirectToOriginalURL)

	log.Fatal(s.App.Listen(s.Cfg.Address))
}

func ExampleServer_shortenAPIHandler() {
	cfg := models.Config{
		Address: "localhost:8080",
		BaseURL: "http://localhost:8080",
	}

	s := server.NewServer(cfg, nil, nil)
	s.App.Post("/api/shorten", s.ShortenAPIHandler)

	log.Fatal(s.App.Listen(s.Cfg.Address))
}

func ExampleServer_getUserURLsHandler() {
	cfg := models.Config{
		Address: "localhost:8080",
		BaseURL: "http://localhost:8080",
	}

	s := server.NewServer(cfg, nil, nil)
	s.App.Get("/api/user/urls", s.GetUserURLsHandler)

	log.Fatal(s.App.Listen(s.Cfg.Address))
}

func ExampleServer_deleteURLsHandler() {
	cfg := models.Config{
		Address: "localhost:8080",
		BaseURL: "http://localhost:8080",
	}

	s := server.NewServer(cfg, nil, nil)
	s.App.Delete("/api/user/urls", s.DeleteURLsHandler)

	log.Fatal(s.App.Listen(s.Cfg.Address))
}

func ExampleServer_shortenBatchURLHandler() {
	cfg := models.Config{
		Address: "localhost:8080",
		BaseURL: "http://localhost:8080",
	}

	s := server.NewServer(cfg, nil, nil)
	s.App.Post("/api/shorten/batch", s.ShortenBatchURLHandler)

	log.Fatal(s.App.Listen(s.Cfg.Address))
}
