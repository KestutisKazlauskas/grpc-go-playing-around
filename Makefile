grpc-compile-greet:
	protoc greet/greetpb/greet.proto --go_out=plugins=grpc:.

grpc-compile-calculator:
	protoc calculator/calculatorpb/calculator.proto --go_out=plugins=grpc:.

grpc-compile-blog:
	protoc blog/blogpb/blog.proto --go_out=plugins=grpc:.

run-server-calculator:
	${GOROOT}/bin/go run calculator/calculator_server/server.go

run-client-calculator:
	${GOROOT}/bin/go run calculator/calculator_client/client.go

run-server-greet:
	${GOROOT}/bin/go run greet/greet_server/server.go

run-client-greet:
	${GOROOT}/bin/go run greet/greet_client/client.go

run-server-blog:
	${GOROOT}/bin/go run blog/blog_server/server.go

run-client-blog:
	${GOROOT}/bin/go run blog/blog_client/client.go

run-mongo:
	mongo admin -u root -p change_this