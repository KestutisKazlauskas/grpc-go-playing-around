package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/KestutisKazlauskas/grpc-go/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {

	fmt.Println("Client started")

	//connection without ssl
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error on initiatning connection %v", err)
	}
	defer conn.Close()

	cc := calculatorpb.NewCalculatorServiceClient(conn)

	//doUnary(cc)
	//doStreaming(cc)
	//doClientStreaming(cc)
	//doBiDiStreaming(cc)
	doSqrtErrorHandling(cc)
}

func doUnary(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting a request")
	req := &calculatorpb.SumRequest{
		X: 9,
		Y: 3,
	}
	res, err := c.Sum(context.Background(), req)
	if err != nil {
		log.Fatalf("Error on response %v", err)
	}

	log.Printf("Response: %v", res.Result)
}

func doStreaming(c calculatorpb.CalculatorServiceClient) {
	var number int32 = 120
	log.Printf("Sending thenubmer %d", number)

	req := &calculatorpb.PrimeNumberDecompositionRequest{Number: number}

	primeStream, err := c.PrimeNumberDecomposition(context.Background(), req)
	if err != nil {
		log.Fatalf("Error on prime number streaming %v", err)
	}

	for {
		primeNumber, err := primeStream.Recv()
		if err == io.EOF {
			log.Println("Done streaming")
			break
		}

		if err != nil {
			log.Fatalf("Error on straming prime numbers %v", err)
		}

		log.Printf("Next prime number result %d", primeNumber.GetPrimeNubmer())
	}
}

func doClientStreaming(c calculatorpb.CalculatorServiceClient) {
	numbers := []int32{1, 2, 3, 4}

	stream, err := c.Average(context.Background())
	if err != nil {
		log.Fatalf("Error on streaming average: %v", err)
	}

	for _, number := range numbers {
		log.Printf("Sending: %d", number)
		stream.Send(&calculatorpb.AverageRequest{Number: number})
	}

	response, err := stream.CloseAndRecv()

	if err != nil {
		log.Fatalf("Error on recieving avg response: %v", err)
	}

	log.Printf("Avg response is: %f", response.GetAvg())
}

func doBiDiStreaming(c calculatorpb.CalculatorServiceClient) {

	fmt.Println("doBiDiStream Max streaming api...")

	numbers := []int32{1, 5, 3, 6, 2, 20}

	//make a clientstream
	stream, err := c.Max(context.Background())
	if err != nil {
		log.Fatalf("Error on creating Max stream %v", err)
	}

	waitc := make(chan int)
	//Sending request in one go routing
	go func() {

		for _, number := range numbers {
			fmt.Printf("Sending the request %v\n", number)
			stream.Send(&calculatorpb.MaxRequest{Number: number})
			// Just for fun
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	//Recieving response in another go routin
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error on recieving a response form Max: %v", err)
			}
			fmt.Printf("Recieved:%v\n", res.GetCurrentMax())
		}
		close(waitc)
	}()

	//for blocking code execution. Waits until recievs or close channel
	<-waitc
}

func doSqrtErrorHandling(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Error call sqrt")

	//correct call
	var number int32 = 64
	doSqrtCall(c, number)

	//bad on
	var negativeNumber int32 = -64
	doSqrtCall(c, negativeNumber)
}

func doSqrtCall(c calculatorpb.CalculatorServiceClient, number int32) {
	res, err := c.SquareRoot(context.Background(), &calculatorpb.SquareRootRequest{Number: number})

	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			//this is a user errror
			fmt.Println(respErr.Message())
			fmt.Println(respErr.Code())
			if respErr.Code() == codes.InvalidArgument {
				fmt.Println("Send negative number!")
				return
			}
		} else {
			//Unexptexted error
			log.Fatalf("Big error bad: %v", err)
			return
		}
	}
	fmt.Printf("Result of %v is %v\n", number, res.GetNumberRoot())
}
