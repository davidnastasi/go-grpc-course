package main

import (
	"fmt"
	"github.com/davidnastasi/go-grpc-course/blog/blogpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
)

func main() {

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("fail to dial: %v ", err.Error())
	}

	defer conn.Close()

	var c = blogpb.NewBlogServiceClient(conn)
	doUnary(c)
}

func doUnary(c blogpb.BlogServiceClient) {
	blog := &blogpb.Blog{
		AuthorId: "David",
		Title:    "My first blog",
		Content:  "Content of the first blog",
	}
	createBlog, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{Blog: blog})
	if err != nil {
		log.Fatalf("unexpected error: %v", err)
	}
	fmt.Printf("blog has been created: %v\n", createBlog)

	_, err = c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: primitive.NewObjectID().Hex()})
	if err != nil {
		log.Printf("unexpected error: %v\n", err)
	}

	readBlog, err := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: createBlog.GetBlog().GetId()})
	if err != nil {
		log.Fatalf("unexpected error: %v\n", err)
	}
	fmt.Printf("blog has been read: %v\n", readBlog)

	updateBlog, err := c.UpdateBlog(context.Background(), &blogpb.UpdateBlogRequest{Blog: &blogpb.Blog{
		Id:       createBlog.GetBlog().GetId(),
		AuthorId: "Franco Nastasi",
		Title:    "my first blog (edited)",
		Content:  "Content of the first blog, with some awesome additions!",
	}})
	if err != nil {
		log.Fatalf("unexpected error: %v\n", err)
	}
	fmt.Printf("blog has been updated: %v\n", updateBlog)

	deleteBlog, err := c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{BlogId: createBlog.GetBlog().GetId()})
	if err != nil {
		log.Fatalf("unexpected error: %v\n", err)
	}

	fmt.Printf("blog has been deleted: %v\n", deleteBlog)

	fmt.Println("list of blogs")

	listBlog, err := c.ListBlog(context.Background(), &blogpb.ListBlogRequest{})
	if err != nil {
		log.Fatalf("unexpected error: %v\n", err)
	}

	for {
		stream, err := listBlog.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("unexpected error: %v\n", err)
		}
		fmt.Println(stream.GetBlog())
	}

}
