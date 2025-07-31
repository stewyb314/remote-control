package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
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
	case "output":
		doOutput(conn, params)

	default:
		printSubCommandsHelp()
		fmt.Printf("Valid parameters for all subcommands")	
		os.Exit(1)
	}
}

func doOutput(conn Connection, params Parameters) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	cmd := pb.OutputRequest{
		Id: params.Cmd[0],
	}
	resp, err := conn.Client.Output(conn.Ctx, &cmd)
	if err != nil {
		fmt.Printf("Executing start command failed: %s\n", err)
		os.Exit(1)
	}

	for {
		select {
		case <-sigChan:
			fmt.Println("Received interrupt signal, exiting...")
			return
		default:
			line, err := resp.Recv()
			if err != nil {
				if err.Error() == "EOF" {
					return
				} else {
					fmt.Printf("Error receiving output: %s\n", err)
					return
				}	
			}
			fmt.Printf("%s\n", line.Output)

		}
	}
}

func doStatus(conn Connection, params Parameters) {
	cmd := pb.StatusRequest{
		Id: params.Cmd[0],
	}
	resp, err := conn.Client.Status(conn.Ctx, &cmd)
	if err != nil {
		fmt.Printf("Executing start command failed: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Job ID: %s\nCommand: %s\n Args: %v\n Status: %s\n Exit code: %d\n", resp.Id, resp.Cmd, resp.State, resp.Args, resp.Exit)
}
func doStop(conn Connection, params Parameters) {
	cmd := pb.StopRequest{
		Id: params.Cmd[0],
	}
	resp, err := conn.Client.Stop(conn.Ctx, &cmd)
	if err != nil {
		fmt.Printf("Executing start command failed: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Job ID: %s stopped\n", resp.Id)
}

func doStart(conn Connection, params Parameters) {
	cmd := pb.StartRequest{
		Command: params.Cmd[0],
		Args: params.Cmd[1:],
	}

	resp, err := conn.Client.Start(conn.Ctx, &cmd)
	if err != nil {
		fmt.Printf("Executing start command failed: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("ID: %s\n", resp)
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


	params := Parameters{
		Port:  *port,
		Host:  *host,
		Ident: *ident,
		Help:  *help,
	}

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

	url := fmt.Sprintf("%s:%d", params.Host, params.Port)

	connect.conn, err = grpc.Dial(
		url,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return connect, err
	}

	connect.Client = pb.NewAgentClient(connect.conn)

	connect.Ctx, connect.Cancel = context.WithTimeout(context.Background(), time.Second)

	return connect, nil
}
