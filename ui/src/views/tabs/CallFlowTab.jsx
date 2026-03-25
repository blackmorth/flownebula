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
        <Box
            w="100%"
            h="80vh"
            overflow="auto"
            borderWidth="1px"
            borderRadius="md"
            bg="white"
            p={2}
        >
            <div ref={ref} style={{ minWidth: "100%", minHeight: "100%" }} />
        </Box>
    );
}

function buildDot(payload) {
    const { nodes = {}, edges = {}, root } = payload;

    const edgeList = Object.values(edges);
    const maxCt = Math.max(1, ...edgeList.map((e) => e.cost?.ct || 0));
    const maxWt = Math.max(1, ...edgeList.map((e) => e.cost?.wt || 0));

    let dot = `
digraph CallFlow {
    graph [
        rankdir=LR,
        splines=true,
        overlap=false,
        nodesep=0.55,
        ranksep=0.8,
        pad=0.2,
        bgcolor="#ffffff"
    ];

    node [
        shape=box,
        style="rounded,filled",
        penwidth=1.1,
        fontname="Inter",
        fontsize=12,
        margin="0.18,0.10",
        fillcolor="#f8fafc",
        color="#334155",
        fontcolor="#0f172a"
    ];

    edge [
        fontname="Inter",
        fontsize=10,
        color="#64748b",
        fontcolor="#334155",
        arrowsize=0.75,
        labeldistance=1.5,
        labelfloat=false
    ];
`;

    // --- NODES ---
    Object.values(nodes).forEach((n) => {
        const isRoot = n.nodeId === root;
        dot += `    "${escapeDot(n.nodeId)}" [label="${escapeDot(n.nodeId)}"${
            isRoot ? ', shape=doubleoctagon, fillcolor="#e2e8f0", color="#1e293b", penwidth=1.6' : ""
        }];\n`;
    });

    dot += "\n";

    // --- EDGES ---
    edgeList.forEach((e) => {
        const ct = e.cost?.ct || 0;
        const wt = e.cost?.wt || 0;

        const label = `${ct} call${ct > 1 ? "s" : ""}\\n${formatMs(wt)}`;

        // Scale thickness by call count and darkness by wall time.
        const penwidth = 1.2 + (ct / maxCt) * 4.2;
        const stroke = interpolateHex("#94a3b8", "#1e293b", wt / maxWt);

        dot += `    "${escapeDot(e.caller)}" -> "${escapeDot(e.callee)}" [label="${label}", penwidth=${penwidth.toFixed(
            2
        )}, color="${stroke}"];\n`;
    });

    dot += "}\n";

    return dot;
}

function formatMs(us) {
    return new Intl.NumberFormat("en-US", {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
    }).format(us / 1000) + "ms";
}

function escapeDot(value = "") {
    return String(value).replace(/\\/g, "\\\\").replace(/\"/g, '\\\"');
}

function interpolateHex(start, end, t) {
    const clamp = Math.max(0, Math.min(1, Number.isFinite(t) ? t : 0));
    const from = hexToRgb(start);
    const to = hexToRgb(end);

    const rgb = {
        r: Math.round(from.r + (to.r - from.r) * clamp),
        g: Math.round(from.g + (to.g - from.g) * clamp),
        b: Math.round(from.b + (to.b - from.b) * clamp),
    };

    return rgbToHex(rgb);
}

function hexToRgb(hex) {
    const clean = hex.replace("#", "");
    const padded = clean.length === 3 ? clean.split("").map((c) => c + c).join("") : clean;
    const int = Number.parseInt(padded, 16);

    return {
        r: (int >> 16) & 255,
        g: (int >> 8) & 255,
        b: int & 255,
    };
}

function rgbToHex({ r, g, b }) {
    return `#${[r, g, b]
        .map((v) => Math.max(0, Math.min(255, v)).toString(16).padStart(2, "0"))
        .join("")}`;
}
