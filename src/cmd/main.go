package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func main() {
  // Loading env variable for parsing
  err := godotenv.Load()
  if err != nil {
    fmt.Println("Failed to read .env\n WARNING: Env variable might not be set correctly.")
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
      fmt.Println(err.Error())
      os.Exit(1)
    }

  default:
    fmt.Println("Unknown command")
  }
}
