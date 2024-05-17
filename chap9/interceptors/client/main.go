package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	svc "github.com/PaulOh5/interceptors/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type wrappedClientStream struct {
	grpc.ClientStream
}

func (s wrappedClientStream) SendMsg(m interface{}) error {
	log.Printf("Send msg called: %T", m)
	return s.ClientStream.SendMsg(m)
}

func (s wrappedClientStream) RecvMsg(m interface{}) error {
	log.Printf("Recv msg called: %T", m)
	return s.ClientStream.RecvMsg(m)
}

func (s wrappedClientStream) CloseSend() error {
	log.Println("CloseSend() called")
	return s.ClientStream.CloseSend()
}

func metadataUnaryInterceptor(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	ctxWithMetadata := metadata.AppendToOutgoingContext(
		ctx, "Request-Id", "request-123",
	)

	return invoker(
		ctxWithMetadata,
		method,
		req,
		reply,
		cc,
		opts...,
	)
}

func metadataStreamInterceptor(
	ctx context.Context,
	desc *grpc.StreamDesc,
	cc *grpc.ClientConn,
	method string,
	streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	ctxWithMetadata := metadata.AppendToOutgoingContext(
		ctx,
		"Request-Id",
		"request-123",
	)
	stream, err := streamer(ctxWithMetadata, desc, cc, method, opts...)
	clientStream := wrappedClientStream{ClientStream: stream}
	return clientStream, err
}

func setupGrpcConnection(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(metadataUnaryInterceptor),
		grpc.WithStreamInterceptor(metadataStreamInterceptor),
	)
}

func getUserServiceClient(conn *grpc.ClientConn) svc.UsersClient {
	return svc.NewUsersClient(conn)
}

func getUser(client svc.UsersClient, u *svc.UserGetRequest) (*svc.UserGetReply, error) {
	return client.GetUser(context.Background(), u)
}

func setupChat(r io.Reader, w io.Writer, c svc.UsersClient) error {
	stream, err := c.GetHelp(context.Background())
	if err != nil {
		return err
	}

	for {
		scanner := bufio.NewScanner(r)
		prompt := "Request: "
		fmt.Fprint(w, prompt)

		scanner.Scan()
		if err := scanner.Err(); err != nil {
			return err
		}

		msg := scanner.Text()
		if msg == "quit" {
			break
		}

		request := svc.UserHelpRequest{Request: msg}
		err := stream.Send(&request)
		if err != nil {
			return err
		}

		resp, err := stream.Recv()
		if err != nil {
			return err
		}

		fmt.Printf("Response: %s\n", resp.Response)
	}

	return stream.CloseSend()
}

// func createUserRequest(jsonQuery string) (*svc.UserGetRequest, error) {
// 	u := svc.UserGetRequest{}
// 	input := []byte(jsonQuery)
// 	return &u, protojson.Unmarshal(input, &u)
// }

// func getUserResponseJson(result *svc.UserGetReply) ([]byte, error) {
// 	return protojson.Marshal(result)
// }

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Must specify a gRPC server address and search query")
	}
	serverAddr := os.Args[1]
	methodName := os.Args[2]

	conn, err := setupGrpcConnection(serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	c := getUserServiceClient(conn)

	switch methodName {
	case "GetUser":
		result, err := getUser(c, &svc.UserGetRequest{Email: "paul@baroai.com"})
		s := status.Convert(err)
		if s.Code() != codes.OK {
			log.Fatalf("Request failed: %v - %v\n", s.Code(), s.Message())
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(
			os.Stdout, "User: %s %s\n",
			result.User.FirstName,
			result.User.LastName,
		)
	case "GetHelp":
		if err = setupChat(os.Stdin, os.Stdout, c); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("Unrecognized method name")
	}
}
