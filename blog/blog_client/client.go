package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/KestutisKazlauskas/grpc-go/blog/blogpb"
	"google.golang.org/grpc"
)

func main() {

	fmt.Println("Client started")

	//connection without ssl
	opts := grpc.WithInsecure() //if want to use without tsl
	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("Error on initiatning connection %v", err)
	}
	defer conn.Close()

	c := blogpb.NewBlogServiceClient(conn)

	fmt.Println("Creating a blog.")
	blog := &blogpb.Blog{
		AuthorId: "AuthorString",
		Title:    "My first blog in Go",
		Content:  "A lot of good struf",
	}

	createBlogRes, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{Blog: blog})
	if err != nil {
		log.Fatalf("Error on creating blog: %v", err)
	}
	fmt.Printf("Blog has been created with a response: %v", createBlogRes)

	//read blog
	fmt.Println("Reading the blog")

	_, readErr := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: "5ecd425c6a09822c9aa1781e"})

	if readErr != nil {
		fmt.Printf("Error while reading: %v\n", readErr)
	}

	blogID := createBlogRes.GetBlog().GetId()
	readBlogReq := &blogpb.ReadBlogRequest{BlogId: blogID}
	readRes, readErr := c.ReadBlog(context.Background(), readBlogReq)

	if readErr != nil {
		fmt.Printf("Error while reading: %v\n", readErr)
	}

	fmt.Printf("Read blog response: %v\n", readRes)

	//Update the blog
	updatedBlog := &blogpb.Blog{
		Id:       blogID,
		AuthorId: "AuthorString the better",
		Title:    "My first blog in Go this change too",
		Content:  "A lot of good struf.Last but not least",
	}

	updateBlogReq := &blogpb.UpdateBlogRequest{Blog: updatedBlog}
	updateRes, updateErr := c.UpdateBlog(context.Background(), updateBlogReq)
	if updateErr != nil {
		fmt.Printf("Error on update blog: %v\n", updateErr)
	}

	fmt.Printf("Blog updated: %v\n", updateRes)

	//Delete blog
	deleteBlogReq := &blogpb.DeleteBlogRequest{BlogId: blogID}
	deleteRes, deleteErr := c.DeleteBlog(context.Background(), deleteBlogReq)
	if deleteErr != nil {
		log.Fatalf("Delete err %v\n", deleteErr)
	}

	fmt.Printf("Delete success %v\n", deleteRes)

	//List all the blogs
	stream, err := c.ListBlog(context.Background(), &blogpb.ListBlogRequest{})
	if err != nil {
		log.Fatalf("Error on streaming %v", err)
	}

	for {
		blog, err := stream.Recv()
		if err == io.EOF {
			//Finished streaming
			break
		}

		if err != nil {
			log.Fatalf("Error while streaming blog %v\n", err)
		}

		fmt.Printf("Blog list item streamed: %v \n", blog)
	}
}
