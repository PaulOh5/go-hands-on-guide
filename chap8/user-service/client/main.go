package main

import (
	"context"
	"fmt"
	"log"
	"os"

	users "github.com/PaulOh5/user-service/service-v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

func setupGrpcConnection(addr string) (*grpc.ClientConn, error) {
	return grpc.DialContext(
		context.Background(),
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
}

func getUserServiceClient(conn *grpc.ClientConn) users.UsersClient {
	return users.NewUsersClient(conn)
}

func getUser(client users.UsersClient, u *users.UserGetRequest) (*users.UserGetReply, error) {
	return client.GetUser(context.Background(), u)
}

func createUserRequest(jsonQuery string) (*users.UserGetRequest, error) {
	u := users.UserGetRequest{}
	input := []byte(jsonQuery)
	return &u, protojson.Unmarshal(input, &u)
}

func getUserResponseJson(result *users.UserGetReply) ([]byte, error) {
	return protojson.Marshal(result)
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Must specify a gRPC server address and search query")
	}
	serverAddr := os.Args[1]
	u, err := createUserRequest(os.Args[2])
	if err != nil {
		log.Fatal("Bad user input: &v", err)
	}

	conn, err := setupGrpcConnection(serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	c := getUserServiceClient(conn)

	result, err := getUser(c, u)
	s := status.Convert(err)
	if s.Code() != codes.OK {
		log.Fatalf("Request failed: %v - %v\n", s.Code(), s.Message())
	}
	if err != nil {
		log.Fatal(err)
	}
	data, err := getUserResponseJson(result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(
		os.Stdout, string(data),
	)
}
