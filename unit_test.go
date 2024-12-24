package main

import (
	"testing"

	"github.com/docker/docker/client"
)

func TestRoundRobin(t *testing.T) {
	services := []BackendService{
		{ID: "1"}, {ID: "2"}, {ID: "3"},
	}
	lb := LoadBalancer{
		Services: services,
		index:    0,
	}

	expected := []string{"2", "3", "1", "2"}
	for _, exp := range expected {
		backend := lb.getNextBackend()
		if backend.ID != exp {
			t.Errorf("Expected %s, got %s", exp, backend.ID)
		}
	}
}

func TestLeastConnection(t *testing.T) {
	services := []BackendService{
		{ID: "1", Connection: 5},
		{ID: "2", Connection: 3},
		{ID: "3", Connection: 8},
	}
	lb := LoadBalancer{
		Services: services,
	}

	backend := lb.getLeastConnection()
	if backend.ID != "2" {
		t.Errorf("Expected '2', got %s", backend.ID)
	}
}

func TestRandomBackend(t *testing.T) {
	services := []BackendService{
		{ID: "1"}, {ID: "2"}, {ID: "3"},
	}
	lb := LoadBalancer{
		Services: services,
	}

	backend := lb.getRandomBackend()
	if backend == nil {
		t.Errorf("Expected a backend, got nil")
	}
}

func TestYamlParsing(t *testing.T) {
  config, err := ParseYAML("service.yml")
	if err != nil {
		t.Fatalf("parse returned an error: %v", err)
	}

	if len(config.Service) != 2 {
		t.Errorf("Expected 2 services, got %d", len(config.Service))
	}
  if len(config.Network) != 1 {
		t.Errorf("Expected 1 services, got %d", len(config.Network))
	}
  if config.Network["app-network"].Driver != "bridge" {
		t.Errorf("Expected Network bridge to be 'bridge', got '%s'", config.Network["app-network"].Driver)
	}
	if config.Service["Backend"].Image != "hello-world" {
		t.Errorf("Expected Backend image to be 'hello-world', got '%s'", config.Service["Backend"].Image)
	}
	if config.Service["Backend"].Network[0] != "app-network" {
		t.Errorf("Expected Backend network to be 'app-network', got '%s'", config.Service["Backend"].Network)
	}
	if config.Service["Frontend"].Image != "hello-world" {
		t.Errorf("Expected Frontend image to be 'hello-world', got '%s'", config.Service["Frontend"].Image)
	}
	if len(config.Service["Frontend"].Network) != 0 {
		t.Errorf("Expected Frontend network to be empty, got '%s'", config.Service["Frontend"].Network)
  }
}

func TestCreateNetwork(t *testing.T) {
  config, _ := ParseYAML("service.yml")
  cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
  err := config.CreateNetworks(cli)
  if err != nil {
    t.Errorf("Expected network creation to work, got %s", err.Error())
  }
}
  
    

func TestPullImage(t *testing.T) {
  config, _ := ParseYAML("service.yml")
  err := config.PullServices()
  if err != nil {
    t.Errorf("Expected image pulling to work, got %s", err.Error())
  }
}

func TestCreateAndStartContainer(t *testing.T) {
  config, err := ParseYAML("service.yml")
  if err != nil {
    t.Fatalf("Failed to parse YAML: %s", err.Error())
  }

  cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
  if err != nil {
    t.Fatalf("Failed to create Docker client: %s", err.Error())
  }

  s := config.Service["Backend"]

  // Test container creation
  id, err := s.CreateService(cli)
  if err != nil {
    t.Fatalf("Expected container creation to succeed, got error: %s", err.Error())
  }
  t.Logf("Container created successfully with ID: %s", id)

  // Test container starting
  err = config.ServiceStart(cli, id)
  if err != nil {
    t.Fatalf("Expected container to start successfully, got error: %s", err.Error())
  }
  t.Log("Container started successfully")
}

