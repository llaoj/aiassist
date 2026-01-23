package sysinfo

import (
	"testing"
)

func TestLoadOrCollect(t *testing.T) {
	info, err := LoadOrCollect()
	if err != nil {
		t.Fatalf("LoadOrCollect failed: %v", err)
	}

	if info == nil {
		t.Fatal("Expected system info to not be nil")
	}

	// Verify basic fields are populated
	if info.OS == "" {
		t.Error("Expected OS to be populated")
	}

	if info.Arch == "" {
		t.Error("Expected Arch to be populated")
	}
}

func TestCollect(t *testing.T) {
	info, err := Collect()
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	if info == nil {
		t.Fatal("Expected collected info to not be nil")
	}

	// Basic fields should be populated
	if info.OS == "" {
		t.Error("Expected OS to be populated")
	}

	if info.Arch == "" {
		t.Error("Expected Arch to be populated")
	}
}

func TestSystemInfo_HasBasicFields(t *testing.T) {
	info, err := LoadOrCollect()
	if err != nil {
		t.Fatalf("LoadOrCollect failed: %v", err)
	}

	// Basic fields should be non-empty
	if info.OS == "" {
		t.Error("OS should not be empty")
	}

	if info.Arch == "" {
		t.Error("Arch should not be empty")
	}

	if info.User == "" {
		t.Error("User should not be empty")
	}

	if info.HomeDir == "" {
		t.Error("HomeDir should not be empty")
	}
}
