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

  go func() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    for {
      select {
        case <- ticker.C:
          newID := lb.Services[0].ScaleUpService(lb.DockerClient)
          backend, err := CreateBackend(lb.DockerClient, newID)
          if err != nil {
            fmt.Println(err.Error())
          }
          lb.Services = append(lb.Services, backend)
        }
      }
    }()
  lb.Services[0].ScaleDownService(lb.DockerClient)

  http.HandleFunc("/", lb.handleRequest)
  http.ListenAndServe(":8080", nil)
}
