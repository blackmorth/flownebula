// CallFlowTab.jsx
import { useEffect, useRef } from "react";
import { Box } from "@chakra-ui/react";
import { graphviz } from "d3-graphviz";

export default function CallFlowTab({ payload }) {
    const ref = useRef(null);

    useEffect(() => {
        if (!payload || !payload.edges || !ref.current) return;

        const dot = buildDot(payload);

        graphviz(ref.current)
            .zoom(true)
            .fit(true)
            .renderDot(dot);
    }, [payload]);

    return (
        <Box w="100%" h="80vh" overflow="hidden" borderWidth="1px" borderRadius="md">
            <div ref={ref} style={{ width: "100%", height: "100%" }} />
        </Box>
    );
}

function buildDot(payload) {
    const { nodes, edges, root } = payload;

    let dot = `
digraph CallFlow {
    rankdir=LR;
    splines=true;
    nodesep=0.4;
    ranksep=0.6;

    node [
        shape=box,
        style="rounded,filled",
        fontname="Inter",
        fontsize=12,
        fillcolor="#f7fafc",
        color="#333333"
    ];

    edge [
        fontname="Inter",
        fontsize=11,
        color="#555555",
        arrowsize=0.7
    ];
`;

    // --- NODES ---
    Object.values(nodes).forEach((n) => {
        const isRoot = n.nodeId === root;
        dot += `    "${n.nodeId}" [label="${n.nodeId}"${
            isRoot ? ', shape=doubleoctagon, fillcolor="#e2e8f0"' : ""
        }];\n`;
    });

    dot += "\n";

    // --- EDGES ---
    Object.values(edges).forEach((e) => {
        const ct = e.cost?.ct || 0;
        const wt = e.cost?.wt || 0;

        const label = `${ct} call${ct > 1 ? "s" : ""}\\n${formatMs(wt)}`;

        // penwidth proportionnel au nombre d’appels
        const pen = Math.max(1, Math.min(6, 1 + ct));

        dot += `    "${e.caller}" -> "${e.callee}" [label="${label}", penwidth=${pen}];\n`;
    });

    dot += "}\n";

    return dot;
}

function formatMs(us) {
    return (us / 1000).toFixed(2) + "ms";
}
