package grpc

import (
	"context"
	"errors"
	"fiber-apis/internal/models"
	"fiber-apis/internal/server"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ShortenerServer struct {
	UnimplementedURLShortenerServer // Автоматически сгенерированный интерфейс

	storage       server.Storable
	cfg           models.Config
	chanForDelete chan []string
	server        *server.Server // Добавляем поле для структуры Server
}

// Метод для создания короткого URL
func (s *ShortenerServer) CreateShortURL(ctx context.Context, in *CreateShortURLRequest) (*CreateShortURLResponse, error) {
	res := &CreateShortURLResponse{}

	// Получение userID из контекста
	userID := getUser(ctx)

	// Проверка валидности URL
	if !server.IsValidURL(in.OriginalUrl) {
		return nil, errors.New("invalid URL")
	}

	// Генерация короткого ID
	shortID := server.GenerateShortID()

	// Сохранение URL в хранилище
	dbID, err := s.storage.SetURL(shortID, in.OriginalUrl, userID)
	if err != nil {
		logrus.Errorf("failed to save URL: %v", err)
		return nil, err
	}

	// Формирование короткой ссылки
	shortURL := s.cfg.BaseURL + dbID
	res.ShortUrl = shortURL

	// Если сгенерированный ID уже существует
	if dbID != shortID {
		return nil, errors.New("conflict: URL already exists")
	}

	// Сохранение данных в файл через объект Server
	if err = s.server.SaveStorageToFile(s.cfg.FileStoragePath); err != nil {
		logrus.Errorf("failed to save storage to file: %v", err)
		return nil, err
	}

	return res, nil
}

// Метод для получения оригинального URL
func (s *ShortenerServer) GetOriginalURL(ctx context.Context, in *GetOriginalURLRequest) (*GetOriginalURLResponse, error) {
	res := &GetOriginalURLResponse{}

	// Получение userID из контекста
	userID := getUser(ctx)

	// Извлечение URL из хранилища
	urlData, err := s.storage.GetURL(in.ShortUrl, userID)
	if err != nil {
		return nil, errors.New("URL not found")
	}

	// Если URL помечен как удаленный
	if urlData.IsDeleted {
		return nil, errors.New("URL is deleted")
	}

	// Запись оригинального URL в ответ
	res.OriginalUrl = urlData.OriginalURL
	return res, nil
}

// Метод для получения всех URL пользователя
func (s *ShortenerServer) GetUserURLs(ctx context.Context, in *GetUserURLsRequest) (*GetUserURLsResponse, error) {
	res := &GetUserURLsResponse{}

	// Получение userID из контекста
	userID := getUser(ctx)

	// Извлечение всех URL пользователя из хранилища
	urls, err := s.storage.GetUserURLs(userID)
	if err != nil {
		return nil, errors.New("failed to get user URLs")
	}

	// Формирование ответа
	for _, url := range urls {
		res.Urls = append(res.Urls, &URL{
			ShortUrl:    s.cfg.BaseURL + url.ShortURL,
			OriginalUrl: url.OriginalURL,
		})
	}

	return res, nil
}

// Interceptor для добавления/проверки userID в метаданных gRPC
func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var user string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("user")
		if len(values) > 0 {
			user = values[0]
		}
	}

	if len(user) == 0 {
		md := metadata.New(map[string]string{"user": uuid.New().String()})
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	return handler(ctx, req)
}

// Функция для получения userID из контекста
func getUser(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if values := md.Get("user"); len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

// GetGRPCServer создает и возвращает gRPC сервер
func GetGRPCServer(cfg models.Config, ch4delete chan []string, store server.Storable, srv *server.Server) (*grpc.Server, error) {
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(unaryInterceptor))

	server := &ShortenerServer{
		storage:       store,
		cfg:           cfg,
		chanForDelete: ch4delete,
		server:        srv, // Передаем экземпляр структуры Server
	}

	RegisterURLShortenerServer(grpcServer, server)

	return grpcServer, nil
}
