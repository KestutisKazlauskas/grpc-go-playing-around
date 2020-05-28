package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"

	"github.com/KestutisKazlauskas/grpc-go/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct{}

func (s *server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	fmt.Printf("Calculator was called %v", req)

	res := &calculatorpb.SumResponse{
		Result: req.GetX() + req.GetY(),
	}

	return res, nil
}

func (s *server) PrimeNumberDecomposition(req *calculatorpb.PrimeNumberDecompositionRequest, stream calculatorpb.CalculatorService_PrimeNumberDecompositionServer) error {
	var primeNumber int32 = 2
	number := req.GetNumber()

	for number > 1 {
		if number%primeNumber == 0 {
			res := &calculatorpb.PrimeNumberDecompositionResponse{
				PrimeNubmer: primeNumber,
			}
			stream.Send(res)
			//send to stream
			number = number / primeNumber
		} else {
			primeNumber = primeNumber + 1
		}
	}
	return nil
}

func (s *server) Average(stream calculatorpb.CalculatorService_AverageServer) error {

	var sum int32
	var length int32
	var avg float64

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Printf("Length %d", length)
			log.Printf("Sum: %v", sum)

			avg = float64(sum) / float64(length)
			return stream.SendAndClose(&calculatorpb.AverageResponse{Avg: avg})
		}

		if err != nil {
			log.Fatalf("Error on streaming calculator: %v", err)
		}
		log.Printf("Recievied %d", req.GetNumber())
		sum += req.GetNumber()
		length++
	}

}

func (s *server) Max(stream calculatorpb.CalculatorService_MaxServer) error {
	log.Println("Max stream was called")

	var max int32 = 0
	var isFirstRequest bool = true

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("Max stream erro on request streaming %v", err)
		}

		currentNumber := req.GetNumber()

		if isFirstRequest || max < currentNumber {
			max = currentNumber
			sendErr := stream.Send(&calculatorpb.MaxResponse{CurrentMax: max})
			if sendErr != nil {
				log.Fatalf("Error on sending response to Max stream: %v", sendErr)
			}
		}

		if isFirstRequest {
			isFirstRequest = false
		}

	}

	return nil
}

func (s *server) SquareRoot(ctx context.Context, req *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error) {
	number := req.GetNumber()

	if number < 0 {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Recieved negative number %v", number),
		)
	}
	return &calculatorpb.SquareRootResponse{
		NumberRoot: math.Sqrt(float64(number)),
	}, nil
}

func main() {
	fmt.Println("Startin server")

	listen, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}

	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	if err := s.Serve(listen); err != nil {
		log.Fatalf("Failder to server %v", err)
	}
}
