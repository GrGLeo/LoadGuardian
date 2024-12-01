package main

import (
	"math/rand"
	"sort"
)

func (lb *LoadBalancer) getBackend() *BackendService {
  switch lb.Algorithm {
  case "leastconnection":
    return lb.getLeastConnection()
  case "roundrobin":
    return lb.getNextBackend()
  case "random":
    return lb.getRandomBackend()
  default:
    return lb.getRandomBackend()
  }
}

func (lb *LoadBalancer) getNextBackend() *BackendService {
  mod := uint8(len(lb.Services))
  lb.index = (lb.index + 1) % mod
  return &lb.Services[lb.index]
}

func (lb *LoadBalancer) getRandomBackend() *BackendService {
  index := rand.Intn(len(lb.Services))
  return &lb.Services[index]
}

func (lb *LoadBalancer) getLeastConnection() *BackendService {
  sort.Slice(lb.Services, func(i, j int) bool {
    return lb.Services[i].Connection < lb.Services[j].Connection
  })
  return &lb.Services[0]
}
