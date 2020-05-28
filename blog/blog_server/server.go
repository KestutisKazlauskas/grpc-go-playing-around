package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/KestutisKazlauskas/grpc-go/blog/blogpb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var collection *mongo.Collection

type blogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Content  string             `bson:"content"`
	Title    string             `bson:"title"`
}

func (item *blogItem) toBlogbp() *blogpb.Blog {
	return &blogpb.Blog{
		Id:       item.ID.Hex(),
		AuthorId: item.AuthorID,
		Title:    item.Title,
		Content:  item.Content,
	}
}

type server struct {
}

func (s *server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	blog := req.GetBlog()

	data := blogItem{
		AuthorID: blog.GetAuthorId(),
		Title:    blog.GetTitle(),
		Content:  blog.GetContent(),
	}

	res, err := collection.InsertOne(context.Background(), data)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}

	objid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Can not convert ot objectId: %v", err),
		)
	}

	return &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       objid.Hex(),
			AuthorId: blog.GetAuthorId(),
			Title:    blog.GetTitle(),
			Content:  blog.GetContent(),
		},
	}, nil
}

func (s *server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	blogID := req.GetBlogId()

	objid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID %v", err),
		)
	}

	data := &blogItem{}
	filter := bson.D{primitive.E{Key: "_id", Value: objid}}

	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog with id %v", err),
		)
	}

	return &blogpb.ReadBlogResponse{
		Blog: data.toBlogbp(),
	}, nil
}

func (s *server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	blog := req.GetBlog()

	objid, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID %v", err),
		)
	}

	data := &blogItem{}
	filter := bson.D{primitive.E{Key: "_id", Value: objid}}

	doc := collection.FindOne(context.Background(), filter)
	if err := doc.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog with id %v", err),
		)
	}

	data.AuthorID = blog.GetAuthorId()
	data.Content = blog.GetContent()
	data.Title = blog.GetTitle()

	_, err = collection.ReplaceOne(context.Background(), filter, data)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot update blog %v", err),
		)
	}

	return &blogpb.UpdateBlogResponse{
		Blog: data.toBlogbp(),
	}, nil

}

func (s *server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	blogID := req.GetBlogId()

	objid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID %v", err),
		)
	}

	filter := bson.D{primitive.E{Key: "_id", Value: objid}}

	res, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error on deleting document %v", err),
		)
	}
	if res.DeletedCount <= 0 {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Document not found %s", blogID),
		)
	}

	return &blogpb.DeleteBlogResponse{
		BlogId: blogID,
	}, nil

}

func (s *server) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {

	cur, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error on reading a blog list %v", err),
		)
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		blog := &blogItem{}
		err := cur.Decode(blog)
		if err != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Error decoding blogItem %v", err),
			)
		}

		stream.Send(&blogpb.ListBlogResponse{
			Blog: blog.toBlogbp(),
		})
	}
	if err := cur.Err(); err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Unknown error %v", err),
		)
	}

	return nil
}

func main() {
	// if somthing crash in go code
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Connecting mongodb")
	client, mongoErr := mongo.NewClient(options.Client().ApplyURI("mongodb://root:change_this@localhost:27017"))
	if mongoErr != nil {
		log.Fatalf("Mongodb error %v", mongoErr)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	mongoErr = client.Connect(ctx)
	if mongoErr != nil {
		log.Fatalf("Mongodb error %v", mongoErr)
	}

	collection = client.Database("blog").Collection("blog")

	fmt.Println("Blog service started")
	listen, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}

	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	blogpb.RegisterBlogServiceServer(s, &server{})

	// Register reflection service on gRPC server.
	//Install  evans and using cli for reflection
	reflection.Register(s)

	go func() {
		fmt.Println("Starting server...")
		if err := s.Serve(listen); err != nil {
			log.Fatalf("Failder to server %v", err)
		}
	}()

	// Waiting for ctr + c to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	//Block until signal is recieved
	//Properly close everything
	<-ch
	fmt.Println("Stoping the server")
	s.Stop()
	fmt.Println("Closing the listener")
	listen.Close()
	fmt.Println("Closing mongodb")
	client.Disconnect(ctx)
	fmt.Println("Program ended")
}
