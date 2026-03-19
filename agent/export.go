package agent

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (s *Session) ExportToDetailedJSON() ([]byte, error) {
	// Note: s.mu is already locked by s.Print() which calls this method
	if s.Root == nil {
		return nil, fmt.Errorf("session root is nil")
	}

	jsonNodes := make(map[string]JSONNode)
	jsonEdges := make(map[string]JSONEdge)
	edgeCount := 0

	var exportNode func(n *Node) string
	exportNode = func(n *Node) string {
		if n == nil {
			return ""
		}
		nodeName := GetFuncName(n.FuncID)
		nodeID := fmt.Sprintf("%d", n.ID)

		jNode := JSONNode{
			NodeID:              nodeName,
			InclusiveCost:       make(map[string]int64),
			ExclusiveCost:       make(map[string]int64),
			InclusivePercentage: make(map[string]float64),
			ExclusivePercentage: make(map[string]float64),
		}

		// Identify metrics/tags for the node
		if strings.Contains(nodeName, "PDO") || strings.Contains(nodeName, "sql") {
			jNode.Metrics = append(jNode.Metrics, "sql")
		} else if strings.Contains(nodeName, "curl") || strings.Contains(nodeName, "socket") {
			jNode.Metrics = append(jNode.Metrics, "nw")
		} else if strings.Contains(nodeName, "file") || strings.Contains(nodeName, "fopen") || strings.Contains(nodeName, "fread") || strings.Contains(nodeName, "fwrite") {
			jNode.Metrics = append(jNode.Metrics, "io")
		}
		// Add default metric if none
		if len(jNode.Metrics) == 0 {
			parts := strings.Split(nodeName, "(")
			if len(parts) > 0 {
				jNode.Metrics = append(jNode.Metrics, parts[0])
			}
		}

		// Accumulate metrics from this node
		for k, v := range n.Metrics {
			if k == "ewt" {
				jNode.ExclusiveCost["wt"] += v
			} else {
				jNode.InclusiveCost[k] += v
			}
		}

		jsonNodes[nodeID] = jNode

		for _, childNode := range n.Children {
			childID := exportNode(childNode)
			if childID != "" {
				edgeCount++
				edgeID := fmt.Sprintf("e%d", edgeCount)

				// Calculate edge cost (cost of this specific call)
				edgeCost := make(map[string]int64)
				for k, v := range childNode.Metrics {
					if k != "ewt" {
						edgeCost[k] = v
					}
				}

				jsonEdges[edgeID] = JSONEdge{
					EdgeID: edgeID,
					Caller: nodeID,
					Callee: childID,
					Cost:   edgeCost,
				}
			}
		}
		return nodeID
	}

	rootID := exportNode(s.Root)

	if len(jsonNodes) == 0 {
		return nil, fmt.Errorf("no nodes to export")
	}

	// Calculate total costs from root to ensure percentages can be computed correctly
	// The root inclusive cost should be the reference (100%)
	totalInclusiveCosts := make(map[string]int64)
	if rootNode, ok := jsonNodes[rootID]; ok {
		for k, v := range rootNode.InclusiveCost {
			totalInclusiveCosts[k] = v
		}
	}

	// Calculate percentages and peaks
	peaks := Peaks{
		Inclusive: make(map[string]int64),
		Exclusive: make(map[string]int64),
	}

	for id, node := range jsonNodes {
		for k, v := range node.InclusiveCost {
			if total, ok := totalInclusiveCosts[k]; ok && total > 0 {
				node.InclusivePercentage[k] = float64(v) * 100 / float64(total)
			}
			if v > peaks.Inclusive[k] {
				peaks.Inclusive[k] = v
			}
		}
		for k, v := range node.ExclusiveCost {
			if total, ok := totalInclusiveCosts[k]; ok && total > 0 {
				node.ExclusivePercentage[k] = float64(v) * 100 / float64(total)
			}
			if v > peaks.Exclusive[k] {
				peaks.Exclusive[k] = v
			}
		}
		jsonNodes[id] = node
	}

	res := DetailedJSON{
		Dimensions: map[string]Dimension{
			"ct":  {"ct", "Calls", false},
			"wt":  {"wt", "Wall Time", true},
			"cpu": {"cpu", "CPU", true},
			"mu":  {"mu", "Memory Delta", false},
			"pmu": {"pmu", "Peak Memory", true},
			"io":  {"io", "I/O", true},
			"nw":  {"nw", "Network", true},
		},
		Root:       rootID,
		Nodes:      jsonNodes,
		Edges:      jsonEdges,
		Comparison: false,
		Peaks:      peaks,
		Language:   "php",
	}

	return json.MarshalIndent(res, "", "  ")
}
