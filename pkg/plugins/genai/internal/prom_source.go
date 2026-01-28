package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/common/model"
)

// This allows decoupling from the prometheus-source module
type PrometheusQuerier interface {
	Query(query string) (interface{}, error)
	QueryWithContext(ctx context.Context, query string) (interface{}, error)
}

type PrometheusProvider struct {
	querier PrometheusQuerier
}

func NewPrometheusProvider(querier PrometheusQuerier) *PrometheusProvider {
	return &PrometheusProvider{querier: querier}
}

// 1. Logic for Counter vs Gauge PromQL
func (p *PrometheusProvider) buildQuery(metric string, window time.Duration, isCounter bool) string {
	windowStr := fmt.Sprintf("%.0fs", window.Seconds())
	if isCounter {
		// Delta increase for counters (tokens, gpu active seconds)
		return fmt.Sprintf("sum(increase(%s[%s])) by (pod, namespace)", metric, windowStr)
	}
	// Average for gauges (gpu utilization)
	return fmt.Sprintf("avg(avg_over_time(%s[%s])) by (pod, namespace)", metric, windowStr)
}

// 2. Simple execution wrapper
func (p *PrometheusProvider) execute(ctx context.Context, query string) (model.Vector, error) {
	// Re-using the existing OpenCost Query interface
	val, err := p.querier.Query(query)
	if err != nil {
		return nil, err
	}

	vector, ok := val.(model.Vector)
	if !ok {
		return nil, fmt.Errorf("unexpected prometheus result type")
	}
	return vector, nil
}

// 3. Mapping logic for standardizing output
func (p *PrometheusProvider) mapResults(vector model.Vector, results map[string]*Metrics, key string) {
	for _, sample := range vector {
		pod := string(sample.Metric["pod"])
		ns := string(sample.Metric["namespace"])
		if pod == "" || ns == "" {
			continue
		}

		id := fmt.Sprintf("%s/%s", ns, pod)
		if _, ok := results[id]; !ok {
			results[id] = &Metrics{}
		}

		val := float64(sample.Value)
		switch key {
		case "input":
			results[id].InputTokens = val
		case "output":
			results[id].OutputTokens = val
		case "gpuSec":
			results[id].GPUActiveSec = val
		case "util":
			results[id].GPUUtilPercent = val
		}
	}
}

// 4. Batch orchestrator for the 4-query sequence
func (p *PrometheusProvider) Fetch(ctx context.Context, start, end time.Time, mapping MetricMapping) (map[string]*Metrics, error) {
	window := end.Sub(start)
	results := make(map[string]*Metrics)

	steps := []struct {
		id      string
		metric  string
		counter bool
	}{
		{"input", mapping.InputTokens, true},
		{"output", mapping.OutputTokens, true},
		{"gpuSec", mapping.GPUActiveSec, true},
		{"util", mapping.GPUUtilization, false},
	}

	for _, s := range steps {
		q := p.buildQuery(s.metric, window, s.counter)
		v, err := p.execute(ctx, q)
		if err != nil {
			return nil, fmt.Errorf("fetch error for %s: %w", s.id, err)
		}
		p.mapResults(v, results, s.id)
	}

	return results, nil
}
