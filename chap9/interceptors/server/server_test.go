package main

import (
	"context"
	"log"
	"net"
	"testing"

	users "github.com/PaulOh5/interceptors/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func startTestGrpcServer() (*grpc.Server, *bufconn.Listener) {
	l := bufconn.Listen(10)
	s := grpc.NewServer()
	registerServices(s)
	go func() {
		err := startServer(s, l)
		if err != nil {
			log.Fatal(err)
		}
	}()
	return s, l
}

func TestUserService(t *testing.T) {
	s, l := startTestGrpcServer()
	defer s.GracefulStop()

	bufconnDialer := func(
		ctx context.Context, addr string,
	) (net.Conn, error) {
		return l.Dial()
	}

	client, err := grpc.DialContext(
		context.Background(),
		"",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(bufconnDialer),
	)
	if err != nil {
		t.Fatal(err)
	}
	userClient := users.NewUsersClient(client)
	resp, err := userClient.GetUser(
		context.Background(),
		&users.UserGetRequest{
			Email: "paul@baroai.com",
			Id:    "JiHwan",
		},
	)

	if err != nil {
		t.Fatal(err)
	}
	if resp.User.FirstName != "paul" {
		t.Errorf(
			"Expected FirstName to be : paul, Got: %s",
			resp.User.FirstName,
		)
	}
}
