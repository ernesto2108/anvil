package mapper

import (
	"fmt"

	"github.com/ernesto2108/anvil/internal/dashboard/dto"
	"github.com/ernesto2108/anvil/internal/dashboard/entity"
)

// ToFlowDTO builds a FlowDTO from agent rows, constructing the DAG from
// depends_on fields with sequential fallback.
//
// Edge logic:
//  1. If the agent has non-empty depends_on, create one edge per dependency.
//  2. If depends_on is empty, use sequential fallback: the previous agent
//     in the slice connects to the current one.
//  3. Duplicate edges are prevented with a seen set.
func ToFlowDTO(rows []entity.AgentRow, parseDependsOn func(string) []string) dto.FlowDTO {
	nodes := make([]dto.FlowNodeDTO, 0, len(rows))
	edges := make([]dto.FlowEdgeDTO, 0, len(rows))
	edgeSeen := make(map[string]struct{}, len(rows))

	for _, r := range rows {
		node := dto.FlowNodeDTO{
			ID:   r.AgentID,
			Type: "agentNode",
			Data: dto.FlowNodeDataDTO{
				Label:      r.AgentRole,
				Status:     r.Status,
				DurationMs: r.DurationMs,
			},
		}
		nodes = append(nodes, node)
	}

	for i, r := range rows {
		deps := parseDependsOn(r.DependsOn)

		if len(deps) > 0 {
			for _, depID := range deps {
				key := depID + "->" + r.AgentID
				if _, already := edgeSeen[key]; already {
					continue
				}
				edgeSeen[key] = struct{}{}
				edges = append(edges, dto.FlowEdgeDTO{
					ID:     fmt.Sprintf("e-%s-%s", depID, r.AgentID),
					Source: depID,
					Target: r.AgentID,
				})
			}
		} else if i > 0 {
			prev := rows[i-1].AgentID
			key := prev + "->" + r.AgentID
			if _, already := edgeSeen[key]; !already {
				edgeSeen[key] = struct{}{}
				edges = append(edges, dto.FlowEdgeDTO{
					ID:     fmt.Sprintf("e-%s-%s", prev, r.AgentID),
					Source: prev,
					Target: r.AgentID,
				})
			}
		}
	}

	return dto.FlowDTO{Nodes: nodes, Edges: edges}
}
