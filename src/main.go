package main

import (
	"flag"
	"fmt"
	"time"

	command "github.com/GrGLeo/LoadBalancer/src/cmd"
  //"github.com/GrGLeo/LoadBalancer/src/internal/cmdclient"
	cmdserver "github.com/GrGLeo/LoadBalancer/src/internal/cmdserver"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var (
  scheduleDelay int
  File          string
  zaplog        = zap.L().Sugar()
)

func init() {
  // Create Logger
  zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))
  // Load .env
  err := godotenv.Load()
  if err != nil {
    zap.L().Sugar().Warn("Failed to read .env\n Env variable might not be set correctly")
  }

  // Flag configuration
  flag.IntVar(&scheduleDelay, "schedule", 0, "add n hour to the command execution")
  flag.StringVar(&File, "file", "service.yml", "path to the config file")
}

func main() {
  command.Execute()
}




func main_old() {
  // Reading cmd
  flag.Parse()
  args := flag.Args()
  cmd := args[0]
  
  fmt.Printf("Command send: %q", cmd)
  fmt.Printf("File send: %q", File)
  switch cmd {
  case "up":
    if scheduleDelay > 0 {
      executeTime := time.Now().Add(time.Duration(scheduleDelay) * time.Hour)
      _ = cmdserver.ScheduleCommand{
        Name: cmd,
        Args: cmdserver.CommandArgs{
          File: File,
        }, 
        ExecuteTime: executeTime,
      }
    } else {
      cmdserver.ExecuteCommand(cmdserver.RunnableCommand{
        Name: cmd,
        Args: cmdserver.CommandArgs{
          File: File,
        }, 
      })
  }

  case "down":
    SendGeneralCommand(cmd, File, scheduleDelay)
  case "update":
    SendGeneralCommand(cmd, File, scheduleDelay)
  default:
    fmt.Println("Unknown command")
  }
}

func SendGeneralCommand(cmd, file string, schedule int) {
  //err := cmdclient.SendCommand(cmdclient.PrepCommand(cmd, File, scheduleDelay))
  //if err != nil {
  //zap.L().Sugar().Fatalln("Fail to clean up process")
  //}
}
