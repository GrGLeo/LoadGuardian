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

func ExecuteCommand(cp CommandProvider) (string, error) {
  lg := loadguardian.GetLoadGuardian()
  cmd := cp.GetCommand()
  commandName := cmd.Name
  zaplog.Infof("comand: %v", cmd)
  switch commandName {
  case "up":
    file := cmd.Args.File
    resp, err := loadguardian.StartProcress(file) 
    if err != nil {
      return "", err
    }
    return resp, nil

  case "down":
    lg.CleanUp()
    os.Exit(0)

  case "update":
    resp, err := loadguardian.UpdateProcess(cmd.Args.File)
    if err != nil {
      return "", err
    }
    return resp, nil


  case "info":
    lg.Logger.Infoln("Gathering services information")
    resp, err := loadguardian.InfoProcess()
    if err != nil {
      return "", err
    }
    return resp, nil


  default:
    lg.Logger.Warnf("Unknown command:", commandName)
    return "Unknown command", nil
  }
  return "", nil
}

