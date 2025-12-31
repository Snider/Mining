package mining

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics provides simple instrumentation counters for the mining package.
// These can be exposed via Prometheus or other metrics systems in the future.
type Metrics struct {
	// API metrics
	RequestsTotal   atomic.Int64
	RequestsErrored atomic.Int64
	RequestLatency  *LatencyHistogram

	// Miner metrics
	MinersStarted atomic.Int64
	MinersStopped atomic.Int64
	MinersErrored atomic.Int64

	// Stats collection metrics
	StatsCollected atomic.Int64
	StatsRetried   atomic.Int64
	StatsFailed    atomic.Int64

	// WebSocket metrics
	WSConnections atomic.Int64
	WSMessages    atomic.Int64

	// P2P metrics
	P2PMessagesSent     atomic.Int64
	P2PMessagesReceived atomic.Int64
	P2PConnectionsTotal atomic.Int64
}

// LatencyHistogram tracks request latencies with basic percentile support.
type LatencyHistogram struct {
	mu      sync.Mutex
	samples []time.Duration
	maxSize int
}

// NewLatencyHistogram creates a new latency histogram with a maximum sample size.
func NewLatencyHistogram(maxSize int) *LatencyHistogram {
	return &LatencyHistogram{
		samples: make([]time.Duration, 0, maxSize),
		maxSize: maxSize,
	}
}

// Record adds a latency sample.
func (h *LatencyHistogram) Record(d time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.samples) >= h.maxSize {
		// Ring buffer behavior - overwrite oldest
		copy(h.samples, h.samples[1:])
		h.samples = h.samples[:len(h.samples)-1]
	}
	h.samples = append(h.samples, d)
}

// Average returns the average latency.
func (h *LatencyHistogram) Average() time.Duration {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.samples) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range h.samples {
		total += d
	}
	return total / time.Duration(len(h.samples))
}

// Count returns the number of samples.
func (h *LatencyHistogram) Count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.samples)
}

// DefaultMetrics is the global metrics instance.
var DefaultMetrics = &Metrics{
	RequestLatency: NewLatencyHistogram(1000),
}

// RecordRequest records an API request.
func RecordRequest(errored bool, latency time.Duration) {
	DefaultMetrics.RequestsTotal.Add(1)
	if errored {
		DefaultMetrics.RequestsErrored.Add(1)
	}
	DefaultMetrics.RequestLatency.Record(latency)
}

// RecordMinerStart records a miner start event.
func RecordMinerStart() {
	DefaultMetrics.MinersStarted.Add(1)
}

// RecordMinerStop records a miner stop event.
func RecordMinerStop() {
	DefaultMetrics.MinersStopped.Add(1)
}

// RecordMinerError records a miner error event.
func RecordMinerError() {
	DefaultMetrics.MinersErrored.Add(1)
}

// RecordStatsCollection records a stats collection event.
func RecordStatsCollection(retried bool, failed bool) {
	DefaultMetrics.StatsCollected.Add(1)
	if retried {
		DefaultMetrics.StatsRetried.Add(1)
	}
	if failed {
		DefaultMetrics.StatsFailed.Add(1)
	}
}

// RecordWSConnection increments or decrements WebSocket connection count.
func RecordWSConnection(connected bool) {
	if connected {
		DefaultMetrics.WSConnections.Add(1)
	} else {
		DefaultMetrics.WSConnections.Add(-1)
	}
}

// RecordWSMessage records a WebSocket message.
func RecordWSMessage() {
	DefaultMetrics.WSMessages.Add(1)
}

// RecordP2PMessage records a P2P message.
func RecordP2PMessage(sent bool) {
	if sent {
		DefaultMetrics.P2PMessagesSent.Add(1)
	} else {
		DefaultMetrics.P2PMessagesReceived.Add(1)
	}
}

// GetMetricsSnapshot returns a snapshot of current metrics.
func GetMetricsSnapshot() map[string]interface{} {
	return map[string]interface{}{
		"requests_total":          DefaultMetrics.RequestsTotal.Load(),
		"requests_errored":        DefaultMetrics.RequestsErrored.Load(),
		"request_latency_avg_ms":  DefaultMetrics.RequestLatency.Average().Milliseconds(),
		"request_latency_samples": DefaultMetrics.RequestLatency.Count(),
		"miners_started":          DefaultMetrics.MinersStarted.Load(),
		"miners_stopped":          DefaultMetrics.MinersStopped.Load(),
		"miners_errored":          DefaultMetrics.MinersErrored.Load(),
		"stats_collected":         DefaultMetrics.StatsCollected.Load(),
		"stats_retried":           DefaultMetrics.StatsRetried.Load(),
		"stats_failed":            DefaultMetrics.StatsFailed.Load(),
		"ws_connections":          DefaultMetrics.WSConnections.Load(),
		"ws_messages":             DefaultMetrics.WSMessages.Load(),
		"p2p_messages_sent":       DefaultMetrics.P2PMessagesSent.Load(),
		"p2p_messages_received":   DefaultMetrics.P2PMessagesReceived.Load(),
	}
}
