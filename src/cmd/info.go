package command

import (
	"github.com/GrGLeo/LoadBalancer/src/internal/cmdclient"
	"github.com/spf13/cobra"
)

func init() {
  rootCmd.AddCommand(InfoCmd)
}

var InfoCmd = &cobra.Command{
  Use:   "info",
  Short: "Retrieve and show services information",
  Long:  `Retrieve per container information:
  - Name
  - Health
  - CPU usage
  - Memory usage
  `,
  Run: func(cmd *cobra.Command, args []string) {
    infocmd := &cmdclient.UpCommand{
      Name: "info",
    }
    cmdclient.SendCommand(infocmd)
  },
} 
