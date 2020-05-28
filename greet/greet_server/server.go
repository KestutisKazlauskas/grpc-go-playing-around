package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/KestutisKazlauskas/grpc-go/greet/greetpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct{}

func (s *server) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	fmt.Printf("Greet was called %v", req)
	firstName := req.GetGreeting().GetFirstName()

	result := "Helo, " + firstName

	res := &greetpb.GreetResponse{
		Result: result,
	}

	return res, nil
}

func (s *server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	fmt.Printf("GreetMany times was called %v", req)
	firstName := req.GetGreeting().GetFirstName()
	for i := 0; i < 10; i++ {
		result := "Hello " + firstName + " " + strconv.Itoa(i) + "time"
		res := &greetpb.GreetManyTimesResponse{
			Result: result,
		}
		stream.Send(res)
		time.Sleep(1000 * time.Millisecond)
	}

	return nil
}

func (s *server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {
	fmt.Printf("LongGreet was called")

	result := "Hello, "
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			//Send and close
			return stream.SendAndClose(&greetpb.LongGreetResponse{
				Result: result,
			})
		}

		if err != nil {
			log.Fatalf("Failed to recieve stream: %v", err)
		}

		firstName := req.GetGreeting().GetFirstName()
		result += firstName + "! "
	}
}

func (s *server) GreetEveryOne(stream greetpb.GreetService_GreetEveryOneServer) error {
	fmt.Printf("GreetEveryOne streem was called")

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			log.Fatalf("Error reading client stream %v", err)
		}

		firstName := req.GetGreeting().GetFirstName()
		result := "Hello, " + firstName + "!"
		sendErr := stream.Send(&greetpb.GreetEveryOneResponse{
			Result: result,
		})

		if sendErr != nil {
			log.Fatalf("Error on sendint dat ato client! %v", err)
		}
	}
}

func (s *server) GreetWithDeadline(ctx context.Context, req *greetpb.GreetWithDeadlineRequest) (*greetpb.GreetWithDeadlineResponse, error) {
	fmt.Printf("GreetWithDeadline was called %v", req)
	for i := 0; i < 3; i++ {
		if ctx.Err() == context.Canceled {
			//the client cancel
			fmt.Println("Clent cancel the  excecution")
			return nil, status.Error(codes.Canceled, "the client canceled")
		}
		time.Sleep(1 * time.Second)
	}
	firstName := req.GetGreeting().GetFirstName()

	result := "Helo, " + firstName

	res := &greetpb.GreetWithDeadlineResponse{
		Result: result,
	}

	return res, nil
}

func main() {
	fmt.Println("Startin server")

	listen, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}

	certFile := "ssl/server.crt"
	keyFile := "ssl/server.pem"
	creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)
	if sslErr != nil {
		log.Fatalf("Failed to loading sertificate %v", sslErr)
	}
	opts := []grpc.ServerOption{}
	// coment append  out if do not want to use tls
	opts = append(opts, grpc.Creds(creds))
	s := grpc.NewServer(opts...)
	greetpb.RegisterGreetServiceServer(s, &server{})

	// Register reflection service on gRPC server.
	//Install  evans and using cli for reflection
	reflection.Register(s)

	if err := s.Serve(listen); err != nil {
		log.Fatalf("Failder to server %v", err)
	}
}
