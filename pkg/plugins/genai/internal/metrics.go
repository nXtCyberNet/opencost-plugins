package internal

type Metrics struct {
	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`

	InputTokens  float64 `json:"inputTokens"`
	OutputTokens float64 `json:"outputTokens"`
	CachedTokens float64 `json:"cachedTokens"`

	// Compute Efficiency
	GPUActiveSec   float64 `json:"gpuActiveSec"`
	GPUUtilPercent float64 `json:"gpuUtilPercent"`

	// Metadata
	ModelName     string `json:"modelName"`
	WorkflowPhase string `json:"workflowPhase"`

	// Performance KPIs (The "Value" signal)
	AvgTTFTms    float64            `json:"avgTTFTms"`
	QualityScore float64            `json:"qualityScore"`
	PodRequests  map[string]float64 `json:"podRequests"`
}

func (m *Metrics) Sanitize() {
	if m.InputTokens < 0 {
		m.InputTokens = 0
	}
	if m.OutputTokens < 0 {
		m.OutputTokens = 0
	}
	if m.GPUUtilPercent < 0 {
		m.GPUUtilPercent = 0
	}
	if m.GPUUtilPercent > 100 {
		m.GPUUtilPercent = 100
	}
	if m.GPUActiveSec < 0 {
		m.GPUActiveSec = 0
	}
}

type MetricMapping struct {
	// Prometheus Metric Names
	InputTokens    string `json:"inputTokensMetric"`
	OutputTokens   string `json:"outputTokensMetric"`
	GPUUtilization string `json:"gpuUtilizationMetric"`
	GPUActiveSec   string `json:"gpuActiveSec"`

	// Label Overrides (to handle different Prometheus exporters)
	PodLabel       string `json:"podLabel"`
	NamespaceLabel string `json:"namespaceLabel"`
	NodeLabel      string `json:"nodeLabel"`

	// GenAI Specific Labels (from your inference server)
	ModelLabel    string `json:"modelLabel"`
	TenantLabel   string `json:"tenantLabel"`
	WorkflowLabel string `json:"workflowLabel"`
}

// DefaultMetricMapping provides a sane baseline for most K8s environments.
func DefaultMetricMapping() MetricMapping {
	return MetricMapping{
		InputTokens:    "genai_input_tokens_total",
		OutputTokens:   "genai_output_tokens_total",
		GPUUtilization: "container_gpu_utilization",
		PodLabel:       "pod",
		NamespaceLabel: "namespace",
		NodeLabel:      "node",
		ModelLabel:     "model",
		TenantLabel:    "tenant",
		WorkflowLabel:  "workflow",
	}
}
