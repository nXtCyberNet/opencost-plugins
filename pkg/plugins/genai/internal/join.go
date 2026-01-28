package internal

import (
	"strings"
	"time"

	"github.com/opencost/opencost-plugins/pkg/plugins/genai/genaiprovider"
)

type NodeCapacityMap map[string]float64

type GenAIReport struct {
	Attributes GenAIAttributes   `json:"attributes"`
	Stats      EfficiencyMetrics `json:"stats"`
}

// based on  OpenTelemetry semantic conventions for GenAI.
type GenAIAttributes struct {
	WorkflowPhase string `json:"workflow_phase"`
	ModelName     string `json:"model_name"`
	ModelVersion  string `json:"model_version"`
	TenantID      string `json:"tenant_id"`
	Accelerator   string `json:"accelerator"`
	GPUUUID       string `json:"gpu_uuid"`
}

const (
	AnnotationWorkflowPhase = "opencost.io/workflow-phase"
	AnnotationModelName     = "opencost.io/gen-ai-model-name"
	AnnotationModelVersion  = "opencost.io/gen-ai-model-version"
	AnnotationTenantID      = "opencost.io/tenant-id"
	AnnotationAccelerator   = "opencost.io/accelerator-type"
	AnnotationGPUUUID       = "opencost.io/gpu-uuid"
)

func JoinPodToNodeMetrics(m *Metrics, nodeCap NodeCapacityMap, podCost float64, attr GenAIAttributes) GenAIReport {

	if m == nil {
		return GenAIReport{
			Attributes: attr,
			Stats:      EfficiencyMetrics{},
		}
	}

	profile, requested := FindMIGProfile(m.PodRequests)

	var infra *MIGEfficiency
	if profile != "" {
		if capacity, ok := nodeCap[profile]; ok && capacity > 0 {
			infra = &MIGEfficiency{
				MIGUtilization: (requested / capacity) * 100,
			}
		}
	}

	stats := CalculateEfficiency(m, infra, nodeCap, podCost)

	return GenAIReport{
		Attributes: attr,
		Stats:      stats,
	}
}

// ExtractGenAIAttributes pulls key metadata from Pod annotations.
func ExtractGenAIAttributes(annotations map[string]string) GenAIAttributes {
	return GenAIAttributes{
		WorkflowPhase: annotations[AnnotationWorkflowPhase],
		ModelName:     annotations[AnnotationModelName],
		ModelVersion:  annotations[AnnotationModelVersion],
		TenantID:      annotations[AnnotationTenantID],
		Accelerator:   annotations[AnnotationAccelerator],
		GPUUUID:       annotations[AnnotationGPUUUID],
	}
}

func GetNodeMIGCapacity(capacity map[string]float64) NodeCapacityMap {
	migCap := make(NodeCapacityMap)
	for resource, quantity := range capacity {
		// Only grab nvidia.com/mig-* resources
		if strings.HasPrefix(resource, "nvidia.com/mig-") {
			migCap[resource] = quantity
		}
	}
	return migCap
}

// GenAIWorkload represents a processed GenAI workload with cost information
type GenAIWorkload struct {
	PodName       string
	ModelName     string
	TotalTokens   int64
	TotalCost     float64
	TenantID      string
	WorkflowPhase string
	MIGProfile    string
	GPUEfficiency float64
}

// CalculateGenAIWorkloads fetches metrics and calculates costs for GenAI workloads
func CalculateGenAIWorkloads(start, end time.Time, config *genaiprovider.Config) ([]GenAIWorkload, error) {
	// TODO: Implement the actual Prometheus fetching and MIG-to-Node join logic
	// This is a placeholder implementation to satisfy the interface

	workloads := []GenAIWorkload{
		{
			PodName:       "genai-inference-pod",
			ModelName:     "gpt-4",
			TotalTokens:   1000,
			TotalCost:     0.02,
			TenantID:      "tenant-1",
			WorkflowPhase: "inference",
			MIGProfile:    "1g.5gb",
			GPUEfficiency: 85.5,
		},
		{
			PodName:       "genai-training-pod",
			ModelName:     "llama-2",
			TotalTokens:   5000,
			TotalCost:     0.15,
			TenantID:      "tenant-2",
			WorkflowPhase: "training",
			MIGProfile:    "3g.20gb",
			GPUEfficiency: 92.3,
		},
	}

	return workloads, nil
}
