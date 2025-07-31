package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/stewyb314/remote-control/internal/db"
	"github.com/stewyb314/remote-control/internal/services"
	pb "github.com/stewyb314/remote-control/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Agent struct {
	pb.UnimplementedAgentServer
	log *logrus.Entry
	addr string
	port int
	tlsCredentials credentials.TransportCredentials
	jobs *services.Jobs
	db db.DB
}

func New(log *logrus.Entry, addr string, port int, tlsCredentials credentials.TransportCredentials, db db.DB, jobs *services.Jobs) *Agent {
	return &Agent{
		log: log,
		addr: addr,
		port: port,
		tlsCredentials: tlsCredentials,	
		db: db,
		jobs: jobs,
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
	id, err := a.jobs.NewJob(in.Command, in.Args)
	if err != nil {
		return nil, fmt.Errorf("failed to create new job: %v", err)
	}
	return &pb.StartResponse{Id: id}, nil
}
func (a *Agent) Status(ctx context.Context, in *pb.StatusRequest) (*pb.StatusResponse, error) {
	a.log.Infof("Received Status request for job ID: %s", in.Id)
	exec, err := a.db.GetExecution(in.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution for job ID %s: %v", in.Id, err)
	}
	if exec == nil {
		return nil, fmt.Errorf("no execution found for job ID %s", in.Id)
	}
	var argsB []byte
	err  = exec.Args.UnmarshalJSON(argsB)
	if err != nil {
		a.log.Error("failed to unmarshal args for job ID %s: %v", in.Id, err.Error())
	}
	var args []string
	err = json.Unmarshal(argsB, &args)
	if err != nil {
		a.log.Errorf("failed to unmarshal args for job ID %s: %v", in.Id, err)
	}


	return &pb.StatusResponse{
		Id: in.Id,
		Cmd: exec.Command,
		Exit: exec.ExitCode,
		State: pb.State(exec.Status),
		Args: args,
	}, nil
}

func (a *Agent) Stop(ctx context.Context, in *pb.StopRequest) (*pb.StopResponse, error) {
	a.log.Infof("Received Stop request for job ID: %s", in.Id)
	err := a.jobs.StopJob(in.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to stop job with ID %s: %v", in.Id, err)
	}
	return &pb.StopResponse{Id: in.Id}, nil
}

func (a *Agent) Output(in *pb.OutputRequest, serv pb.Agent_OutputServer) error {
	a.log.Infof("Received Output request for job ID: %s", in.Id)
	exec, err := a.db.GetExecution(in.Id)
	if err != nil {
		return  fmt.Errorf("failed to get execution for job ID %s: %v", in.Id, err)
	}
	if exec == nil {
		return  fmt.Errorf("no execution found for job ID %s", in.Id)
	}
	file, err := os.Open(exec.Output)
	if err != nil {
		return  fmt.Errorf("failed to open output file for job ID %s: %v", in.Id, err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Bytes()
		serv.Send(&pb.OutputResponse{
			Output: line,
		})
	}
	return nil
}
