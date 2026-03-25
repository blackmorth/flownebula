package internal

import (
	"encoding/json"
	"fmt"
	"math"
	"path"
	"strings"
)

// parseNodeName transforme le nom brut de la probe en identifiant Blackfire-compatible.
//
// Format probe : "filename::ClassName::method_name"
//
//	"{closure}::filename/line_start-line_end"
func parseNodeName(raw string) (nodeID, name, calledClass string) {
	// Closures émises directement par la probe avec le bon format
	if strings.HasPrefix(raw, "{closure}::") {
		return raw, "", ""
	}

	// Format "filename::ClassName::method_name"
	idx1 := strings.Index(raw, "::")
	if idx1 < 0 {
		return raw, "", ""
	}
	rest := raw[idx1+2:]
	idx2 := strings.Index(rest, "::")
	if idx2 < 0 {
		return raw, "", ""
	}

	filename := raw[:idx1]
	className := rest[:idx2]
	funcName := rest[idx2+2:]

	// Script-level (pas de nom de fonction)
	if funcName == "" || funcName == "unknown" {
		base := path.Base(filename)
		if base == "" || base == "." {
			base = filename
		}
		nodeID = "run_init::" + base
		return nodeID, "", ""
	}

	// Méthode d'une classe
	if className != "" {
		nodeID = className + "::" + funcName
		return nodeID, nodeID, className
	}

	// Fonction globale
	return funcName, "", ""
}

// round3 arrondit à 3 chiffres significatifs (comme Blackfire).
func round3(v float64) float64 {
	if v == 0 {
		return 0
	}
	mag := math.Pow(10, 3-math.Ceil(math.Log10(math.Abs(v))))
	return math.Round(v*mag) / mag
}

