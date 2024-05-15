package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	svc "github.com/PaulOh5/mync/cmd/grpc-service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

type grpcConfig struct {
	service string
	method  string
	request string
	url     string
}

func HandleGrpc(w io.Writer, args []string) error {
	c := grpcConfig{}
	fs := flag.NewFlagSet("grpc", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.StringVar(&c.service, "service", "", "Service of gRPC")
	fs.StringVar(&c.method, "method", "", "Method to call")
	fs.StringVar(&c.request, "request", "", "Request for gRPC")

	fs.Usage = func() {
		var usageString = `

grpc: A gRPC client.

grpc: <options> server`
		fmt.Fprint(w, usageString)
		fmt.Fprintln(w)
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options: ")
		fs.PrintDefaults()
	}

	err := fs.Parse(args)
	if err != nil {
		return err
	}

	if fs.NArg() != 1 {
		return ErrorNoServerSpecified
	}

	c.url = fs.Arg(0)
	result, err := sendGRPCRequest(c)
	if err != nil {
		return err
	}
	fmt.Fprintln(w, result)

	return nil
}

func sendGRPCRequest(config grpcConfig) (string, error) {
	conn, err := setupGrpcConnection(config.url)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	switch config.service {
	case "Users":
		client := getUserServiceClient(conn)
		request, err := createUserRequest(config.request)
		if err != nil {
			return "", err
		}
		result, err := getUser(client, request)
		if err != nil {
			return "", err
		}
		return getUserResponseJson(result)
	case "Repos":
		client := getRepoServiceClient(conn)
		request, err := createRepoRequest(config.request)
		if err != nil {
			return "", err
		}
		result, err := getRepo(client, request)
		if err != nil {
			return "", err
		}
		return getRepoResponseJson(result)
	default:
		return "", errors.New("invalid grpc service")
	}
}

func setupGrpcConnection(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

func getUserServiceClient(conn *grpc.ClientConn) svc.UsersClient {
	return svc.NewUsersClient(conn)
}

func getRepoServiceClient(conn *grpc.ClientConn) svc.RepoClient {
	return svc.NewRepoClient(conn)
}

func getUser(
	client svc.UsersClient,
	u *svc.UserGetRequest,
) (*svc.UserGetReply, error) {
	return client.GetUser(context.Background(), u)
}

func getRepo(
	client svc.RepoClient,
	r *svc.RepoGetRequest,
) (*svc.RepoGetReply, error) {
	return client.GetRepos(context.Background(), r)
}

func createUserRequest(jsonQuery string) (*svc.UserGetRequest, error) {
	u := svc.UserGetRequest{}
	input := []byte(jsonQuery)
	return &u, protojson.Unmarshal(input, &u)
}

func createRepoRequest(jsonQuery string) (*svc.RepoGetRequest, error) {
	r := svc.RepoGetRequest{}
	input := []byte(jsonQuery)
	return &r, protojson.Unmarshal(input, &r)
}

func getUserResponseJson(result *svc.UserGetReply) (string, error) {
	data, err := protojson.Marshal(result)
	return string(data), err
}

func getRepoResponseJson(result *svc.RepoGetReply) (string, error) {
	data, err := protojson.Marshal(result)
	return string(data), err
}
