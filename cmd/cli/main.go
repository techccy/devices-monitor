package main

import (
	"flag"
	"fmt"
	"github.com/ccy/devices-monitor/internal/cli"
	"github.com/ccy/devices-monitor/pkg/config"
	"os"
)

func main() {
	configFile := flag.String("config", "", "Configuration file path")
	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	var cfg *config.CLIConfig
	var err error
	var srvURL string

	if *configFile != "" {
		cfg, err = config.LoadCLIConfig(*configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
			os.Exit(1)
		}
		srvURL = cfg.ServerURL
	} else {
		srvURL = *serverURL
	}

	c := cli.NewCLI(srvURL)

	command := args[0]

	switch command {
	case "register":
		if len(args) < 3 {
			fmt.Println("Usage: ccy register -u <email> -p <password>")
			os.Exit(1)
		}
		email := ""
		password := ""
		for i := 1; i < len(args); i++ {
			if args[i] == "-u" && i+1 < len(args) {
				email = args[i+1]
				i++
			} else if args[i] == "-p" && i+1 < len(args) {
				password = args[i+1]
				i++
			}
		}
		if email == "" || password == "" {
			fmt.Println("Usage: ccy register -u <email> -p <password>")
			os.Exit(1)
		}
		if err := c.Register(email, password); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "login":
		if len(args) < 3 {
			fmt.Println("Usage: ccy login -u <email> -p <password>")
			os.Exit(1)
		}
		email := ""
		password := ""
		for i := 1; i < len(args); i++ {
			if args[i] == "-u" && i+1 < len(args) {
				email = args[i+1]
				i++
			} else if args[i] == "-p" && i+1 < len(args) {
				password = args[i+1]
				i++
			}
		}
		if email == "" || password == "" {
			fmt.Println("Usage: ccy login -u <email> -p <password>")
			os.Exit(1)
		}
		if err := c.Login(email, password); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "ls":
		if err := c.ListDevices(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "status":
		if len(args) < 2 {
			fmt.Println("Usage: ccy status <deviceID>")
			os.Exit(1)
		}
		if err := c.GetStatus(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "net":
		if len(args) < 2 {
			fmt.Println("Usage: ccy net <deviceID>")
			os.Exit(1)
		}
		if err := c.GetNetworkInfo(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "ssh":
		if len(args) < 2 {
			fmt.Println("Usage: ccy ssh <deviceID>")
			os.Exit(1)
		}
		if err := c.SSH(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "logout":
		if err := c.Logout(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "on":
		if err := c.StartAgent(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "add-device":
		if len(args) < 3 {
			fmt.Println("Usage: ccy add-device -n <name> -i <identifier>")
			os.Exit(1)
		}
		name := ""
		identifier := ""
		for i := 1; i < len(args); i++ {
			if args[i] == "-n" && i+1 < len(args) {
				name = args[i+1]
				i++
			} else if args[i] == "-i" && i+1 < len(args) {
				identifier = args[i+1]
				i++
			}
		}
		if name == "" || identifier == "" {
			fmt.Println("Usage: ccy add-device -n <name> -i <identifier>")
			os.Exit(1)
		}
		if err := c.RegisterDevice(name, identifier); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("CCY Terminal Remote Monitoring and Management System")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ccy register -u <email> -p <password>  Register a new user")
	fmt.Println("  ccy login -u <email> -p <password>    Login to the system")
	fmt.Println("  ccy add-device -n <name> -i <identifier>  Register a new device")
	fmt.Println("  ccy on                                 Start agent daemon")
	fmt.Println("  ccy ls                                 List all devices")
	fmt.Println("  ccy status <deviceID>                  Query device status")
	fmt.Println("  ccy net <deviceID>                    Query network info")
	fmt.Println("  ccy ssh <deviceID>                    SSH to device")
	fmt.Println("  ccy logout                             Logout and clear credentials")
	fmt.Println()
}
