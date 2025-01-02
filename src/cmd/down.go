package command

import (
	"github.com/GrGLeo/LoadBalancer/src/internal/cmdclient"
	"github.com/spf13/cobra"
)

func init() {
  rootCmd.AddCommand(DownCmd)
  DownCmd.Flags().IntVarP(&schedule, "schedule", "s", 0, "Add n hours to the command execution")
}

var DownCmd = &cobra.Command{
  Use:   "down",
  Short: "Stop the process",
  Long:  `Complitly stop and exit the orchestration`,
  Run: func(cmd *cobra.Command, args []string) {
    downcmd := &cmdclient.DownCommand{
      Name: "down",
      Schedule: schedule,
    }
    cmdclient.SendCommand(downcmd)
  },
} 
