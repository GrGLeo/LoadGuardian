package servicemanager

import (
	"errors"
	"strconv"
	"strings"
)

func (s *Service) GetPort() ([]int, error) {
  ports := []int{-1, -1}
  
  strPorts := strings.Split(s.Port[0], ":")
  cport := strPorts[0]
  hport := strPorts[1]
  if s.NextPort.Load() == 0 {
    nextPort, err := strconv.Atoi(hport)
    if err != nil {
      return ports, errors.New("Invalid external port set")
    }
    cPort, err := strconv.Atoi(cport)
    if err != nil {
      return ports, errors.New("Invalid internal port set")
    }
    s.NextPort.Store(uint32(nextPort))
    return []int{cPort, nextPort}, nil

  } else {
    s.NextPort.Add(1)
    cPort, err := strconv.Atoi(cport)
    if err != nil {
      return ports, errors.New("Invalid internal port set")
    }
    nextPort := s.NextPort.Load()
    return []int{cPort, int(nextPort)}, nil
  }
}

