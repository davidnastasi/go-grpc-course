package main

import (
	"fmt"
	"github.com/davidnastasi/go-grpc-course/calculator/calculatorpb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"log"
)

func main() {

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("fail to dial: %v ", err.Error())
	}

	defer conn.Close()

	var c = calculatorpb.NewCalculatorServiceClient(conn)
	doErrorUnary(c)
}

func doErrorUnary(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("staring to do a SquareRoot Unary RPC...")
	number := int32(-10)
	res, err := c.SquareRoot(context.Background(), &calculatorpb.SquareRootRequest{Number: number})
	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			// actual error from gRPC (user error)
			fmt.Println(respErr.Message())
			fmt.Println(respErr.Code())
			if respErr.Code() == codes.InvalidArgument {
				fmt.Println("we probably sent a negative number")
			}
			return

		} else {
			log.Fatalf("severe error calling SquareRoot: %v", respErr)
		}
	}
	fmt.Printf("result of square root of %v: %v\n", number, res.GetNumberRoot())
}
