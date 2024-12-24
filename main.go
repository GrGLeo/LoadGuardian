package main

import (
	"fmt"
	"os"
)

func main() {
  switch os.Args[1] {
    case "up":
      fmt.Println("heyo")
      Up()

    default:
      fmt.Println("Unknown command")
    }
  }
