package cmdclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"go.uber.org/zap"
)

const socketPath = "/tmp/loadguardian.sock"
var zaplog = zap.L().Sugar()

type Command struct {
  Name string `json:"name"`
  File string `json:"file"`
  Schedule int `json:"schedule"`
}


func PrepCommand(name, file string, schedule int) []byte {
  if name == "" {
    zaplog.Fatalln("No action passed")
  }
  cmd := Command{
    Name: name,
    File: file,
    Schedule: schedule,
  }
  cmdJson, err := json.Marshal(cmd)
  if err != nil {
    zaplog.Fatalf("Failed to parsed command: %q", err.Error())
  }
  zaplog.Infoln(string(cmdJson))
  return cmdJson
}


func SendCommand(byteCommand []byte) error {
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
  fmt.Println("Response from guardian:", string(buff[:n]))
  return nil
}
