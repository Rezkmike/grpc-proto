package main

import (
    "context"
    "log"
    "time"

    "google.golang.org/grpc"
    pb "example.com/mypackage/client/generated"
)

func main() {
    conn, err := grpc.Dial("server:50051", grpc.WithInsecure())
    if err != nil {
        log.Fatalf("did not connect: %v", err)
    }
    defer conn.Close()

    c := pb.NewPersonServiceClient(conn)

    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    r, err := c.GetPerson(ctx, &pb.PersonRequest{Id: 1})
    if err != nil {
        log.Fatalf("could not get person: %v", err)
    }
    log.Printf("Person: %s", r.GetPerson().GetName())
}

