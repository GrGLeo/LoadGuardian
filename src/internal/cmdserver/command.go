package cmdserver

import (
	"os"
	"time"

	"github.com/GrGLeo/LoadBalancer/src/internal/loadguardian"
	"go.uber.org/zap"
)

var zaplog = zap.L().Sugar()

type RunnableCommand struct {
  Name string
  Args CommandArgs 
}

type ScheduleCommand struct {
  Name string 
  Args CommandArgs 
  ExecuteTime time.Time
}

type CommandArgs  struct {
  File string
}

type CommandProvider interface {
  GetCommand() RunnableCommand
}

func (rc RunnableCommand) GetCommand() RunnableCommand {
  return rc
}

func (sc ScheduleCommand) GetCommand() RunnableCommand {
  return RunnableCommand{
    Name: sc.Name,
    Args: sc.Args,
  }
}

func ExecuteCommand(cp CommandProvider) {
  lg := loadguardian.GetLoadGuardian()
  cmd := cp.GetCommand()
  commandName := cmd.Name
  zaplog.Infof("comand: %v", cmd)
  switch commandName {
  case "up":
    file := cmd.Args.File
    loadguardian.StartProcress(file) 

  case "down":
    lg.CleanUp()
    os.Exit(0)

  case "update":
    zaplog.Infoln("I am called here")
    loadguardian.UpdateProcess(cmd.Args.File)

  default:
    zaplog.Warnf("Unknown command:", commandName)
  }
}

