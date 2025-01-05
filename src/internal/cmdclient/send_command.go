package cmdclient

import (
	"errors"
	"fmt"
	"net"

	"go.uber.org/zap"
)

const socketPath = "/tmp/loadguardian.sock"
var zaplog = zap.L().Sugar()


func SendCommand(command commandConfig) error {
  byteCommand := command.PrepCommand()
  conn, err := net.Dial("unix", socketPath)
  if err != nil {
    return errors.New("Failed to connect to the running guardian process")
  }
  defer conn.Close()

  // Write command
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
  fmt.Println("Response from guardian")
  fmt.Print(string(buff[:n]))
  return nil
}
