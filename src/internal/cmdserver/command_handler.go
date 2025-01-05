package cmdserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/GrGLeo/LoadBalancer/src/internal/cmdclient"
	"github.com/GrGLeo/LoadBalancer/src/internal/loadguardian"
)


const socketPath = "/tmp/loadguardian.sock"

var scheduleCmdCh = make(chan *ScheduleCommand, 100)

func HandleSocketCommand(conn net.Conn, lg *loadguardian.LoadGuardian, scheduleCmd *[]*ScheduleCommand) {
  zaplog.Errorf("Am i here??")
  defer conn.Close()
  var baseCmd struct {
    Name string `json:"name"`
  }
  buff := make([]byte, 1024)
  n, err := conn.Read(buff)
  if err != nil {
    zaplog.Errorf("Failed to read command: %s\n", err.Error())
    return
  }

  data := buff[:n]
  if err := json.Unmarshal(data, &baseCmd); err != nil {
    zaplog.Errorf("Failed to unmarshal command: %s\n", err.Error())
    return
  }
  zaplog.Infof("command receive: %q\n", baseCmd.Name)

  switch baseCmd.Name {
  case "up":
    var upCmd cmdclient.UpCommand
    if err := json.Unmarshal(data, &upCmd); err != nil {
      zaplog.Errorf("Failed to parse up command: %s\n", err.Error())
      return
    }
    zaplog.Infof("Processing UpCommand: %+v\n", upCmd)

    scheduleDelay := upCmd.Schedule
    File := upCmd.File
    if upCmd.Schedule > 0 {
      executeTime := time.Now().Add(time.Duration(scheduleDelay) * time.Minute)
      UpCmd := ScheduleCommand{
        Name: upCmd.Name,
        Args: CommandArgs{
          File: File,
        }, 
        ExecuteTime: executeTime,
      }
      scheduleCmdCh <- &UpCmd
      conn.Write([]byte(fmt.Sprintf("Command schedule for: %s", executeTime.Format(time.ANSIC))))
    } else {
      ExecuteCommand(RunnableCommand{
        Name: upCmd.Name,
        Args: CommandArgs{
          File: File,
        }, 
      })
      conn.Write([]byte("Command executed successfully"))
    }

  case "down":
    var downCmd cmdclient.DownCommand
    if err := json.Unmarshal(data, &downCmd); err != nil {
      zaplog.Errorf("Failed to parse up command: %s\n", err.Error())
      return
    }
    lg.Logger.Infof("Processing UpCommand: %+v\n", downCmd)

    scheduleDelay := downCmd.Schedule
    if downCmd.Schedule > 0 {
      executeTime := time.Now().Add(time.Duration(scheduleDelay) * time.Hour)
      DownCmd := ScheduleCommand{
        Name: downCmd.Name,
        Args: CommandArgs{
        }, 
        ExecuteTime: executeTime,
      }
      scheduleCmdCh <- &DownCmd
      conn.Write([]byte(fmt.Sprintf("Command schedule for: %s", executeTime.Format("ANSIC"))))
    } else {
      ExecuteCommand(RunnableCommand{
        Name: downCmd.Name,
        Args: CommandArgs{
        }, 
      })
      conn.Write([]byte("Command executed successfully"))
    }

  case "update":
    var updateCmd cmdclient.UpdateCommand
    if err := json.Unmarshal(data, &updateCmd); err != nil {
      zaplog.Errorf("Failed to parse up command: %s\n", err.Error())
      return
    }
    zaplog.Infof("Processing UpCommand: %+v\n", updateCmd)
    scheduleDelay := updateCmd.Schedule
    File := updateCmd.File
    if updateCmd.Schedule > 0 {
      executeTime := time.Now().Add(time.Duration(scheduleDelay) * time.Hour)
      UpdateCmd := ScheduleCommand{
        Name: updateCmd.Name,
        Args: CommandArgs{
        File: File,
        }, 
        ExecuteTime: executeTime,
      }
      scheduleCmdCh <- &UpdateCmd
      conn.Write([]byte(fmt.Sprintf("Command schedule for: %s", executeTime.Format("ANSIC"))))
    } else {
      ExecuteCommand(RunnableCommand{
        Name: updateCmd.Name,
        Args: CommandArgs{
          File: File,
        }, 
      })
      conn.Write([]byte("Command executed successfully"))
    }

  case "info":
    // since info command only has one field Name we dont need to unmarshal the rest
    resp, _ := ExecuteCommand(RunnableCommand{
      Name: "info",
    })
    conn.Write([]byte(resp))
  }
}

func Dowk() error {
  err := SendCommand("down")
  return err
}

func Updahe(file string) error {
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
