package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func init() {
  zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))
}

func main() {
  // Loading env variable for parsing
  err := godotenv.Load()
  if err != nil {
    zap.L().Sugar().Warn("Failed to read .env\n Env variable might not be set correctly")
  }

  // Reading command
  command := os.Args[1]
  switch command {
  case "up":
    if len(os.Args) < 3 {
      fmt.Println("Usage up ./LoadGuardian up <file>")
      os.Exit(1)
    }
    file := os.Args[2]
    Up(file)

  case "down":
    if err := Down(); err != nil {
      zap.L().Sugar().Fatalln("Fail to clean up process")
      fmt.Println(err.Error())
      os.Exit(1)
    }

  case "update":
    if len(os.Args) < 3 {
      fmt.Println("Usage update ./LoagdGuardian update <file>")
      os.Exit(1)
    }
    file := os.Args[2]
    if err := Update(file); err != nil {
      fmt.Println(err.Error())
      os.Exit(1)
    }

  default:
    fmt.Println("Unknown command")
  }
}