func (s *Session) ExportToDetailedJSON() ([]byte, error) {
	// Note: s.mu is already locked by s.Print() which calls this method
	if s.Root == nil {
		return nil, fmt.Errorf("session root is nil")
	}

	jsonNodes := make(map[string]JSONNode)
	jsonEdges := make(map[string]JSONEdge)
	edgesByKey := make(map[string]string) // "caller→callee" → edgeId
	edgeCount := 0

	rootCanonicalName, _, _ := parseNodeName(GetFuncName(s.Root.FuncID))

	var walkNode func(n *Node, parentName string, parentFuncID uint32)
	walkNode = func(n *Node, parentName string, parentFuncID uint32) {
		if n == nil {
			return
		}
		nodeName, name, calledClass := parseNodeName(GetFuncName(n.FuncID))
		if parentName == nodeName && n.FuncID == s.Root.FuncID {
			// Skip only the duplicated root self-edge node, but keep traversing
			// its children so the real call flow is not dropped.
			for _, child := range n.Children {
				walkNode(child, nodeName, n.FuncID)
			}
			return
		}
		// Récupère ou crée le nœud (fusion si même nom)
		jNode, exists := jsonNodes[nodeName]
		if !exists {
			jNode = JSONNode{
				NodeID:              nodeName,
				Name:                name,
				CalledClass:         calledClass,
				InclusiveCost:       map[string]int64{"ct": 0, "wt": 0, "cpu": 0, "mu": 0, "pmu": 0, "io": 0, "nw": 0, "nw_in": 0, "nw_out": 0},
				ExclusiveCost:       map[string]int64{"ct": 0, "wt": 0, "cpu": 0, "mu": 0, "pmu": 0, "io": 0, "nw": 0, "nw_in": 0, "nw_out": 0},
				InclusivePercentage: make(map[string]float64),
				ExclusivePercentage: make(map[string]float64),
			}
			// Tag métier
			switch {
			case strings.Contains(nodeName, "PDO") || strings.Contains(strings.ToLower(nodeName), "sql"):
				jNode.Metrics = []string{"sql"}
			case strings.Contains(strings.ToLower(nodeName), "curl") || strings.Contains(strings.ToLower(nodeName), "socket"):
				jNode.Metrics = []string{"nw"}
			case strings.Contains(strings.ToLower(nodeName), "fopen") || strings.Contains(strings.ToLower(nodeName), "fread") || strings.Contains(strings.ToLower(nodeName), "fwrite"):
				jNode.Metrics = []string{"io"}
			default:
				base := strings.SplitN(nodeName, "::", 2)[0]
				jNode.Metrics = []string{base}
			}
		}

		// Accumulation des métriques
		for k, v := range n.Metrics {
			switch k {
			case "ewt":
				jNode.ExclusiveCost["wt"] += v
			case "ct":
				jNode.InclusiveCost["ct"] += v
				// exclusive ct : 0 (non calculé séparément)
			default:
				jNode.InclusiveCost[k] += v
			}
		}
		jsonNodes[nodeName] = jNode

		// Arête vers ce nœud depuis son parent.
		// Keep recursive self-edges (foo -> foo), they carry recursion call counts.
		// The duplicated root self-edge is already filtered above.
		if parentName != "" {
			// Ignore only direct synthetic request-entry self edge (same rendered
			// name as root but different func IDs), while keeping true recursion.
			isSyntheticRootSelfEdge := parentFuncID == s.Root.FuncID &&
				n.FuncID != s.Root.FuncID &&
				parentName == nodeName &&
				parentName == rootCanonicalName
			if !isSyntheticRootSelfEdge {
				edgeKey := parentName + "→" + nodeName
				if existingEdgeID, ok := edgesByKey[edgeKey]; ok {
					// Fusion : additionner les coûts de l'arête
					e := jsonEdges[existingEdgeID]
					for k, v := range n.Metrics {
						if k != "ewt" {
							e.Cost[k] += v
						}
					}
					jsonEdges[existingEdgeID] = e
				} else {
					edgeCount++
					edgeID := fmt.Sprintf("e%d", edgeCount)
					edgesByKey[edgeKey] = edgeID

					cost := map[string]int64{"ct": 0, "wt": 0, "cpu": 0, "mu": 0, "pmu": 0, "io": 0, "nw": 0, "nw_in": 0, "nw_out": 0}
					for k, v := range n.Metrics {
						if k != "ewt" {
							cost[k] = v
						}
					}
					jsonEdges[edgeID] = JSONEdge{
						EdgeID: edgeID,
						Caller: parentName,
						Callee: nodeName,
						Cost:   cost,
					}
				}
			}
		}

		// Récursion sur les enfants
		for _, child := range n.Children {
			walkNode(child, nodeName, n.FuncID)
		}
	}

	// Point de départ : si le root synthétique n'a pas de métriques et a un seul enfant,
	// on promeut cet enfant comme root réel.
	startNode := s.Root
	startName := ""
	walkNode(startNode, startName, 0)

	if len(jsonNodes) == 0 {
		return nil, fmt.Errorf("no nodes to export")
	}

	// Le nom du nœud racine réel
	rootName, _, _ := parseNodeName(GetFuncName(startNode.FuncID))

	// Baseline 100% = coûts inclusifs du root
	totalIncl := make(map[string]int64)
	if rootNode, ok := jsonNodes[rootName]; ok {
		for k, v := range rootNode.InclusiveCost {
			totalIncl[k] = v
		}
	}

	// Calcul des pourcentages et des pics
	peaks := Peaks{
		Inclusive: make(map[string]int64),
		Exclusive: make(map[string]int64),
	}

	// ct est un compteur : pas de pourcentage
	noPct := map[string]bool{"ct": true}

	for id, node := range jsonNodes {
		for k, v := range node.InclusiveCost {
			if !noPct[k] {
				if total, ok := totalIncl[k]; ok && total > 0 {
					node.InclusivePercentage[k] = round3(float64(v) * 100 / float64(total))
				}
			}
			if v > peaks.Inclusive[k] {
				peaks.Inclusive[k] = v
			}
		}
		for k, v := range node.ExclusiveCost {
			if !noPct[k] {
				if total, ok := totalIncl[k]; ok && total > 0 {
					node.ExclusivePercentage[k] = round3(float64(v) * 100 / float64(total))
				}
			}
			if v > peaks.Exclusive[k] {
				peaks.Exclusive[k] = v
			}
		}
		jsonNodes[id] = node
	}

	res := DetailedJSON{
		AgentSessionID: fmt.Sprintf("%016x", s.ID),
		Dimensions: map[string]Dimension{
			"ct":  {"ct", "Calls", false},
			"wt":  {"wt", "Wall Time", true},
			"cpu": {"cpu", "CPU", true},
			"mu":  {"mu", "Instant memory", false},
			"pmu": {"pmu", "Memory", true},
			"io":  {"io", "I/O Wait", true},
			"nw":  {"nw", "Network", true},
		},
		Root:       rootName,
		Nodes:      jsonNodes,
		Edges:      jsonEdges,
		Comparison: false,
		Peaks:      peaks,
		Language:   "php",
	}

	return json.MarshalIndent(res, "", "  ")
}
