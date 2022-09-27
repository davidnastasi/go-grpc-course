package main

import (
	"fmt"
	"github.com/davidnastasi/go-grpc-course/greet/greetpb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"time"
)

func main() {

	certFile := "ssl/ca-cert.pem"
	creds, err := credentials.NewClientTLSFromFile(certFile, "")
	if err != nil {
		log.Fatalf("error while loading CA trust certificate: %v", err)
		return
	}
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("fail to dial: %v ", err.Error())
	}

	defer conn.Close()

	var c = greetpb.NewGreetServiceClient(conn)
	doUnary(c)
	//doServerStreaming(c)
	//doClientStreaming(c)
	//doBiDirectionalStreaming(c)
	//doUnaryWithDeadline(c, 5*time.Second)
}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Println("Staring doing Unary RPC")

	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "David",
			LastName:  "Nastasi",
		}}

	res, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling Greet RPC: %v", err.Error())
	}

	fmt.Printf("Response from greet: %v ", res.Result)
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Staring doing Stream RPC")
	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "David",
			LastName:  "Nastasi",
		},
	}
	resSteam, err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling GreetManyTimes RCP: %v", err)
	}

	for {
		msg, err := resSteam.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error while calling GreetManyTimes RCP: %v", err)
		}
		log.Printf("response from GreetManyTimes: %v", msg.GetResult())
	}

}

func doClientStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Staring doing client Streaming RPC")

	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("error while calling LongGreet RCP: %v", err)
	}

	requests := []*greetpb.LongGreetRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "David",
				LastName:  "Nastasi",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Franco",
				LastName:  "Nastasi",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Victoria",
				LastName:  "Nastasi",
			},
		},
	}

	for _, request := range requests {
		fmt.Printf("Sending req: %v\n", request)
		err = stream.Send(request)
		if err != nil {
			log.Fatalf("error while sending LongRequest RCP: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error while closing LongRequest RCP: %v", err)
	}

	fmt.Printf("LongGreet response: %v\n", res)

}

func doBiDirectionalStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Staring to do BiDi Streaming RPC")
	requests := []*greetpb.GreetEveryoneRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "David",
				LastName:  "Nastasi",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Franco",
				LastName:  "Nastasi",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Victoria",
				LastName:  "Nastasi",
			},
		},
	}

	// we create a stream by invoking the clien
	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("error while creating stream %v", err)
		return
	}

	waitc := make(chan struct{})

	// we send a bunch of messages to the client
	go func() {
		for _, request := range requests {
			fmt.Printf("sending message: %v\n", request)
			errSend := stream.Send(request)
			if errSend != nil {
				close(waitc)
				log.Fatalf("error while sending message: %v", errSend)
				return
			}
			time.Sleep(1 * time.Second)
		}
		errClose := stream.CloseSend()
		if errClose != nil {
			close(waitc)
			log.Fatalf("error while closing message: %v", errClose)
			return
		}
	}()

	// we receive a bunch of messages to the client
	go func() {
		for {
			resp, errRecv := stream.Recv()
			if errRecv == io.EOF {
				close(waitc)
				return
			}
			if errRecv != nil {
				close(waitc)
				log.Fatalf("error while receiving message: %v", errRecv)
				return
			}
			fmt.Printf("received: %v\n", resp.GetResult())
		}
	}()

	// block until everything is done
	<-waitc
}

func doUnaryWithDeadline(c greetpb.GreetServiceClient, duration time.Duration) {
	fmt.Println("Staring doing Unary with deadline RPC")

	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "David",
			LastName:  "Nastasi",
		}}

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	res, err := c.GreetWithDeadline(ctx, req)
	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				fmt.Println("timeout was hit! Deadline was exceeded")
			} else {
				fmt.Printf("unexpected error: %v", statusErr)
			}
			return
		} else {
			log.Fatalf("error while calling Greet RPC: %v", err.Error())

		}
	}

	fmt.Printf("Response from greet: %v ", res.Result)
}
