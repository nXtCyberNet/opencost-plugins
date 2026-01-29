package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/go-plugin"
	"github.com/opencost/opencost-plugins/pkg/plugins/genai/genaiprovider"
	"github.com/opencost/opencost-plugins/pkg/plugins/genai/internal"
)

type GenAISource struct {
	Config *genaiprovider.Config
}

// GetCustomCosts is called by OpenCost to retrieve the enriched GenAI costs.
func (s *GenAISource) GetCustomCosts(req *genaiprovider.CustomCostRequest) ([]*genaiprovider.GenAIWorkloadData, error) {
	log.Printf("GenAI Plugin: Fetching costs for window %s - %s", req.Start, req.End)

	// 1. Call your internal math logic (from internal/join.go)
	// This performs the Prometheus fetch and the MIG-to-Node join.
	workloads, err := internal.CalculateGenAIWorkloads(req.Start, req.End, s.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate GenAI workloads: %w", err)
	}

	var responses []*genaiprovider.GenAIWorkloadData
	for _, w := range workloads {
		// Map internal results to the genaiprovider format
		responses = append(responses, &genaiprovider.GenAIWorkloadData{
			PodName:       w.PodName,
			ModelName:     w.ModelName,
			TotalTokens:   w.TotalTokens,
			TotalCost:     w.TotalCost,
			TenantID:      w.TenantID,
			WorkflowPhase: w.WorkflowPhase,
			MIGProfile:    w.MIGProfile,
			GPUEfficiency: w.GPUEfficiency,
		})
	}

	return responses, nil
}

func main() {
	// 1. Load plugin configuration (e.g., Prometheus URL, Cluster ID)
	configPath := os.Getenv("PLUGIN_CONFIG_PATH")
	if configPath == "" {
		configPath = "config/genai-config.json"
	}

	cfg, err := genaiprovider.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("GenAI Plugin: Failed to load config: %v", err)
	}

	// 2. Create the source instance
	source := &GenAISource{Config: cfg}

	// 3. Define the Handshake (Security "Cookie" required by go-plugin)
	var handshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "OPENCOST_PLUGIN_MAGIC_COOKIE",
		MagicCookieValue: "genai-visibility",
	}

	// 4. Start the gRPC Server
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"customcost": &genaiprovider.CustomCostPlugin{
				Impl: source,
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
