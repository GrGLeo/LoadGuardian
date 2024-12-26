package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/GrGLeo/LoadBalancer/src/internal/loadguardian"
)


const socketPath = "/tmp/loadguardian.sock"


func Up(file string) {
  // Setting up socket to listen for upcoming command
  os.Remove(socketPath)
  listener, err := net.Listen("unix", socketPath)
  if err != nil {
    fmt.Println("Failed to open socket. Will not listen for upcoming command")
  }
  defer listener.Close()
  defer os.Remove(socketPath)

  // Start process
  lg := loadguardian.StartProcress(file)

  // Handle socket command
  go func() {
    for {
      conn, err := listener.Accept()
      if err != nil  {
        fmt.Println("Error accepting connection")
        continue
      }
      go handleSocketCommand(conn, &lg)
    }
  }()

  // Handle keyboard shutdown
  signalChannel := make(chan os.Signal, 1)
  signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
  <-signalChannel
  // Clean up
  lg.CleanUp()
}

func handleSocketCommand(conn net.Conn, lg *loadguardian.LoadGuardian) {
  defer conn.Close()

  buf := make([]byte, 1024)
  n, err := conn.Read(buf)
  if err != nil {
    fmt.Println("Error reading the command")
    return
  }
  command := string(buf[:n])
  // Parse command
  parsedCommand := strings.Split(command, "|")
  command = parsedCommand[0]
  switch command {
  case "down":
    lg.CleanUp()
    conn.Write([]byte("Command executed successfully"))
    os.Exit(0)

  case "update":
    if len(parsedCommand) < 2 {
      msg := "Incomplete update command"
      fmt.Println(msg)
      conn.Write([]byte(msg))
    }
    file := parsedCommand[1]
    fmt.Println(file)
    loadguardian.UpdateProcess(file)
    
    conn.Write([]byte("Command executed successfully"))

  default:
    fmt.Fprintln(conn, "Unknown command:", command)
    conn.Write([]byte("Unknown command"))
  }
}

func Down() error {
  err := SendCommand("down")
  return err
}

func Update(file string) error {
  command := fmt.Sprintf("update|%s", file)
  err := SendCommand(command)
  return err
}

func SendCommand(command string) error {
  conn, err := net.Dial("unix", socketPath)
  if err != nil {
    return errors.New("Failed to connect to the running guardian process")
  }
  defer conn.Close()

  // Write command
  byteCommand := []byte(command)
  _, err = conn.Write(byteCommand)
  if err != nil {
    return errors.New("Failed to send down command")
  }

  //Read response
  buff := make([]byte, 1024)
  n, err := conn.Read(buff)
  if err != nil {
    return errors.New("Failed to read response")
  }
  fmt.Println("Response from guardian:", string(buff[:n]))
  return nil
}
