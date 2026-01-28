package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

type Scraper struct {
	Client  *http.Client
	Timeout time.Duration
}

func NewScraper(timeout time.Duration) *Scraper {
	return &Scraper{
		Client:  &http.Client{Timeout: timeout},
		Timeout: timeout,
	}
}

func (s *Scraper) fetchRawMetrics(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("scrape failed: status %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// Updated to use dto.MetricFamily
func (s *Scraper) parseToFamilies(data io.Reader) (map[string]*dto.MetricFamily, error) {
	var parser expfmt.TextParser
	return parser.TextToMetricFamilies(data)
}

// Updated mapping logic for the DTO structs
func (s *Scraper) extractMetricValue(families map[string]*dto.MetricFamily, name string) float64 {
	family, ok := families[name]
	if !ok || family == nil {
		return 0
	}

	var sum float64
	for _, m := range family.Metric {
		// Prometheus DTO uses pointers for values
		if m.Counter != nil && m.Counter.Value != nil {
			sum += *m.Counter.Value
		} else if m.Gauge != nil && m.Gauge.Value != nil {
			sum += *m.Gauge.Value
		} else if m.Untyped != nil && m.Untyped.Value != nil {
			sum += *m.Untyped.Value
		}
	}
	return sum
}

func (s *Scraper) Scrape(ctx context.Context, ip, port string, mapping MetricMapping) (*Metrics, error) {
	url := fmt.Sprintf("http://%s:%s/metrics", ip, port)

	raw, err := s.fetchRawMetrics(ctx, url)
	if err != nil {
		return nil, err
	}
	defer raw.Close()

	families, err := s.parseToFamilies(raw)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		InputTokens:    s.extractMetricValue(families, mapping.InputTokens),
		OutputTokens:   s.extractMetricValue(families, mapping.OutputTokens),
		GPUActiveSec:   s.extractMetricValue(families, mapping.GPUActiveSec),
		GPUUtilPercent: s.extractMetricValue(families, mapping.GPUUtilization),
	}, nil
}
