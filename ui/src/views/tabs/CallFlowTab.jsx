import { useEffect, useMemo, useRef, useState } from "react";
import { Box, Flex, Table, Text } from "@chakra-ui/react";
import { graphviz } from "d3-graphviz";

export default function CallFlowTab({ payload }) {
    const ref = useRef(null);
    const [expandedRows, setExpandedRows] = useState(() => new Set());
    const [selectedNodeId, setSelectedNodeId] = useState(null);

    const rows = useMemo(() => buildRows(payload), [payload]);
    const relations = useMemo(() => buildRelations(payload), [payload]);

    useEffect(() => {
        if (!rows.length || !selectedNodeId) {
            setSelectedNodeId(rows[0]?.id ?? null);
        }
    }, [rows, selectedNodeId]);

    useEffect(() => {
        if (!payload?.edges || !ref.current || !rows.length) return;

        const dot = buildDot(payload, selectedNodeId);

        graphviz(ref.current)
            .zoom(true)
            .fit(true)
            .renderDot(dot)
            .on("end", () => {
                highlightSelectedSvgNode(ref.current, selectedNodeId);
            });
    }, [payload, selectedNodeId, rows]);

    const toggleExpanded = (nodeId) => {
        setExpandedRows((prev) => {
            const next = new Set(prev);
            if (next.has(nodeId)) {
                next.delete(nodeId);
            } else {
                next.add(nodeId);
            }
            return next;
        });
    };

    const selectNode = (nodeId) => {
        setSelectedNodeId(nodeId);
        setExpandedRows((prev) => {
            if (prev.has(nodeId)) return prev;
            const next = new Set(prev);
            next.add(nodeId);
            return next;
        });
    };

    return (
        <Flex gap={3} h="80vh" align="stretch">
            <Box
                flex="0 0 44%"
                borderWidth="1px"
                borderColor="border"
                borderRadius="xl"
                bg="bg.panel"
                boxShadow="glow"
                overflow="auto"
                p={2}
            >
                <Table.Root size="sm" variant="outline">
                    <Table.Header position="sticky" top={0} bg="bg.subtle" zIndex={1}>
                        <Table.Row>
                            <Table.ColumnHeader w="34px" />
                            <Table.ColumnHeader>Node</Table.ColumnHeader>
                            <Table.ColumnHeader>Wall Time</Table.ColumnHeader>
                            <Table.ColumnHeader>CPU</Table.ColumnHeader>
                            <Table.ColumnHeader textAlign="right">Calls</Table.ColumnHeader>
                        </Table.Row>
                    </Table.Header>
                    <Table.Body>
                        {rows.map((row) => {
                            const expanded = expandedRows.has(row.id);
                            const selected = selectedNodeId === row.id;
                            const relation = relations[row.id] || { callers: [], callees: [], callsIn: 0, callsOut: 0 };

                            return (
                                <FragmentRow
                                    key={row.id}
                                    expanded={expanded}
                                    selected={selected}
                                    row={row}
                                    relation={relation}
                                    onToggle={() => toggleExpanded(row.id)}
                                    onSelect={() => selectNode(row.id)}
                                />
                            );
                        })}
                    </Table.Body>
                </Table.Root>
            </Box>

            <Box
                flex="1"
                borderWidth="1px"
                borderColor="border"
                borderRadius="xl"
                bg="bg.panel"
                boxShadow="glow"
                overflow="auto"
                p={2}
            >
                <div ref={ref} style={{ minWidth: "100%", minHeight: "100%" }} />
            </Box>
        </Flex>
    );
}

