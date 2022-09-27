package main

import (
	"fmt"
	"github.com/davidnastasi/go-grpc-course/blog/blogpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	_ "go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"os"
	"os/signal"
)

var collection *mongo.Collection

type blogItem struct {
	ID       primitive.ObjectID `bson:"_id"`
	AuthorID string             `bson:"author_id"`
	Content  string             `bson:"content"`
	Title    string             `bson:"title"`
}

type BlogServer struct {
	blogpb.BlogServiceServer
}

func (s *BlogServer) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	blog := req.GetBlog()
	data := blogItem{
		ID:       primitive.NewObjectID(),
		AuthorID: blog.GetAuthorId(),
		Title:    blog.GetTitle(),
		Content:  blog.GetContent(),
	}

	res, err := collection.InsertOne(
		context.Background(),
		data,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("internal error: %v", err))
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal, "cannot convert OID")
	}

	return &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       oid.Hex(),
			AuthorId: blog.GetAuthorId(),
			Title:    blog.GetTitle(),
			Content:  blog.GetContent(),
		},
	}, nil

}

func (s *BlogServer) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	blogID := req.GetBlogId()

	oid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("can not parse id"))
	}

	data := &blogItem{}
	filter := bson.D{{"_id", oid}}
	res := collection.FindOne(context.Background(), filter)
	err = res.Decode(data)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("internal error: %v", err))
	}

	return &blogpb.ReadBlogResponse{Blog: &blogpb.Blog{
		Id:       data.ID.Hex(),
		AuthorId: data.AuthorID,
		Title:    data.Title,
		Content:  data.Content,
	}}, nil

}

func (s *BlogServer) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	blog := req.GetBlog()
	oid, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("can not parse id"))
	}

	data := &blogItem{}
	filter := bson.D{{"_id", oid}}
	res := collection.FindOne(context.Background(), filter)

	err = res.Decode(data)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("internal error: %v", err))
	}

	data.AuthorID = blog.GetAuthorId()
	data.Content = blog.GetContent()
	data.Title = blog.GetTitle()

	_, err = collection.ReplaceOne(context.Background(), filter, data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("can not update object: %v", err))
	}

	return &blogpb.UpdateBlogResponse{Blog: dataToBlogPb(data)}, nil
}

func dataToBlogPb(data *blogItem) *blogpb.Blog {
	return &blogpb.Blog{
		Id:       data.ID.Hex(),
		AuthorId: data.AuthorID,
		Title:    data.Title,
		Content:  data.Content,
	}
}

func (s *BlogServer) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	blogID := req.GetBlogId()

	oid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("can not parse id"))
	}

	filter := bson.D{{"_id", oid}}
	res, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("cannot delete  object: %v", err))
	}
	if res.DeletedCount == 0 {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("cannot find the object id: %v", blogID))
	}

	return &blogpb.DeleteBlogResponse{BlogId: blogID}, nil

}
func (s *BlogServer) ListBlog(req *blogpb.ListBlogRequest, strem blogpb.BlogService_ListBlogServer) error {
	find, err := collection.Find(context.Background(), bson.D{{}})
	if err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("cannot list object: %v", err))
	}
	defer find.Close(context.Background())
	for find.Next(context.Background()) {
		data := &blogItem{}
		err = find.Decode(data)
		if err != nil {
			return status.Errorf(codes.Internal, fmt.Sprintf("cannot decode error %v", err))
		}
		err = strem.Send(&blogpb.ListBlogResponse{Blog: dataToBlogPb(data)})
		if err != nil {
			return status.Errorf(codes.Internal, fmt.Sprintf("cannot send data %v", err))
		}
	}

	return nil
}

func main() {
	// if wwe crash the go code, we get the file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting blog service")

	fmt.Println("Connecting to mongo db")
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	collection = client.Database("mydb").Collection("blog")

	fmt.Println("Starting listener")
	l, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err.Error())
	}
	s := grpc.NewServer()
	blogpb.RegisterBlogServiceServer(s, &BlogServer{})
	reflection.Register(s)
	go func() {
		if err = s.Serve(l); err != nil {
			log.Fatalf("failed to server: %v", err.Error())
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	// block until a signal is received
	<-ch
	fmt.Println("Stopping the server")
	s.Stop()
	fmt.Println("Close the listener")
	l.Close()
	fmt.Println("Close the client")
	client.Disconnect(context.TODO())
	fmt.Println("End program ")
}
