package command

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	cmdserver "github.com/GrGLeo/LoadBalancer/src/internal/cmdserver"
	"github.com/GrGLeo/LoadBalancer/src/internal/loadguardian"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)


const socketPath = "/tmp/loadguardian.sock"

var zaplog = zap.L().Sugar()
var rootCmd = &cobra.Command{
  Use:   "loadguardian",
  Short: "LoadGuardian is container orchestrator",
  Long: `A Fast and Flexible Static Site Generator built with
                love by spf13 and friends in Go.
                Complete documentation is available at https://gohugo.io/documentation/`,
  Run: func(cmd *cobra.Command, args []string) {
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

    // Handle socket command
    go func() {
      for {
        conn, err := listener.Accept()
        if err != nil  {
          zaplog.Warnf("Error accepting connection: %s", err.Error())
          continue
        }
        go cmdserver.HandleSocketCommand(conn, lg)

      }
  }()
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
