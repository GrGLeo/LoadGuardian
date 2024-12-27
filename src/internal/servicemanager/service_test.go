package servicemanager_test

import (
	"testing"

	"github.com/GrGLeo/LoadBalancer/src/internal/servicemanager"
)

func TestPortInitializing(t *testing.T) {
	t.Run("Initial Port Assignment", func(t *testing.T) {
		service := servicemanager.Service{
			Image: "hello",
			Port:  []string{"8080:8080"},
		}

		nextPort, err := service.GetPort()
		expectedPort := []int{8080, 8080}

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(nextPort) != len(expectedPort) {
			t.Fatalf("Expected %d ports, got %d", len(expectedPort), len(nextPort))
		}

		for i := range nextPort {
			if nextPort[i] != expectedPort[i] {
				t.Errorf("Incorrect port mapping at index %d: got %d, want %d", i, nextPort[i], expectedPort[i])
			}
		}
	})

	t.Run("Incremented Port Assignment", func(t *testing.T) {
		service := servicemanager.Service{
			Image:    "hello",
			Port:     []string{"8080:8080"},
		}
    service.NextPort.Store(8080)


		nextPort, err := service.GetPort()
		expectedPort := []int{8080, 8081}

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(nextPort) != len(expectedPort) {
			t.Fatalf("Expected %d ports, got %d", len(expectedPort), len(nextPort))
		}

		for i := range nextPort {
			if nextPort[i] != expectedPort[i] {
				t.Errorf("Incorrect port mapping at index %d: got %d, want %d", i, nextPort[i], expectedPort[i])
			}
		}
	})
  
  t.Run("Invalid Port setting", func(t *testing.T) {
    service := servicemanager.Service{
      Image: "hello",
      Port: []string{"hello:false"},
    }
    _, err := service.GetPort()
    if err == nil {
      t.Fatalf("Expected error, got nil")
    }
  })
}
