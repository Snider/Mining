package mining

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// MockMiner is a mock implementation of the Miner interface for testing.
type MockMiner struct {
	InstallFunc               func() error
	UninstallFunc             func() error
	StartFunc                 func(config *Config) error
	StopFunc                  func() error
	GetStatsFunc              func() (*PerformanceMetrics, error)
	GetNameFunc               func() string
	GetPathFunc               func() string
	GetBinaryPathFunc         func() string
	CheckInstallationFunc     func() (*InstallationDetails, error)
	GetLatestVersionFunc      func() (string, error)
	GetHashrateHistoryFunc    func() []HashratePoint
	AddHashratePointFunc      func(point HashratePoint)
	ReduceHashrateHistoryFunc func(now time.Time)
	GetLogsFunc               func() []string
}

func (m *MockMiner) Install() error                         { return m.InstallFunc() }
func (m *MockMiner) Uninstall() error                       { return m.UninstallFunc() }
func (m *MockMiner) Start(config *Config) error             { return m.StartFunc(config) }
func (m *MockMiner) Stop() error                            { return m.StopFunc() }
func (m *MockMiner) GetStats() (*PerformanceMetrics, error) { return m.GetStatsFunc() }
func (m *MockMiner) GetName() string                        { return m.GetNameFunc() }
func (m *MockMiner) GetPath() string                        { return m.GetPathFunc() }
func (m *MockMiner) GetBinaryPath() string                  { return m.GetBinaryPathFunc() }
func (m *MockMiner) CheckInstallation() (*InstallationDetails, error) {
	return m.CheckInstallationFunc()
}
func (m *MockMiner) GetLatestVersion() (string, error)    { return m.GetLatestVersionFunc() }
func (m *MockMiner) GetHashrateHistory() []HashratePoint  { return m.GetHashrateHistoryFunc() }
func (m *MockMiner) AddHashratePoint(point HashratePoint) { m.AddHashratePointFunc(point) }
func (m *MockMiner) ReduceHashrateHistory(now time.Time)  { m.ReduceHashrateHistoryFunc(now) }
func (m *MockMiner) GetLogs() []string                    { return m.GetLogsFunc() }

// MockManager is a mock implementation of the Manager for testing.
type MockManager struct {
	ListMinersFunc              func() []Miner
	ListAvailableMinersFunc     func() []AvailableMiner
	StartMinerFunc              func(minerType string, config *Config) (Miner, error)
	StopMinerFunc               func(minerName string) error
	GetMinerFunc                func(minerName string) (Miner, error)
	GetMinerHashrateHistoryFunc func(minerName string) ([]HashratePoint, error)
	UninstallMinerFunc          func(minerType string) error
	StopFunc                    func()
}

func (m *MockManager) ListMiners() []Miner                   { return m.ListMinersFunc() }
func (m *MockManager) ListAvailableMiners() []AvailableMiner { return m.ListAvailableMinersFunc() }
func (m *MockManager) StartMiner(minerType string, config *Config) (Miner, error) {
	return m.StartMinerFunc(minerType, config)
}
func (m *MockManager) StopMiner(minerName string) error { return m.StopMinerFunc(minerName) }
func (m *MockManager) GetMiner(minerName string) (Miner, error) {
	return m.GetMinerFunc(minerName)
}
func (m *MockManager) GetMinerHashrateHistory(minerName string) ([]HashratePoint, error) {
	return m.GetMinerHashrateHistoryFunc(minerName)
}
func (m *MockManager) UninstallMiner(minerType string) error { return m.UninstallMinerFunc(minerType) }
func (m *MockManager) Stop()                                 { m.StopFunc() }

var _ ManagerInterface = (*MockManager)(nil)

func setupTestRouter() (*gin.Engine, *MockManager) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	mockManager := &MockManager{
		ListMinersFunc:              func() []Miner { return []Miner{} },
		ListAvailableMinersFunc:     func() []AvailableMiner { return []AvailableMiner{} },
		StartMinerFunc:              func(minerType string, config *Config) (Miner, error) { return nil, nil },
		StopMinerFunc:               func(minerName string) error { return nil },
		GetMinerFunc:                func(minerName string) (Miner, error) { return nil, nil },
		GetMinerHashrateHistoryFunc: func(minerName string) ([]HashratePoint, error) { return nil, nil },
		UninstallMinerFunc:          func(minerType string) error { return nil },
		StopFunc:                    func() {},
	}
	service := &Service{
		Manager:       mockManager,
		Router:        router,
		APIBasePath:   "/",
		SwaggerUIPath: "/swagger",
	}
	service.SetupRoutes()
	return router, mockManager
}

func TestHandleListMiners(t *testing.T) {
	router, mockManager := setupTestRouter()
	mockManager.ListMinersFunc = func() []Miner {
		return []Miner{&XMRigMiner{BaseMiner: BaseMiner{Name: "test-miner"}}}
	}

	req, _ := http.NewRequest("GET", "/miners", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleGetInfo(t *testing.T) {
	router, _ := setupTestRouter()

	// Case 1: Successful response
	req, _ := http.NewRequest("GET", "/info", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleDoctor(t *testing.T) {
	router, mockManager := setupTestRouter()
	mockManager.ListAvailableMinersFunc = func() []AvailableMiner {
		return []AvailableMiner{{Name: "xmrig"}}
	}

	// Case 1: Successful response
	req, _ := http.NewRequest("POST", "/doctor", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleInstallMiner(t *testing.T) {
	router, _ := setupTestRouter()

	// Test installing a miner
	req, _ := http.NewRequest("POST", "/miners/xmrig/install", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Installation endpoint should be accessible
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 200 or 500, got %d", w.Code)
	}
}

func TestHandleStopMiner(t *testing.T) {
	router, mockManager := setupTestRouter()
	mockManager.StopMinerFunc = func(minerName string) error {
		return nil
	}

	req, _ := http.NewRequest("DELETE", "/miners/test-miner", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleGetMinerStats(t *testing.T) {
	router, mockManager := setupTestRouter()
	mockManager.GetMinerFunc = func(minerName string) (Miner, error) {
		return &MockMiner{
			GetStatsFunc: func() (*PerformanceMetrics, error) {
				return &PerformanceMetrics{Hashrate: 100}, nil
			},
			GetLogsFunc: func() []string { return []string{} },
		}, nil
	}

	req, _ := http.NewRequest("GET", "/miners/test-miner/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleGetMinerHashrateHistory(t *testing.T) {
	router, mockManager := setupTestRouter()
	mockManager.GetMinerHashrateHistoryFunc = func(minerName string) ([]HashratePoint, error) {
		return []HashratePoint{{Timestamp: time.Now(), Hashrate: 100}}, nil
	}

	req, _ := http.NewRequest("GET", "/miners/test-miner/hashrate-history", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
