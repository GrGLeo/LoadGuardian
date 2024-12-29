package config_test

import (
	"sync/atomic"
	"testing"

	"github.com/GrGLeo/LoadBalancer/src/internal/config"
	servicemanager "github.com/GrGLeo/LoadBalancer/src/internal/servicemanager"
	"github.com/GrGLeo/LoadBalancer/src/pkg/utils"
)

func TestCompareConfig(t *testing.T) {
	tests := []struct {
		name          string
		oldConfig     config.Config
		newConfig     config.Config
		expectedDiff  config.ConfigDiff
	}{
		{
			name:      "Empty configs",
			oldConfig: config.Config{},
			newConfig: config.Config{},
			expectedDiff: config.ConfigDiff{
				AddedService:   map[string]servicemanager.Service{},
				RemovedService: map[string]servicemanager.Service{},
				UpdatedService: map[string]servicemanager.Service{},
			},
		},
		{
			name: "Service added",
			oldConfig: config.Config{},
			newConfig: config.Config{
				Service: map[string]servicemanager.Service{
					"new-service": {
						Image: "new-image",
						Port:  []string{"8080"},
					},
				},
			},
			expectedDiff: config.ConfigDiff{
				AddedService: map[string]servicemanager.Service{
					"new-service": {
						Image: "new-image",
						Port:  []string{"8080"},
					},
				},
				RemovedService: map[string]servicemanager.Service{},
				UpdatedService: map[string]servicemanager.Service{},
			},
		},
		{
			name: "Service removed",
			oldConfig: config.Config{
				Service: map[string]servicemanager.Service{
					"old-service": {
						Image: "old-image",
						Port:  []string{"9090"},
					},
				},
			},
			newConfig: config.Config{},
			expectedDiff: config.ConfigDiff{
				AddedService: map[string]servicemanager.Service{},
				RemovedService: map[string]servicemanager.Service{
					"old-service": {
						Image: "old-image",
						Port:  []string{"9090"},
					},
				},
				UpdatedService: map[string]servicemanager.Service{},
			},
		},
		{
			name: "Service updated",
			oldConfig: config.Config{
				Service: map[string]servicemanager.Service{
					"existing-service": {
						Image: "old-image",
						Port:  []string{"8080"},
            NextPort: &atomic.Uint32{},
					},
				},
			},
			newConfig: config.Config{
				Service: map[string]servicemanager.Service{
					"existing-service": {
						Image: "new-image",
						Port:  []string{"8080", "9090"},
            NextPort: &atomic.Uint32{},
					},
				},
			},
			expectedDiff: config.ConfigDiff{
				AddedService:   map[string]servicemanager.Service{},
				RemovedService: map[string]servicemanager.Service{},
				UpdatedService: map[string]servicemanager.Service{
					"existing-service": {
						Image: "new-image",
						Port:  []string{"8080", "9090"},
            NextPort: &atomic.Uint32{},
					},
				},
			},
		},
		{
			name: "Multiple changes",
			oldConfig: config.Config{
				Service: map[string]servicemanager.Service{
					"service-A": {Image: "image-A"},
					"service-B": {Image: "image-B", NextPort: &atomic.Uint32{}},
				},
			},
			newConfig: config.Config{
				Service: map[string]servicemanager.Service{
					"service-B": {Image: "updated-image-B", NextPort: &atomic.Uint32{}},
					"service-C": {Image: "image-C"},
				},
			},
			expectedDiff: config.ConfigDiff{
				AddedService: map[string]servicemanager.Service{
					"service-C": {Image: "image-C"},
				},
				RemovedService: map[string]servicemanager.Service{
					"service-A": {Image: "image-A"},
				},
				UpdatedService: map[string]servicemanager.Service{
					"service-B": {Image: "updated-image-B", NextPort: &atomic.Uint32{}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff, err := tt.oldConfig.CompareConfig(tt.newConfig)
			if err != nil {
				t.Fatalf("CompareConfig returned error: %v", err)
			}

			// Assert AddedService
			for name, expected := range tt.expectedDiff.AddedService {
				actual, ok := diff.AddedService[name]
				if !ok {
					t.Errorf("Expected service %s in AddedService, but not found", name)
					continue
				}
				if actual.Image != expected.Image || utils.CompareStrings(true, actual.Port, expected.Port) {
					t.Errorf("Unexpected AddedService[%s]. Got: %+v, Want: %+v", name, actual, expected)
				}
			}

			// Assert RemovedService
			for name, expected := range tt.expectedDiff.RemovedService {
				actual, ok := diff.RemovedService[name]
				if !ok {
					t.Errorf("Expected service %s in RemovedService, but not found", name)
					continue
				}
				if actual.Image != expected.Image || utils.CompareStrings(true, actual.Port, expected.Port) {
					t.Errorf("Unexpected RemovedService[%s]. Got: %+v, Want: %+v", name, actual, expected)
				}
			}

			// Assert UpdatedService
			for name, expected := range tt.expectedDiff.UpdatedService {
				actual, ok := diff.UpdatedService[name]
				if !ok {
					t.Errorf("Expected service %s in UpdatedService, but not found", name)
					continue
				}
				if actual.Image != expected.Image || utils.CompareStrings(true, actual.Port, expected.Port) {
					t.Errorf("Unexpected UpdatedService[%s]. Got: %+v, Want: %+v", name, actual, expected)
				}
			}
		})
	}
}
