package main

const URL string = ""
var PORTS [2]string 

type BackendService struct {
  endpoint string
  connection int
}

type LoadBalancer struct {
  services []BackendService
}

func main () {
}
