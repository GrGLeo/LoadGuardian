package main

import (
	"fmt"
	"os"
)

func main() {
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
