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
  index int
}

func getNextBackend() string {
  return "http://backend:8081"
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
  BackendURL := getNextBackend()
  targetURL := BackendURL + r.URL.Path
  
  req, err := http.NewRequest(r.Method, targetURL, r.Body)
  if err != nil {
    http.Error(w, "Failed to create request", 500)
  }
  for key, values := range r.Header {
    for _, value := range values {
      req.Header.Add(key, value)
    }
  }

  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    http.Error(w, "Failed to forward request", 500)
  }
  defer resp.Body.Close()

  for key, values := range resp.Header{
    for _, value := range values {
      w.Header().Add(key, value)
    }
  }
  
  w.WriteHeader(resp.StatusCode)
  io.Copy(w, resp.Body)
}

func main () {

  mux := http.NewServeMux()
  http.HandleFunc("/", handleRequest)
  Server := &http.Server{
    Addr: ":8080",
    Handler: mux,
  }
  
  if err := Server.ListenAndServe(); err != nil {
    fmt.Printf("Error: %q. Shutting down.", err.Error())
  }
}
