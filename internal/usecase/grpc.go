package usecase

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/denmor86/go-url-shortener/internal/config"
	pb "github.com/denmor86/go-url-shortener/internal/gen"
	"github.com/denmor86/go-url-shortener/internal/storage"
	"github.com/denmor86/go-url-shortener/internal/workerpool"
)

// UsecaseGRPC - модель основной бизнес логики для GRPC
type UsecaseGRPC struct {
	pb.UnimplementedShortenerServer
	use *Usecase // основная логика
}

// NewUsecaseGRPC - метод создания объекта бизнес логики для GRPC запросов
func NewUsecaseGRPC(cfg *config.Config, storage storage.IStorage, workerpool *workerpool.WorkerPool) *UsecaseGRPC {
	return &UsecaseGRPC{use: &Usecase{Config: cfg, Storage: storage, WorkerPool: workerpool}}
}

// DecodeURL - метод получения оригинального URL по короткой ссылке на основе proto запроса
func (u *UsecaseGRPC) DecodeURL(ctx context.Context, in *pb.DecodeURLRequest) (*pb.DecodeURLResponse, error) {
	if len(in.GetUrl()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid url")
	}

	URL, err := u.use.DecodeURL(ctx, in.GetUrl())
	if err != nil {
		return nil, status.Error(codes.Unknown, "error decode url")
	}

	response := &pb.DecodeURLResponse{
		Result: URL,
	}

	return response, nil
}

// EncodeURL - метод формирования короткой ссылки на основе proto запроса
func (u *UsecaseGRPC) EncodeURL(ctx context.Context, in *pb.EncodeURLRequest) (*pb.EncodeURLResponse, error) {
	if len(in.GetUrl()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid url")
	}

	shortURL, err := u.use.EncodeURL(ctx, in.GetUrl(), in.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Unknown, "error encode url")
	}

	response := &pb.EncodeURLResponse{
		Result: shortURL,
	}

	return response, nil
}

// EncodeURLs - метод формирования массива коротких ссылок на основе proto запроса
func (u *UsecaseGRPC) EncodeURLs(ctx context.Context, in *pb.EncodeURLsRequest) (*pb.EncodeURLsResponse, error) {
	if len(in.GetUrls()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid urls")
	}
	requestItems := make([]RequestItem, 0, len(in.GetUrls()))
	responseItems, err := u.use.EncodeURLBatch(ctx, requestItems, in.GetUserId())

	if err != nil {
		return nil, err
	}

	results := make([]*pb.ShortURL, 0, len(responseItems))
	for _, url := range responseItems {
		results = append(results, &pb.ShortURL{
			Url: url.URL,
			Id:  url.ID,
		})
	}

	response := &pb.EncodeURLsResponse{
		Results: results,
	}

	return response, nil
}

// GetURLs - метод получения информации об имеющихся записях URL на основе proto запроса
func (u *UsecaseGRPC) GetURLs(ctx context.Context, in *pb.GetURLsRequest) (*pb.GetURLsResponse, error) {
	responseItems, err := u.use.GetURLs(ctx, in.GetUserId())

	if err != nil {
		return nil, err
	}

	results := make([]*pb.URL, 0, len(responseItems))
	for _, url := range responseItems {
		results = append(results, &pb.URL{
			Shorten:  url.ShortURL,
			Original: url.OriginalURL,
		})
	}

	response := &pb.GetURLsResponse{
		Results: results,
	}

	return response, nil
}

// DeleteURLs - метод запроса на удаление информации об имеющихся записях URL по пользователю
func (u *UsecaseGRPC) DeleteURLs(ctx context.Context, in *pb.DeleteURLsRequest) (*pb.DeleteURLsResponse, error) {
	u.use.DeleteURLs(ctx, in.GetUrls(), in.GetUserId())
	response := &pb.DeleteURLsResponse{
		Urls: in.GetUrls(),
	}
	return response, nil
}

// GetStatistic - метод получения статистики об имеющихся записях URL и пользователях на основе proto запроса
func (u *UsecaseGRPC) GetStatistic(ctx context.Context, in *pb.StatisticRequest) (*pb.StatisticResponse, error) {
	URLs, Users := u.use.GetStatistic(ctx)
	response := &pb.StatisticResponse{
		Urls:  int32(URLs),
		Users: int32(Users),
	}
	return response, nil
}
