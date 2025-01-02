package cmdserver

import (
	"time"
)

func ScheduleChecker() {
  schedulesCmd := []*ScheduleCommand{}
  for {
    select {
    case newCmd := <- scheduleCmdCh:
      schedulesCmd = append(schedulesCmd, newCmd)
    
    default:
      for i := 0; i < len(schedulesCmd); i++ {
        command := schedulesCmd[i]
        if time.Now().After(command.ExecuteTime) {
          ExecuteCommand(command.GetCommand())
          schedulesCmd = append(schedulesCmd[:i], schedulesCmd[i+1:]...)
        }
      }
      time.Sleep(1 * time.Minute)
    }
  }
}
