package command

import (
	"github.com/GrGLeo/LoadBalancer/src/internal/cmdclient"
	"github.com/spf13/cobra"
)

func init() {
  rootCmd.AddCommand(UpdateCmd)
  UpdateCmd.Flags().IntVarP(&schedule, "schedule", "s", 0, "Add n hours to the command execution")
  UpdateCmd.Flags().StringVarP(&file, "file", "f", "service.yml", "Path to the config file")
}

var UpdateCmd = &cobra.Command{
  Use:   "update",
  Short: "Update the config",
  Long:  `Update the orchestration with a new config file`,
  Run: func(cmd *cobra.Command, args []string) {
    updatecmd := &cmdclient.UpdateCommand{
      Name: "update",
      File: file,
      Schedule: schedule,
    }
    cmdclient.SendCommand(updatecmd)
  },
} 
