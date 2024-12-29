package command

import (
	"github.com/GrGLeo/LoadBalancer/src/internal/cmdclient"
	"github.com/spf13/cobra"
)

var (
  file string
  schedule int
)

func init() {
  rootCmd.AddCommand(UpCmd)
  UpCmd.Flags().StringVarP(&file, "file", "f", "service.yml", "Path to the config file")
  UpCmd.Flags().IntVarP(&schedule, "schedule", "s", 0, "Add n hours to the command execution")
}

var UpCmd = &cobra.Command{
  Use:   "up",
  Short: "Start the orchestration given the config file",
  Long:  `Start the orchestration of services defined in the configuration file.`,
  Run: func(cmd *cobra.Command, args []string) {
    upcmd := &cmdclient.UpCommand{
      Name: "up",
      File: file,
      Schedule: schedule,
    }
    cmdclient.SendCommand(upcmd)
  },
} 
