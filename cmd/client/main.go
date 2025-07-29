package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	pb "github.com/stewyb314/remote-control/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)
type Parameters struct {
	Port   int
	Host   string
	Ident  string
	Help   bool
	SubCmd string
	Cmd    []string
}

type Connection struct {
	conn   *grpc.ClientConn
	Client pb.AgentClient
	Cancel context.CancelFunc
	Ctx    context.Context
}


func main() {
    // Define flags
	params := argParse()
	fmt.Printf("args: %v\n", params.SubCmd)
	conn, err := NewConnection(params)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	switch params.SubCmd {
	case "start":
		doStart(conn, params)
	case "status":
		doStatus(conn, params)
	case "stop":
		doStop(conn, params)
	default:
		printSubCommandsHelp()
		fmt.Printf("Valid parameters for all subcommands")	
		os.Exit(1)
	}
}

func doStatus(conn Connection, params Parameters) {
	fmt.Printf("command: %+v\n", params.Cmd)
	cmd := pb.StatusRequest{
		Id: params.Cmd[0],
	}
	resp, err := conn.Client.Status(conn.Ctx, &cmd)
	if err != nil {
		fmt.Printf("Executing start command failed: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Job ID: %s\n\t%+v", resp.Id, resp)
}
func doStop(conn Connection, params Parameters) {
	fmt.Printf("command: %+v\n", params.Cmd)
	cmd := pb.StopRequest{
		Id: params.Cmd[0],
	}
	resp, err := conn.Client.Stop(conn.Ctx, &cmd)
	if err != nil {
		fmt.Printf("Executing start command failed: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Job ID: %+v", resp)
}

func doStart(conn Connection, params Parameters) {
	fmt.Printf("command: %+v\n", params.Cmd)
	cmd := pb.StartRequest{
		Command: params.Cmd[0],
		Args: params.Cmd[1:],
	}

	resp, err := conn.Client.Start(conn.Ctx, &cmd)
	if err != nil {
		fmt.Printf("Executing start command failed: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("%v\n", resp)
}

func (c *Connection) Done() {
	c.conn.Close()
	c.Cancel()
}

func argParse() Parameters {

	port := flag.Int("port", 50051, "remote port to connect to")
	host := flag.String("host", "127.0.0.1", "remote host to connect to")
	ident := flag.String("ident", "", "config file with paths to ssl certs and keys (required)")
	help := flag.Bool("help", false, "print help")
	flag.Parse()
	args := flag.Args()

	/*
	if *ident == "" {
		fmt.Println("missing ident config")
		//os.Exit(1)
	}
		*/

	params := Parameters{
		Port:  *port,
		Host:  *host,
		Ident: *ident,
		Help:  *help,
	}
	fmt.Printf("params: %+v\n", params)

	if len(args) == 0 {
		printSubCommandsHelp()
		fmt.Println("Valid parameters for all subcommands")
		flag.PrintDefaults()
		os.Exit(1)
	}
	params.SubCmd = args[0]
	if args[1] == "--" {
		params.Cmd = args[2:]
	} else {
		params.Cmd = args[1:]
	}
	return params
}

func printOptions() {
	fmt.Print("\nOptions:\n\n")
	flag.PrintDefaults()
	configHelp := `
Where -ident is the path to a JSON file with information about SSL certs and keys:
{
	"ca-cert": "<path to self-signed CA certificate>",
	"public-key": "<path to host's public key>",
	"host-cert": "<path to host cert signed by 'ca-cert'">
}

`
	fmt.Print(configHelp)
	fmt.Println("")
}

func printSubCommandsHelp() {
	fmt.Println("Valid subcommands:")
	fmt.Println("\tstart")
	fmt.Println("\tstop")
	fmt.Println("\toutput")
	fmt.Println("\tstatus\n")
	fmt.Println("Get help for a subcommand:")
	fmt.Println("\ttrc-client -help <subcommand>")
}

func NewConnection(params Parameters) (Connection, error) {
	connect := Connection{}
	var err error

	// Setup TLS
	// read ca's cert
	/*
	caCert, err := ioutil.ReadFile(
		"/Users/stoo/playgound/golang/teleport/certs/ca/ca-cert.pem")
	if err != nil {
		return connect, err
	}

	// create cert pool and append ca's cert
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		return connect, fmt.Errorf("error appending ca cert: %v", certPool)
	}

	//read client cert
	clientCert, err := tls.LoadX509KeyPair(
		"/Users/stoo/playgound/golang/teleport/certs/user1/user1-cert.pem",
		"/Users/stoo/playgound/golang/teleport/certs/user1/user1-key.pem")

	if err != nil {
		return connect, err
	}
		*/

	// set config of tls credential
	/*
	config := &tls.Config{
		Certificates: []tls.Certificate{},
		RootCAs:      nil,
		MinVersion: tls.VersionTLS13,
	}
		*/

	//tlsCredential := credentials.NewTLS(config)
	//var tlsCredential credentials.TransportCredentials

	url := fmt.Sprintf("%s:%d", params.Host, params.Port)

	connect.conn, err = grpc.Dial(
		url,
		//grpc.WithTransportCredentials(tlsCredential),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return connect, err
	}

	connect.Client = pb.NewAgentClient(connect.conn)

	connect.Ctx, connect.Cancel = context.WithTimeout(context.Background(), time.Second)

	return connect, nil
}
