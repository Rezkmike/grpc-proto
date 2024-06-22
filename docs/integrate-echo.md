The Echo framework is a high-performance, extensible, and minimalist Go web framework. While it primarily handles HTTP RESTful APIs, we can integrate it with a gRPC service to serve both gRPC and HTTP endpoints.

### Step-by-Step Guide to Integrate Echo with gRPC

1. **Ensure Your Directory Structure**:
   Make sure your directory structure looks like this:

   ```
   grpc-protobuf-poc/
   ├── proto/
   │   └── example.proto
   ├── server/
   │   ├── main.go
   │   ├── go.mod
   │   └── generated/
   │       ├── example.pb.go
   │       └── example_grpc.pb.go
   ├── client/
   │   ├── main.go
   │   ├── go.mod
   │   └── generated/
   │       ├── example.pb.go
   │       └── example_grpc.pb.go
   ├── Dockerfile.server
   ├── Dockerfile.client
   └── docker-compose.yml
   ```

2. **Define the `.proto` File**:
   Ensure your `proto/example.proto` file is defined as follows:

   ```proto
   syntax = "proto3";

   package example;

   message Person {
       string name = 1;
       int32 id = 2;
       string email = 3;
   }

   message PersonRequest {
       int32 id = 1;
   }

   message PersonResponse {
       Person person = 1;
   }

   service PersonService {
       rpc GetPerson (PersonRequest) returns (PersonResponse);
   }
   ```

3. **Generate Protobuf Code**:
   Generate the Go code from your `.proto` file:

   ```sh
   docker run --rm -v $(pwd):/workspace -w /workspace rvolosatovs/protoc \
     -I/workspace/proto --go_out=paths=source_relative:/workspace/server/generated --go-grpc_out=paths=source_relative:/workspace/server/generated \
     /workspace/proto/example.proto
   ```

4. **Server `go.mod`**:
   Place this file in the `server` directory:

   ```go
   module example.com/mypackage/server

   go 1.20

   require (
       github.com/labstack/echo/v4 v4.7.0
       google.golang.org/grpc v1.64.0
       google.golang.org/protobuf v1.34.2
   )
   ```

5. **Server Main Code**: `server/main.go`

   ```go
   package main

   import (
       "context"
       "log"
       "net"
       "net/http"

       "google.golang.org/grpc"
       "google.golang.org/grpc/reflection"

       "github.com/labstack/echo/v4"
       "github.com/labstack/echo/v4/middleware"

       pb "example.com/mypackage/server/generated"
   )

   type server struct {
       pb.UnimplementedPersonServiceServer
   }

   func (s *server) GetPerson(ctx context.Context, req *pb.PersonRequest) (*pb.PersonResponse, error) {
       person := &pb.Person{
           Name:  "John Doe",
           Id:    req.GetId(),
           Email: "johndoe@example.com",
       }
       return &pb.PersonResponse{Person: person}, nil
   }

   func main() {
       // Start gRPC server
       go func() {
           lis, err := net.Listen("tcp", ":50051")
           if err != nil {
               log.Fatalf("failed to listen: %v", err)
           }

           s := grpc.NewServer()
           pb.RegisterPersonServiceServer(s, &server{})
           reflection.Register(s)

           log.Println("gRPC server is running on port :50051")
           if err := s.Serve(lis); err != nil {
               log.Fatalf("failed to serve: %v", err)
           }
       }()

       // Start Echo HTTP server
       e := echo.New()

       e.Use(middleware.Logger())
       e.Use(middleware.Recover())

       e.GET("/person/:id", func(c echo.Context) error {
           id := c.Param("id")
           personID, err := strconv.Atoi(id)
           if err != nil {
               return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
           }

           conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
           if err != nil {
               return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect to gRPC server"})
           }
           defer conn.Close()

           client := pb.NewPersonServiceClient(conn)
           res, err := client.GetPerson(context.Background(), &pb.PersonRequest{Id: int32(personID)})
           if err != nil {
               return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get person"})
           }

           return c.JSON(http.StatusOK, res.Person)
       })

       log.Println("HTTP server is running on port :8080")
       e.Logger.Fatal(e.Start(":8080"))
   }
   ```

6. **Dockerfile for Server**:

   ```Dockerfile
   FROM golang:1.20-alpine

   WORKDIR /app

   COPY server /app/server

   RUN apk add --no-cache git \
       && cd /app/server \
       && go mod tidy \
       && go build -o server .

   EXPOSE 50051
   EXPOSE 8080

   CMD ["/app/server/server"]
   ```

7. **Docker Compose Configuration**:

   Ensure your `docker-compose.yml` file is set up correctly to build and run the services.

   ```yaml
   version: '3.8'

   services:
     server:
       build:
         context: .
         dockerfile: Dockerfile.server
       ports:
         - "50051:50051"
         - "8080:8080"
   ```

### Running the PoC

1. **Generate Go Code**:
   Run the Docker command to generate the Go code from your `.proto` file and place it in the `generated` directory.

   ```sh
   docker run --rm -v $(pwd):/workspace -w /workspace rvolosatovs/protoc \
     -I/workspace/proto --go_out=paths=source_relative:/workspace/server/generated --go-grpc_out=paths=source_relative:/workspace/server/generated \
     /workspace/proto/example.proto
   ```

2. **Build and Run**:
   From the `grpc-protobuf-poc` directory, use Docker Compose to build and run the services.

   ```sh
   docker-compose up --build
   ```

### Testing the Setup

1. **Test gRPC Service**:
   Use `grpcurl` or a gRPC client to test the gRPC service.

   ```sh
   grpcurl -plaintext -d '{"id": 1}' localhost:50051 example.PersonService/GetPerson
   ```

   You should see a response similar to:

   ```json
   {
     "person": {
       "name": "John Doe",
       "id": 1,
       "email": "johndoe@example.com"
     }
   }
   ```

2. **Test HTTP Endpoint**:
   Use `curl` or a web browser to test the HTTP endpoint served by Echo.

   ```sh
   curl http://localhost:8080/person/1
   ```

   You should see a response similar to:

   ```json
   {
     "name": "John Doe",
     "id": 1,
     "email": "johndoe@example.com"
   }
   ```

### Summary

This setup demonstrates how to integrate gRPC with the Echo framework. The Echo framework serves HTTP requests and communicates with the gRPC server to retrieve data. This allows you to have both gRPC and RESTful HTTP endpoints in your application.
