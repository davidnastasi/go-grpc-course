package main

import (
	"fmt"
	"github.com/davidnastasi/go-grpc-course/calculator/calculatorpb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"log"
	"math"
	"net"
)

type CalculatorServer struct {
	calculatorpb.CalculatorServiceServer
}

func (server *CalculatorServer) SquareRoot(ctx context.Context, req *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error) {
	fmt.Println("receive SquareRoot RPC")
	number := req.GetNumber()
	if number < 0 {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("receive a negative number: %v ", number))
	}

	return &calculatorpb.SquareRootResponse{
		NumberRoot: math.Sqrt(float64(number)),
	}, nil
}

func main() {

	log.Println("Starting calculator application")

	l, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err.Error())
	}
	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s, &CalculatorServer{})

	reflection.Register(s)

	if err = s.Serve(l); err != nil {
		log.Fatalf("failed to server: %v", err.Error())
	}

}
