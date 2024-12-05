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

