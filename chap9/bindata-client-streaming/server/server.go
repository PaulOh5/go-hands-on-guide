package main

import (
	"fmt"
	"io"
	"log"
	"net"

	svc "github.com/PaulOh5/bindata-client-streaming/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type repoService struct {
	svc.UnimplementedRepoServer
}

func (s *repoService) CreateRepo(stream svc.Repo_CreateRepoServer) error {
	var repoContext *svc.RepoContext
	var data []byte

	for {
		r, err := stream.Recv()
		if err == io.EOF {
			break
		}
		switch t := r.Body.(type) {
		case *svc.RepoCreateRequest_Context:
			repoContext = r.GetContext()
		case *svc.RepoCreateRequest_Data:
			b := r.GetData()
			data = append(data, b...)
		case nil:
			return status.Errorf(
				codes.FailedPrecondition,
				"Unexpected message type: %s",
				t,
			)
		default:
			return status.Errorf(
				codes.FailedPrecondition,
				"Unexpected message type: %s",
				t,
			)
		}
	}

	repo := svc.Repository{
		Name: repoContext.Name,
		Url: fmt.Sprintf(
			"https://git.example.com/%s/%s",
			repoContext.CreatorId, repoContext.Name,
		),
	}
	r := svc.RepoCreateReply{
		Repo: &repo,
		Size: int32(len(data)),
	}
	return stream.SendAndClose(&r)
}

func registerServices(s *grpc.Server) {
	svc.RegisterRepoServer(s, &repoService{})
}

func startServer(s *grpc.Server, l net.Listener) error {
	return s.Serve(l)
}

func main() {
	listenAddr := ":50051"
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	registerServices(s)
	log.Fatal(startServer(s, lis))
}
