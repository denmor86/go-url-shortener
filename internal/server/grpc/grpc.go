package server

import (
	"google.golang.org/grpc"

	pb "github.com/denmor86/go-url-shortener/internal/gen"
	"github.com/denmor86/go-url-shortener/internal/usecase"
)

// NewServer - метод создаёт новый GRPC сервер
func NewServer(use *usecase.UsecaseGRPC) *grpc.Server {

	s := grpc.NewServer()
	pb.RegisterShortenerServer(s, use)
	return s
}
