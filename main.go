package main

import (
	"fmt"
	"io"
	"net/http"
)

const URL string = ""
var PORTS [2]string 


func main () {
  services := CreateBackendServices()
  lb := LoadBalancer{
    Services: services,
    index: 0,
  }

  http.HandleFunc("/", lb.handleRequest)
  http.ListenAndServe(":8080", nil)
}
