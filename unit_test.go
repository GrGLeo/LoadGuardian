package main

import (
	"testing"
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

func TestPullImage(t *testing.T) {
  config, _ := ParseYAML("service.yml")
  err := PullServices(&config)
  if err != nil {
    t.Errorf("Expected image pulling to work, got %s", err.Error())
  }
}
