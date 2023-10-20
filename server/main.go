// package main

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"net"
// 	"os"
// 	"os/signal"
// 	"sync"
// 	"time"

// 	"google.golang.org/grpc"

// 	pb "google.golang.org/grpc/examples/features/proto/echo"
// 	"google.golang.org/grpc/reflection"
// )

// var (
// 	addrs = []string{":50051", ":50052"}
// )

// type ecServer struct {
// 	pb.UnimplementedEchoServer
// 	addr string
// }

// func (s *ecServer) UnaryEcho(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
// 	return &pb.EchoResponse{Message: fmt.Sprintf("%s (from %s)", req.Message, s.addr)}, nil
// }

// func startServer(addr string) {
// 	lis, err := net.Listen("tcp", addr)
// 	if err != nil {
// 		log.Fatalf("failed to listen: %v", err)
// 	}
// 	s := grpc.NewServer()
// 	reflection.Register(s)
// 	pb.RegisterEchoServer(s, &ecServer{addr: addr})
// 	log.Printf("serving on %s\n", addr)
// 	if err := s.Serve(lis); err != nil {
// 		log.Fatalf("failed to serve: %v", err)
// 	}
// 	// Setting up a signal handler to catch Ctrl+C
//     c := make(chan os.Signal, 1)
//     signal.Notify(c, os.Interrupt)
//     <-c

//     // Gracefully shutting down the server
//     fmt.Println("Shutting down gRPC server...")

//     // First, stop accepting new connections
//     s.GracefulStop()

// 	// Wait a certain duration to finish processing ongoing requests
//     // This is optional and depends on your use case
//     // You can remove it if you don't have any ongoing requests to handle
//     time.Sleep(5 * time.Second)

//     fmt.Println("Server gracefully stopped")
// }

// func main() {
// 	var wg sync.WaitGroup
// 	for _, addr := range addrs {
// 		wg.Add(1)
// 		go func(addr string) {
// 			defer wg.Done()
// 			startServer(addr)
// 		}(addr)
// 	}

// 	wg.Wait()
// }

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"time"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/features/proto/echo"
)



var (
	addrs         = []string{":50051", ":50052"}
	serverMap     = make(map[string]*grpc.Server)
	serverMapLock sync.Mutex
)

type ecServer struct {
	pb.UnimplementedEchoServer
	addr string
}

func (s *ecServer) UnaryEcho(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	return &pb.EchoResponse{Message: fmt.Sprintf("%s (from %s)", req.Message, s.addr)}, nil
}

func startServer(addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	serverMapLock.Lock()
	serverMap[addr] = s
	serverMapLock.Unlock()

	pb.RegisterEchoServer(s, &ecServer{addr: addr})
	log.Printf("serving on %s\n", addr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func stopServer(addr string) {
	serverMapLock.Lock()
	if s, ok := serverMap[addr]; ok {
		delete(serverMap, addr)
		s.GracefulStop()
		log.Printf("Server at %s has been shut down\n", addr)
	} else {
		log.Printf("Server at %s not found\n", addr)
	}
	serverMapLock.Unlock()
}

func main() {
	var wg sync.WaitGroup
	for _, addr := range addrs {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			startServer(addr)
		}(addr)
	}

	// Example to shut down the server at ":50052"
	go func() {
		// Sleep for some time to ensure the server is up before shutting it down
		// Adjust this time as needed
		time.Sleep(10 * time.Second)
		stopServer(":50052")
	}()

	wg.Wait()
}