function FragmentRow({ row, relation, expanded, selected, onToggle, onSelect }) {
    return (
        <>
            <Table.Row
                className="function-overview tableinfo-focus"
                bg={selected ? "nebula.100" : "transparent"}
                color={selected ? "nebula.500" : "text"}
                _hover={{ bg: "bg.subtle", cursor: "pointer" }}
                onClick={onSelect}
            >
                <Table.Cell onClick={(event) => {
                    event.stopPropagation();
                    onToggle();
                }}>
                    <Text fontWeight="bold">{expanded ? "−" : "+"}</Text>
                </Table.Cell>
                <Table.Cell title={row.title}>
                    <Text className="bf-ellipsis-left" overflow="hidden" textOverflow="ellipsis" whiteSpace="nowrap">
                        {row.label}
                    </Text>
                </Table.Cell>
                <Table.Cell>
                    <MetricBar primary={row.wtPct} secondary={row.cpuPct} value={formatUs(row.wt)} />
                </Table.Cell>
                <Table.Cell>{formatUs(row.cpu)}</Table.Cell>
                <Table.Cell textAlign="right">{formatCount(row.calls)}</Table.Cell>
            </Table.Row>

            {expanded && (
                <Table.Row className="function-info" name={row.id}>
                    <Table.Cell colSpan={5}>
                        <Box p={3} borderRadius="lg" bg="bg.subtle" borderWidth="1px" borderColor="border">
                            <Text fontWeight="semibold" mb={2}>{row.label}</Text>

                            <Flex gap={4} wrap="wrap" mb={3}>
                                <SummaryBadge title={`Callers (${relation.callers.length})`} value={`${formatCount(relation.callsIn)} calls`} />
                                <SummaryBadge title={`Callees (${relation.callees.length})`} value={`${formatCount(relation.callsOut)} calls`} />
                                <SummaryBadge title="Wall Time" value={formatUs(row.wt)} />
                                <SummaryBadge title="CPU" value={formatUs(row.cpu)} />
                                <SummaryBadge title="Memory" value={formatBytes(row.memory)} />
                            </Flex>

                            <Table.Root size="xs" variant="line">
                                <Table.Header>
                                    <Table.Row>
                                        <Table.ColumnHeader>Metric</Table.ColumnHeader>
                                        <Table.ColumnHeader>Value</Table.ColumnHeader>
                                        <Table.ColumnHeader>Exclusive</Table.ColumnHeader>
                                    </Table.Row>
                                </Table.Header>
                                <Table.Body>
                                    <Table.Row>
                                        <Table.Cell>Wall Time</Table.Cell>
                                        <Table.Cell>{formatUs(row.wt)}</Table.Cell>
                                        <Table.Cell>{formatUs(row.exclusiveWt)}</Table.Cell>
                                    </Table.Row>
                                    <Table.Row>
                                        <Table.Cell>CPU</Table.Cell>
                                        <Table.Cell>{formatUs(row.cpu)}</Table.Cell>
                                        <Table.Cell>{formatUs(row.exclusiveCpu)}</Table.Cell>
                                    </Table.Row>
                                    <Table.Row>
                                        <Table.Cell>I/O Wait</Table.Cell>
                                        <Table.Cell>{formatUs(row.ioWait)}</Table.Cell>
                                        <Table.Cell>{formatUs(row.exclusiveIoWait)}</Table.Cell>
                                    </Table.Row>
                                    <Table.Row>
                                        <Table.Cell>Memory</Table.Cell>
                                        <Table.Cell>{formatBytes(row.memory)}</Table.Cell>
                                        <Table.Cell>{formatBytes(row.exclusiveMemory)}</Table.Cell>
                                    </Table.Row>
                                </Table.Body>
                            </Table.Root>
                        </Box>
                    </Table.Cell>
                </Table.Row>
            )}
        </>
    );
}

function SummaryBadge({ title, value }) {
    return (
        <Box borderWidth="1px" borderColor="border" borderRadius="md" px={3} py={2} minW="140px" bg="bg.panel">
            <Text fontSize="xs" color="text.muted">{title}</Text>
            <Text fontSize="sm" fontWeight="semibold">{value}</Text>
        </Box>
    );
}

function MetricBar({ primary, secondary, value }) {
    return (
        <Box>
            <Text>{value}</Text>
            <Box className="bf-progress bf-progress-dim-wt" position="relative" h="6px" borderRadius="sm" overflow="hidden" bg="bg.subtle" mt={1}>
                <Box className="bf-progress-bar bf-progress-bar-secondary" position="absolute" inset={0} w={`${secondary}%`} bg="nebula.200" />
                <Box className="bf-progress-bar bf-progress-bar-primary" position="absolute" inset={0} w={`${primary}%`} bg="nebula.400" />
            </Box>
        </Box>
    );
}

function buildRows(payload) {
    const nodes = Object.values(payload?.nodes || {});
    const edges = Object.values(payload?.edges || {});

    if (!nodes.length) return [];

    const callsByNode = {};
    edges.forEach((edge) => {
        const ct = edge.cost?.ct || 0;
        callsByNode[edge.callee] = (callsByNode[edge.callee] || 0) + ct;
    });

    const maxWt = Math.max(1, ...nodes.map((node) => node.inclusive_cost?.wt || 0));
    const maxCpu = Math.max(1, ...nodes.map((node) => node.inclusive_cost?.cpu || 0));

    return nodes
        .map((node) => {
            const label = node.name || node.nodeId;
            return {
                id: node.nodeId,
                label,
                title: node.nodeId,
                wt: node.inclusive_cost?.wt || 0,
                cpu: node.inclusive_cost?.cpu || 0,
                ioWait: node.inclusive_cost?.io || 0,
                memory: node.inclusive_cost?.mu || node.inclusive_cost?.pmu || 0,
                exclusiveWt: node.exclusive_cost?.wt || 0,
                exclusiveCpu: node.exclusive_cost?.cpu || 0,
                exclusiveIoWait: node.exclusive_cost?.io || 0,
                exclusiveMemory: node.exclusive_cost?.mu || node.exclusive_cost?.pmu || 0,
                calls: callsByNode[node.nodeId] || 0,
                wtPct: ((node.inclusive_cost?.wt || 0) / maxWt) * 100,
                cpuPct: ((node.inclusive_cost?.cpu || 0) / maxCpu) * 100,
                pct: node.inclusive_percentage?.wt || 0,
            };
        })
        .sort((a, b) => b.wt - a.wt);
}

