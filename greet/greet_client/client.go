package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/KestutisKazlauskas/grpc-go/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func main() {

	fmt.Println("Client started")

	//connection without ssl
	creds, sslErr := credentials.NewClientTLSFromFile("ssl/ca.crt", "")
	if sslErr != nil {
		log.Fatalf("Error geting crediantilsa from Authorety trust certificate Certificate %v", sslErr)
	}

	opts := grpc.WithTransportCredentials(creds)
	// opts := grpc.WithInsecure() if want to use without tsl
	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("Error on initiatning connection %v", err)
	}
	defer conn.Close()

	cc := greetpb.NewGreetServiceClient(conn)

	doUnary(cc)
	//doServerStreaming(cc)
	//doClientStream(cc)
	//doBiDiStream(cc)
	//doGreatWithDeadline(cc, 5*time.Second)
	//doGreatWithDeadline(cc, 1*time.Second)
}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Println("Starting a request")
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Name",
			LastName:  "Name",
		},
	}
	res, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("Error on response %v", err)
	}

	log.Printf("Response: %v", res.Result)
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Server streaming api...")
	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Stream me",
			LastName:  "Close me",
		},
	}
	resStream, err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("Error on streaming %v", err)
	}

	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			//Finished streaming
			break
		}

		if err != nil {
			log.Fatalf("Error on streaming message %v", err)
		}

		log.Printf("Response of stream: %v", msg.GetResult())
	}
}

func doClientStream(c greetpb.GreetServiceClient) {
	fmt.Println("Client streaming api...")

	requests := []*greetpb.LongGreetRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "First",
				LastName:  "Last",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Second",
				LastName:  "SecondLast",
			},
		},
	}

	stream, err := c.LongGreet(context.Background())

	if err != nil {
		log.Fatalf("Error on LongGreet client %v", err)
	}

	for _, req := range requests {
		stream.Send(req)
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error recieving response %v", err)
	}

	fmt.Printf("Respnse: %v", response.GetResult())
}

func doBiDiStream(c greetpb.GreetServiceClient) {
	fmt.Println("doBiDiStream streaming api...")

	// Create a stream by incoking client
	stream, err := c.GreetEveryOne(context.Background())
	if err != nil {
		log.Fatalf("Error on doBi client %v", err)
	}

	requests := []*greetpb.GreetEveryOneRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "First",
				LastName:  "Last",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Second",
				LastName:  "SecondLast",
			},
		},
	}

	//For teh code blocking with wait channel
	waitc := make(chan struct{})
	// Send message to the clients
	go func() {
		//Send a lot of message
		for _, req := range requests {
			fmt.Printf("Sending the request %v\n", req)
			// Can handle the error here also!
			stream.Send(req)
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()
	// Recieved messages
	go func() {
		for {
			//recieve a bunch of messages
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while recieving %v", err)
				break
			}

			fmt.Printf("Recieved:%v\n", res.GetResult())
		}
		close(waitc)
	}()

	//block until its done
	<-waitc
}

func doGreatWithDeadline(c greetpb.GreetServiceClient, duration time.Duration) {
	fmt.Println("Starting a request with deadline")
	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Name",
			LastName:  "Name",
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	res, err := c.GreetWithDeadline(ctx, req)
	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				fmt.Println("Deadline was exceed")
			} else {
				fmt.Printf("Unexpected error! %v", statusErr)
			}
		} else {
			log.Fatalf("Error on response %v", err)
		}

		return
	}

	log.Printf("Response: %v", res.Result)

}
