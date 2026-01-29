package genaiplugin

type MetricMapping struct {
	InputTokens  string
	OutputTokens string
	GPUActiveSec string
	GPUUtil      string
}

func DefaultMetricMapping() MetricMapping {
	return MetricMapping{
		InputTokens:  "vllm:prompt_tokens_total",
		OutputTokens: "vllm:generation_tokens_total",
		GPUUtil:      "DCGM_FI_DEV_GPU_UTIL",
		GPUActiveSec: "DCGM_FI_DEV_GPU_UTIL",
	}
}

func GetMapping(annotations map[string]string) MetricMapping {

	// "fallback" if no annotations are present
	m := MetricMapping{
		InputTokens:  "llm_tokens_input_total",
		OutputTokens: "llm_tokens_output_total",
		GPUActiveSec: "llm_gpu_active_seconds_total",
		GPUUtil:      "llm_gpu_utilization_percent",
	}

	// Override only if specific annotations exist
	if val, ok := annotations["opencost.io/metric-input"]; ok {
		m.InputTokens = val
	}
	if val, ok := annotations["opencost.io/metric-output"]; ok {
		m.OutputTokens = val
	}
	if val, ok := annotations["opencost.io/metric-gpu-sec"]; ok {
		m.GPUActiveSec = val
	}
	if val, ok := annotations["opencost.io/metric-gpu-util"]; ok {
		m.GPUUtil = val
	}

	return m
}