function buildRelations(payload) {
    const relation = {};

    Object.values(payload?.edges || {}).forEach((edge) => {
        relation[edge.caller] ??= { callers: [], callees: [], callsIn: 0, callsOut: 0 };
        relation[edge.callee] ??= { callers: [], callees: [], callsIn: 0, callsOut: 0 };

        relation[edge.caller].callees.push(edge.callee);
        relation[edge.caller].callsOut += edge.cost?.ct || 0;

        relation[edge.callee].callers.push(edge.caller);
        relation[edge.callee].callsIn += edge.cost?.ct || 0;
    });

    Object.values(relation).forEach((item) => {
        item.callers = Array.from(new Set(item.callers));
        item.callees = Array.from(new Set(item.callees));
    });

    return relation;
}

function buildDot(payload, selectedNodeId) {
    const { nodes = {}, edges = {}, root } = payload;

    const edgeList = Object.values(edges);
    const maxCt = Math.max(1, ...edgeList.map((e) => e.cost?.ct || 0));
    const maxWt = Math.max(1, ...edgeList.map((e) => e.cost?.wt || 0));

    let dot = `
digraph CallFlow {
    graph [rankdir=LR, splines=true, overlap=false, nodesep=0.45, ranksep=0.7, pad=0.2, bgcolor="#ffffff"];
    node [shape=box, style="rounded,filled", penwidth=1.1, fontname="Inter", fontsize=11, margin="0.18,0.10", fillcolor="#f8fafc", color="#334155", fontcolor="#0f172a", class="profile-node"];
    edge [fontname="Inter", fontsize=10, color="#64748b", fontcolor="#334155", arrowsize=0.75, labeldistance=1.5, labelfloat=false];
`;

    Object.values(nodes).forEach((n) => {
        const isRoot = n.nodeId === root;
        const isSelected = n.nodeId === selectedNodeId;
        const wtPct = n.inclusive_percentage?.wt || 0;
        const label = `${shortName(n.name || n.nodeId)}\\n${shortFile(n.nodeId)}\\n${wtPct.toFixed(2)}%`;

        dot += `    "${escapeDot(n.nodeId)}" [label="${escapeDot(label)}"${
            isRoot ? ', shape=doubleoctagon, fillcolor="#e2e8f0", color="#1e293b", penwidth=1.6' : ""
        }${isSelected ? ', class="selected-node active-node"' : ""}];\n`;
    });

    dot += "\n";

    edgeList.forEach((e) => {
        const ct = e.cost?.ct || 0;
        const wt = e.cost?.wt || 0;

        const label = `${formatCount(ct)} call${ct > 1 ? "s" : ""}\\n${formatUs(wt)}`;
        const penwidth = 1.2 + (ct / maxCt) * 4.2;
        const stroke = interpolateHex("#94a3b8", "#1e293b", wt / maxWt);

        dot += `    "${escapeDot(e.caller)}" -> "${escapeDot(e.callee)}" [label="${label}", penwidth=${penwidth.toFixed(2)}, color="${stroke}"];\n`;
    });

    dot += "}\n";

    return dot;
}

function highlightSelectedSvgNode(container, selectedNodeId) {
    if (!container || !selectedNodeId) return;

    const nodes = container.querySelectorAll("g.node");
    nodes.forEach((node) => {
        const title = node.querySelector("title")?.textContent || "";
        if (title === selectedNodeId) {
            node.classList.add("selected-node", "active-node");
            const mainRect = node.querySelector("polygon, rect");
            if (mainRect) {
                mainRect.setAttribute("stroke", "#6b46c1");
                mainRect.setAttribute("stroke-width", "2");
            }
        }
    });
}

function shortName(value = "") {
    const str = String(value);
    return str.length > 28 ? `${str.slice(0, 28)}…` : str;
}

function shortFile(value = "") {
    const str = String(value);
    if (!str.includes("/")) return str;
    const parts = str.split("/");
    return parts.slice(-2).join("/");
}

function formatUs(value) {
    const us = Number.isFinite(value) ? Math.max(value, 0) : 0;
    if (us >= 1_000_000) return `${formatCount(us / 1_000_000, 2)} s`;
    if (us >= 1_000) return `${formatCount(us / 1_000, 2)} ms`;
    return `${formatCount(us)} µs`;
}

function formatBytes(value) {
    const bytes = Number.isFinite(value) ? Math.max(value, 0) : 0;
    if (bytes >= 1024 * 1024) return `${formatCount(bytes / (1024 * 1024), 2)} MB`;
    if (bytes >= 1024) return `${formatCount(bytes / 1024, 2)} kB`;
    return `${formatCount(bytes)} B`;
}

function formatCount(value, fractionDigits = 0) {
    return new Intl.NumberFormat("en-US", {
        minimumFractionDigits: fractionDigits,
        maximumFractionDigits: fractionDigits,
    }).format(value);
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
