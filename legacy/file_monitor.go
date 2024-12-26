package main

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
)

func MonitorFile(lb *LoadBalancer) {
  watcher, err := fsnotify.NewWatcher()
  if err != nil {
    fmt.Println("WARNING: Error watching")
  }
  defer watcher.Close()
  err = watcher.Add("/app/docker-compose.yaml")
  if err != nil {
    panic(err)
  }

  // Start listening for change
  go func() {
    for {
      select {
      case event ,ok := <- watcher.Events:
        if !ok {
          fmt.Println("hey")
          return
        }
        fmt.Println("Events: ", event)
      case event, ok := <- watcher.Errors:
        if !ok {
          return
        }
        fmt.Println("Error: ", event)
      }
    }
  }()
  
}
