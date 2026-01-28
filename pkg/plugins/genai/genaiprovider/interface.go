package genaiprovider

import (
	"context"
	"time"
)

// CustomCostSource is the interface that your GenAI plugin must satisfy.
// This is the "Brain" that your main.go will implement.
type CustomCostSource interface {
	GetCustomCosts(req *CustomCostRequest) ([]*GenAIWorkloadData, error)
}

// GenAIWorkloadData represents a processed GenAI workload with cost information
type GenAIWorkloadData struct {
	PodName       string
	ModelName     string
	TotalTokens   int64
	TotalCost     float64
	TenantID      string
	WorkflowPhase string
	MIGProfile    string
	GPUEfficiency float64
}

// CustomCostRequest represents the time window OpenCost is asking for.
type CustomCostRequest struct {
	Start      time.Time
	End        time.Time
	Resolution string // e.g., "hourly" or "daily"
}

// --- gRPC Implementation ---

// GRPCServer handles incoming requests from OpenCost and sends them to your Plugin.
type GRPCServer struct {
	// This is a standard OpenCost requirement to avoid proto-version mismatch.
	UnimplementedCustomCostSourceServer
	Impl CustomCostSource
}

func (m *GRPCServer) GetCustomCosts(ctx context.Context, req *GetCustomCostsRequest) (*GetCustomCostsResponse, error) {
	// Translate the gRPC request into your Go struct
	r := &CustomCostRequest{
		Start: time.Unix(req.Start, 0),
		End:   time.Unix(req.End, 0),
	}

	// Call your math logic
	results, err := m.Impl.GetCustomCosts(r)
	if err != nil {
		return nil, err
	}

	// Translate results back to gRPC response
	var workloads []*GenAIWorkload
	for _, res := range results {
		workloads = append(workloads, &GenAIWorkload{
			PodName:       res.PodName,
			ModelName:     res.ModelName,
			TotalTokens:   res.TotalTokens,
			TotalCost:     res.TotalCost,
			TenantId:      res.TenantID,
			WorkflowPhase: res.WorkflowPhase,
			MigProfile:    res.MIGProfile,
			GpuEfficiency: res.GPUEfficiency,
		})
	}

	return &GetCustomCostsResponse{Workloads: workloads}, nil
}

// GRPCClient is what the OpenCost Core uses to talk to your Plugin.
type GRPCClient struct {
	client CustomCostSourceClient
}

func (m *GRPCClient) GetCustomCosts(req *CustomCostRequest) ([]*GenAIWorkloadData, error) {
	// Translate Go struct to gRPC
	resp, err := m.client.GetCustomCosts(context.Background(), &GetCustomCostsRequest{
		Start: req.Start.Unix(),
		End:   req.End.Unix(),
	})
	if err != nil {
		return nil, err
	}

	// Translate gRPC response back to Go slice
	var results []*GenAIWorkloadData
	for _, w := range resp.Workloads {
		results = append(results, &GenAIWorkloadData{
			PodName:       w.PodName,
			ModelName:     w.ModelName,
			TotalTokens:   w.TotalTokens,
			TotalCost:     w.TotalCost,
			TenantID:      w.TenantId,
			WorkflowPhase: w.WorkflowPhase,
			MIGProfile:    w.MigProfile,
			GPUEfficiency: w.GpuEfficiency,
		})
	}
	return results, nil
}
