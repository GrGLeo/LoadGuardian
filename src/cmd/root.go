package command

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	cmdserver "github.com/GrGLeo/LoadBalancer/src/internal/cmdserver"
	"github.com/GrGLeo/LoadBalancer/src/internal/loadguardian"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)


const (
  socketPath = "/tmp/loadguardian.sock"
  text = `
888                            888  .d8888b.                                888 d8b                   
888                            888 d88P  Y88b                               888 Y8P                   
888                            888 888    888                               888                       
888      .d88b.   8888b.   .d88888 888        888  888  8888b.  888d888 .d88888 888  8888b.  88888b.  
888     d88""88b     "88b d88" 888 888  88888 888  888     "88b 888P"  d88" 888 888     "88b 888 "88b 
888     888  888 .d888888 888  888 888    888 888  888 .d888888 888    888  888 888 .d888888 888  888 
888     Y88..88P 888  888 Y88b 888 Y88b  d88P Y88b 888 888  888 888    Y88b 888 888 888  888 888  888 
88888888 "Y88P"  "Y888888  "Y88888  "Y8888P88  "Y88888 "Y888888 888     "Y88888 888 "Y888888 888  888 
%s
`
)


var version = os.Getenv("VERSION")
var zaplog = zap.L().Sugar()
var rootCmd = &cobra.Command{
  Use:   "loadguardian",
  Short: "LoadGuardian is container orchestrator",
  Long: `A Fast and Flexible Static Site Generator built with
                love by spf13 and friends in Go.
                Complete documentation is available at https://gohugo.io/documentation/`,
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Printf(text, version)
    fmt.Printf("v%s\n", version)
    fmt.Println("Ready and listening for upcoming command.")
    // Setting up bare loadguardian
    lg := loadguardian.GetLoadGuardian()
    // Setting up socket to listen for upcoming command
    os.Remove(socketPath)
    listener, err := net.Listen("unix", socketPath)
    if err != nil {
      zaplog.Fatalf("Failed to open socket. Will not listen for upcoming command: %s", err.Error())
    }
    defer listener.Close()
    defer os.Remove(socketPath)

    var scheduleCmd []*cmdserver.ScheduleCommand
    // Handle socket command
    go func() {
      for {
        conn, err := listener.Accept()
        if err != nil  {
          zaplog.Warnf("Error accepting connection: %s", err.Error())
          continue
        }
        go cmdserver.HandleSocketCommand(conn, lg, &scheduleCmd)
      }
  }()
  // Check and run schedule command
  go cmdserver.ScheduleChecker()

  // Handle keyboard shutdown
  signalChannel := make(chan os.Signal, 1)
  signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
  <-signalChannel
  // Clean up
  lg.CleanUp()
  },
}

func Execute() {
  if err := rootCmd.Execute(); err != nil {
    zaplog.Fatalf("Failed to execute root cmd: %s\n", err.Error())
  }
}
