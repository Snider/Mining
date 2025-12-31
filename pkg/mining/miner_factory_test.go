package mining

import (
	"testing"
)

func TestMinerFactory_Create(t *testing.T) {
	factory := NewMinerFactory()

	tests := []struct {
		name      string
		minerType string
		wantErr   bool
	}{
		{"xmrig lowercase", "xmrig", false},
		{"xmrig uppercase", "XMRIG", false},
		{"xmrig mixed case", "XmRig", false},
		{"tt-miner", "tt-miner", false},
		{"ttminer alias", "ttminer", false},
		{"unknown type", "unknown", true},
		{"empty type", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			miner, err := factory.Create(tt.minerType)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Create(%q) expected error, got nil", tt.minerType)
				}
			} else {
				if err != nil {
					t.Errorf("Create(%q) unexpected error: %v", tt.minerType, err)
				}
				if miner == nil {
					t.Errorf("Create(%q) returned nil miner", tt.minerType)
				}
			}
		})
	}
}

func TestMinerFactory_IsSupported(t *testing.T) {
	factory := NewMinerFactory()

	tests := []struct {
		minerType string
		want      bool
	}{
		{"xmrig", true},
		{"tt-miner", true},
		{"ttminer", true}, // alias
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.minerType, func(t *testing.T) {
			if got := factory.IsSupported(tt.minerType); got != tt.want {
				t.Errorf("IsSupported(%q) = %v, want %v", tt.minerType, got, tt.want)
			}
		})
	}
}

func TestMinerFactory_ListTypes(t *testing.T) {
	factory := NewMinerFactory()

	types := factory.ListTypes()
	if len(types) < 2 {
		t.Errorf("ListTypes() returned %d types, expected at least 2", len(types))
	}

	// Check that expected types are present
	typeMap := make(map[string]bool)
	for _, typ := range types {
		typeMap[typ] = true
	}

	expectedTypes := []string{"xmrig", "tt-miner"}
	for _, expected := range expectedTypes {
		if !typeMap[expected] {
			t.Errorf("ListTypes() missing expected type %q", expected)
		}
	}
}

func TestMinerFactory_Register(t *testing.T) {
	factory := NewMinerFactory()

	// Register a custom miner type
	called := false
	factory.Register("custom-miner", func() Miner {
		called = true
		return NewXMRigMiner() // Return something valid for testing
	})

	if !factory.IsSupported("custom-miner") {
		t.Error("custom-miner should be supported after registration")
	}

	_, err := factory.Create("custom-miner")
	if err != nil {
		t.Errorf("Create custom-miner failed: %v", err)
	}
	if !called {
		t.Error("custom constructor was not called")
	}
}

func TestMinerFactory_RegisterAlias(t *testing.T) {
	factory := NewMinerFactory()

	// Register an alias for xmrig
	factory.RegisterAlias("x", "xmrig")

	if !factory.IsSupported("x") {
		t.Error("alias 'x' should be supported")
	}

	miner, err := factory.Create("x")
	if err != nil {
		t.Errorf("Create with alias failed: %v", err)
	}
	if miner == nil {
		t.Error("Create with alias returned nil miner")
	}
}

func TestGlobalFactory_CreateMiner(t *testing.T) {
	// Test global convenience functions
	miner, err := CreateMiner("xmrig")
	if err != nil {
		t.Errorf("CreateMiner failed: %v", err)
	}
	if miner == nil {
		t.Error("CreateMiner returned nil")
	}
}

func TestGlobalFactory_IsMinerSupported(t *testing.T) {
	if !IsMinerSupported("xmrig") {
		t.Error("xmrig should be supported")
	}
	if IsMinerSupported("nosuchminer") {
		t.Error("nosuchminer should not be supported")
	}
}

func TestGlobalFactory_ListMinerTypes(t *testing.T) {
	types := ListMinerTypes()
	if len(types) < 2 {
		t.Errorf("ListMinerTypes() returned %d types, expected at least 2", len(types))
	}
}
