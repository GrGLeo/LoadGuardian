package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/GrGLeo/LoadBalancer/src/internal/cmdclient"
	"github.com/GrGLeo/LoadBalancer/src/internal/cmdserver"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var scheduleDelay int
var File string
var zaplog = zap.L().Sugar()

func init() {
  zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))

  // Flag configuration
  flag.IntVar(&scheduleDelay, "schedule", 0, "add n hour to the command execution")
  flag.StringVar(&File, "file", "service.yml", "path to the config file")
}

func main() {
  // Loading env variable for parsing
  err := godotenv.Load()
  if err != nil {
    zap.L().Sugar().Warn("Failed to read .env\n Env variable might not be set correctly")
  }

  // Reading cmd
  flag.Parse()
  args := flag.Args()
  cmd := args[0]
  
  fmt.Printf("Command send: %q", cmd)
  switch cmd {
  case "up":
    if scheduleDelay > 0 {
      executeTime := time.Now().Add(time.Duration(scheduleDelay) * time.Hour)
      _ = command.ScheduleCommand{
        Name: cmd,
        Args: command.CommandArgs{
          File: File,
        }, 
        ExecuteTime: executeTime,
      }
    } else {
      command.ExecuteCommand(command.RunnableCommand{
        Name: cmd,
        Args: command.CommandArgs{
          File: File,
        }, 
      })
  }

  case "down":
    SendGeneralCommand(cmd, File, scheduleDelay)
  case "update":
    zaplog.Infof("Command send: %s, %s, %d", cmd, File, scheduleDelay)
    SendGeneralCommand(cmd, File, scheduleDelay)
  default:
    fmt.Println("Unknown command")
  }
}

func SendGeneralCommand(cmd, file string, schedule int) {
  err := cmdclient.SendCommand(cmdclient.PrepCommand(cmd, File, scheduleDelay))
  if err != nil {
    zap.L().Sugar().Fatalln("Fail to clean up process")
  }
}
