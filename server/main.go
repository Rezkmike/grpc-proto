package main

import (
    "context"
    "log"
    "net"

    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
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
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    s := grpc.NewServer()
    pb.RegisterPersonServiceServer(s, &server{})

    // Register reflection service on gRPC server
    reflection.Register(s)

    log.Println("Server is running on port :50051")
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}

