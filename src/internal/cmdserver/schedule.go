package cmdserver

import (
	"net"
	"time"
)

func ScheduleChecker(schedule []*ScheduleCommand, conn net.Conn) {
  for {
    for i := 0; i < len(schedule); i++ {
      command := schedule[i]
      if time.Now().After(command.ExecuteTime) {
        ExecuteCommand(command.GetCommand(), conn)
        schedule = append(schedule[:i], schedule[i+1:]...)
      }
    }
    time.Sleep(1 * time.Minute)
  }
}
