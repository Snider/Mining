package mining

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchJSONStats(t *testing.T) {
	t.Run("SuccessfulFetch", func(t *testing.T) {
		// Create a test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/test/endpoint" {
				t.Errorf("Unexpected path: %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"value": 42,
				"name":  "test",
			})
		}))
		defer server.Close()

		// Get port from server listener
		addr := server.Listener.Addr().(*net.TCPAddr)

		config := HTTPStatsConfig{
			Host:     "127.0.0.1",
			Port:     addr.Port,
			Endpoint: "/test/endpoint",
		}

		var result struct {
			Value int    `json:"value"`
			Name  string `json:"name"`
		}

		ctx := context.Background()
		err := FetchJSONStats(ctx, config, &result)
		if err != nil {
			t.Fatalf("FetchJSONStats failed: %v", err)
		}

		if result.Value != 42 {
			t.Errorf("Expected value 42, got %d", result.Value)
		}
		if result.Name != "test" {
			t.Errorf("Expected name 'test', got '%s'", result.Name)
		}
	})

	t.Run("ZeroPort", func(t *testing.T) {
		config := HTTPStatsConfig{
			Host:     "localhost",
			Port:     0,
			Endpoint: "/test",
		}

		var result map[string]interface{}
		err := FetchJSONStats(context.Background(), config, &result)
		if err == nil {
			t.Error("Expected error for zero port")
		}
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		config := HTTPStatsConfig{
			Host:     "127.0.0.1",
			Port:     12345, // Intentionally wrong port to trigger connection timeout
			Endpoint: "/test",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		var result map[string]interface{}
		err := FetchJSONStats(ctx, config, &result)
		if err == nil {
			t.Error("Expected error for cancelled context")
		}
	})
}

func TestMinerTypeRegistry(t *testing.T) {
	t.Run("KnownTypes", func(t *testing.T) {
		if !IsMinerSupported(MinerTypeXMRig) {
			t.Error("xmrig should be a known miner type")
		}
		if !IsMinerSupported(MinerTypeTTMiner) {
			t.Error("tt-miner should be a known miner type")
		}
		if !IsMinerSupported(MinerTypeSimulated) {
			t.Error("simulated should be a known miner type")
		}
	})

	t.Run("UnknownType", func(t *testing.T) {
		if IsMinerSupported("unknown-miner") {
			t.Error("unknown-miner should not be a known miner type")
		}
	})

	t.Run("ListMinerTypes", func(t *testing.T) {
		types := ListMinerTypes()
		if len(types) == 0 {
			t.Error("ListMinerTypes should return registered types")
		}
	})
}

func TestGetType(t *testing.T) {
	t.Run("XMRigMiner", func(t *testing.T) {
		miner := NewXMRigMiner()
		if miner.GetType() != MinerTypeXMRig {
			t.Errorf("Expected type %s, got %s", MinerTypeXMRig, miner.GetType())
		}
	})

	t.Run("TTMiner", func(t *testing.T) {
		miner := NewTTMiner()
		if miner.GetType() != MinerTypeTTMiner {
			t.Errorf("Expected type %s, got %s", MinerTypeTTMiner, miner.GetType())
		}
	})

	t.Run("SimulatedMiner", func(t *testing.T) {
		miner := NewSimulatedMiner(SimulatedMinerConfig{
			Name:         "test-sim",
			Algorithm:    "rx/0",
			BaseHashrate: 1000,
		})
		if miner.GetType() != MinerTypeSimulated {
			t.Errorf("Expected type %s, got %s", MinerTypeSimulated, miner.GetType())
		}
	})
}
