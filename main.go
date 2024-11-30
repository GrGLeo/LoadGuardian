package main

import (
	"net/http"
)

const URL string = ""
var PORTS [2]string 


func main () {
  lb, err := NewLoadBalancer()
  if err != nil {
    panic(err)
  }
  http.HandleFunc("/", lb.handleRequest)
  http.ListenAndServe(":8080", nil)
}
