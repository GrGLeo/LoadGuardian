package main

import (
	"fmt"
	"io"
	"net/http"
)

const URL string = ""
var PORTS [2]string 

type BackendService struct {
  Endpoint string
  Connection int
}

type LoadBalancer struct {
  Services []BackendService
  index uint8
}

func (lb *LoadBalancer) getNextBackend() string {
  mod := uint8(len(lb.Services))
  lb.index = (lb.index + 1) % mod
  return lb.Services[lb.index].Endpoint
}

func (lb *LoadBalancer) handleRequest(w http.ResponseWriter, r *http.Request) {
  BackendURL := lb.getNextBackend()
  targetURL := BackendURL + r.URL.Path
  fmt.Printf("\n1: %q\n", targetURL)

  var body io.Reader
      if r.Body != nil {
          body = r.Body
      } else {
          body = http.NoBody
      }

  req, err := http.NewRequest(r.Method, targetURL, body)
  fmt.Printf("\n2: %q\n", targetURL)
  if err != nil {
    http.Error(w, "Failed to create request", 500)
    return
  }
  for key, values := range r.Header {
    for _, value := range values {
      req.Header.Add(key, value)
    }
  }
  fmt.Printf("\n3: %q\n", targetURL)

  client := &http.Client{}
  resp, err := client.Do(req)
  fmt.Printf("\n4: %q\n", targetURL)
  if err != nil {
    fmt.Printf("\n5: %q\n", targetURL)
    http.Error(w, "Failed to forward request", 500)
    return
  }
  defer resp.Body.Close()

  for key, values := range resp.Header{
    for _, value := range values {
      w.Header().Add(key, value)
    }
  }
  fmt.Printf("\n%q\n", targetURL)
  
  w.WriteHeader(resp.StatusCode)
  io.Copy(w, resp.Body)
}

func main () {
  services := CreateBackendServices()
  lb := LoadBalancer{
    Services: services,
    index: 0,
  }

  http.HandleFunc("/", lb.handleRequest)
  http.ListenAndServe(":8080", nil)
}
