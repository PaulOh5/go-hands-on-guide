package main

import (
	"context"
	"log"
	"net"

	svc "github.com/PaulOh5/multiple-services/service"
	"google.golang.org/grpc"
)

type userService struct {
	svc.UnimplementedUsersServer
}

type repoService struct {
	svc.UnimplementedRepoServer
}

func (s *repoService) GetRepos(
	ctx context.Context,
	in *svc.RepoGetRequest,
) (*svc.RepoGetReply, error) {
	log.Printf(
		"Recevied request for repo with CreatId: %s Id: %s\n",
		in.CreatorId,
		in.Id,
	)
	repo := svc.Repository{
		Id:    in.Id,
		Name:  "test repo",
		Url:   "https://git.example.com/test/repo",
		Owner: &svc.User{Id: in.CreatorId, FirstName: "Paul"},
	}
	r := svc.RepoGetReply{Repo: []*svc.Repository{&repo}}
	return &r, nil
}

func registerServices(s *grpc.Server) {
	svc.RegisterRepoServer(s, &repoService{})
	svc.RegisterUsersServer(s, &userService{})
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
