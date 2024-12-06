package main

import (
	"fmt"
	"net/http"
	"time"
)

const URL string = ""
var PORTS [2]string 


func main () {

  lb, err := NewLoadBalancer()
  if err != nil {
    panic(err)
  }
  // Routine to check changes on docker-compose
  go MonitorFile(lb)
  // Routine to periodically update Stats
  go func() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    for {
      select {
        case <- ticker.C:
          err := lb.Monitor()
          if err != nil {
            fmt.Printf("Error getting docker stats: %v\n", err)
          }
        }
      }
    }()

  http.HandleFunc("/", lb.handleRequest)
  http.ListenAndServe(":8080", nil)
}
