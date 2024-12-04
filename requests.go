package main

import (
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
  "errors"
)

func (lb *LoadBalancer) handleRequest(w http.ResponseWriter, r *http.Request) {
  var backend *BackendService
  var resp *http.Response
  var err error

  for len(lb.Services) > 0 {
    backend = lb.getBackend()
    targetURL := backend.Endpoint + r.URL.Path
    resp, err = ForwardRequests(targetURL, w, r)
    // Forwarded to the service
    if err == nil {
      break
    }
    // Verifying container
    if err.Error() == "Container not responding" {
      fmt.Println("Checking container status")
      backend.CheckStatus(lb.DockerClient)
      lb.RemoveDeadServices()
    }
  }

  // No more service available
  if resp == nil {
    fmt.Println("Failed to found avaible service")
    http.Error(w, "Service unavailable", 503)
    return
  }

  atomic.AddInt32(&backend.Connection, 1)
  defer resp.Body.Close()

  // Return response
  for key, values := range resp.Header{
    for _, value := range values {
      w.Header().Add(key, value)
    }
  }
  w.WriteHeader(resp.StatusCode)
  io.Copy(w, resp.Body)
  atomic.AddInt32(&backend.Connection, -1)
}
 
func ForwardRequests(targetURL string, w http.ResponseWriter, r *http.Request) (*http.Response, error) {
  var body io.Reader
  if r.Body != nil {
    body = r.Body
  } else {
    body = http.NoBody
  }

  req, err := http.NewRequest(r.Method, targetURL, body)
  if err != nil {
    http.Error(w, "Failed to create request", 500)
    return nil, err
  }
  for key, values := range r.Header {
    for _, value := range values {
      req.Header.Add(key, value)
    }
  }

  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    return nil, errors.New("Container not responding")
  }
  return resp, nil
}
