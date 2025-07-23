package agent

import (
	"context"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"github.com/stewyb314/remote-control/internal/db"
	pb "github.com/stewyb314/remote-control/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Agent struct {
	pb.UnimplementedAgentServer
	log *logrus.Logger
	addr string
	port int
	tlsCredentials credentials.TransportCredentials
	db db.DB
}

func New(log *logrus.Logger, addr string, port int, tlsCredentials credentials.TransportCredentials, db db.DB) *Agent {
	return &Agent{
		log: log,
		addr: addr,
		port: port,
		tlsCredentials: tlsCredentials,	
		db: db,
	}
}
func (a *Agent) StartAgent()  error {
	s := grpc.NewServer(grpc.Creds(a.tlsCredentials))
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", a.addr, 50051))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	pb.RegisterAgentServer(s, a)
	a.log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
	return nil
}

func (a *Agent) Start(ctx context.Context, in *pb.StartRequest) (*pb.StartResponse, error) {
	a.log.Infof("Received Start request: %+v", in)
	return &pb.StartResponse{Id: "0x12345"}, nil
}
