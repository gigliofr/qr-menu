package main

import (
	"context"
	"testing"
	"time"

	"qr-menu/pkg/config"
	"qr-menu/pkg/container"
)

// TestNewServiceContainer tests container creation
func TestNewServiceContainer(t *testing.T) {
	cfg := config.Load()
	
	cont, err := container.NewServiceContainer(cfg)
	if err != nil {
		t.Fatalf("Failed to create service container: %v", err)
	}

	if cont == nil {
		t.Fatal("Service container is nil")
	}

	if !cont.IsInitialized() {
		t.Fatal("Service container should be initialized")
	}

	t.Log("✅ Service container created successfully")
}

// TestServiceContainerNilConfig tests container with nil config
func TestServiceContainerNilConfig(t *testing.T) {
	cont, err := container.NewServiceContainer(nil)

	if err == nil {
		t.Error("Expected error with nil config")
	}

	if cont != nil {
		t.Error("Container should be nil with nil config")
	}

	t.Log("✅ Nil config properly rejected")
}

// TestServiceContainerGetters tests all getter methods
func TestServiceContainerGetters(t *testing.T) {
	cfg := config.Load()
	cont, err := container.NewServiceContainer(cfg)
	if err != nil {
		t.Fatalf("Failed to create service container: %v", err)
	}

	tests := []struct {
		name   string
		getter func() interface{}
	}{
		{"Config", func() interface{} { return cont.Config() }},
		{"Analytics", func() interface{} { return cont.Analytics() }},
		{"Backup", func() interface{} { return cont.Backup() }},
		{"Notifications", func() interface{} { return cont.Notifications() }},
		{"Localization", func() interface{} { return cont.Localization() }},
		{"PWA", func() interface{} { return cont.PWA() }},
		{"Database", func() interface{} { return cont.Database() }},
		{"Migration", func() interface{} { return cont.Migration() }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = tt.getter()
			// All services should either be initialized or nil (for optional ones)
			// Just verify the getter doesn't panic
			t.Logf("✅ %s getter works", tt.name)
		})
	}
}

// TestServiceContainerHealth tests the health check
func TestServiceContainerHealth(t *testing.T) {
	cfg := config.Load()
	cont, err := container.NewServiceContainer(cfg)
	if err != nil {
		t.Fatalf("Failed to create service container: %v", err)
	}

	health := cont.Health()
	
	if health == nil {
		t.Fatal("Health should not be nil")
	}

	initialized, ok := health["initialized"].(bool)
	if !ok || !initialized {
		t.Error("Container should be marked as initialized")
	}

	services, ok := health["services"].(map[string]bool)
	if !ok {
		t.Error("Health should contain services map")
	}

	if len(services) == 0 {
		t.Error("Services map should not be empty")
	}

	t.Logf("✅ Health check passed with %d services", len(services))
}

// TestServiceContainerShutdown tests graceful shutdown
func TestServiceContainerShutdown(t *testing.T) {
	cfg := config.Load()
	cont, err := container.NewServiceContainer(cfg)
	if err != nil {
		t.Fatalf("Failed to create service container: %v", err)
	}

	if !cont.IsInitialized() {
		t.Fatal("Container should be initialized before shutdown")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := cont.Shutdown(ctx); err != nil {
		t.Errorf("Unexpected error during shutdown: %v", err)
	}

	if cont.IsInitialized() {
		t.Error("Container should not be initialized after shutdown")
	}

	t.Log("✅ Shutdown completed successfully")
}

// TestServiceContainerConcurrentAccess tests thread-safety
func TestServiceContainerConcurrentAccess(t *testing.T) {
	cfg := config.Load()
	cont, err := container.NewServiceContainer(cfg)
	if err != nil {
		t.Fatalf("Failed to create service container: %v", err)
	}

	done := make(chan bool, 10)

	// Concurrent getter operations
	for i := 0; i < 10; i++ {
		go func() {
			_ = cont.Config()
			_ = cont.Analytics()
			_ = cont.IsInitialized()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	t.Log("✅ Concurrent access test passed")
}

// TestServiceContainerConfigIntegration tests container with config
func TestServiceContainerConfigIntegration(t *testing.T) {
	cfg := config.Load()
	cont, err := container.NewServiceContainer(cfg)
	if err != nil {
		t.Fatalf("Failed to create service container: %v", err)
	}

	retrievedCfg := cont.Config()
	if retrievedCfg == nil {
		t.Fatal("Config should not be nil")
	}

	if retrievedCfg.Server.Port == 0 {
		t.Error("Config port should be set")
	}

	if retrievedCfg.Database.Engine == "" {
		t.Error("Config database engine should be set")
	}

	t.Log("✅ Config integration test passed")
}


