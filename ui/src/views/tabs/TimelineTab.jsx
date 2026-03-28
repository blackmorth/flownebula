import { useMemo, useState } from "react";
import { Alert, Box, Flex, Input, Text } from "@chakra-ui/react";
import buildTree from "../../utils/buildTree";
import { buildDiagnostics, filterRows, flattenTree } from "../../utils/profileInsights";

const ROW_HEIGHT = 30;

function formatDuration(us) {
    const value = Number.isFinite(us) ? us : 0;

    if (value >= 1000) {
        return `${new Intl.NumberFormat("en-US", {
            minimumFractionDigits: 2,
            maximumFractionDigits: 2,
        }).format(value / 1000)} ms`;
    }

    return `${new Intl.NumberFormat("en-US").format(Math.round(value))} µs`;
}

function readTemporalBounds(node) {
    const tStart = node?.t_start ?? node?.tStart ?? node?.start ?? null;
    const tEnd = node?.t_end ?? node?.tEnd ?? node?.end ?? null;

    if (Number.isFinite(tStart) && Number.isFinite(tEnd) && tEnd >= tStart) {
        return { tStart, tEnd, absolute: true };
    }

    return null;
}

function buildTimelineRows(node) {
    const rows = [];
    let usesAbsoluteTime = false;

    function visit(current, start, end, depth, parentId = null) {
        const temporal = readTemporalBounds(current.meta || current);
        const rowStart = temporal ? temporal.tStart : start;
        const rowEnd = temporal ? temporal.tEnd : end;

        if (temporal) {
            usesAbsoluteTime = true;
        }

        const row = {
            id: `${current.id}-${depth}-${rowStart}`,
            nodeId: current.id,
            name: current.name,
            cost: current.cost,
            start: rowStart,
            end: rowEnd,
            depth,
            parentId,
            path: current.path,
            kind: current.kind,
        };

        rows.push(row);

        if (!current.children.length) return;

        const rawTotal = current.children.reduce((sum, child) => sum + Math.max(child.cost, 0), 0);
        const total = rawTotal > 0 ? rawTotal : current.children.length;
        let cursor = rowStart;

        current.children.forEach((child, index) => {
            const childTemporal = readTemporalBounds(child.meta || child);
            if (childTemporal) {
                visit(child, childTemporal.tStart, childTemporal.tEnd, depth + 1, row.id);
                return;
            }

            const remaining = rowEnd - cursor;
            const ratio = rawTotal > 0 ? Math.max(child.cost, 0) / total : 1 / total;
            const span = index === current.children.length - 1 ? remaining : (rowEnd - rowStart) * ratio;
            const childStart = cursor;
            const childEnd = Math.max(childStart, childStart + span);

            visit(child, childStart, childEnd, depth + 1, row.id);
            cursor = childEnd;
        });
    }

    visit(node, 0, Math.max(node.cost, 1), 0);
    return { rows, usesAbsoluteTime };
}

function DiagnosticsPanel({ diagnostics }) {
    const cards = [
        { label: "SQL hot spots", values: diagnostics.sqlHotspots },
        { label: "N+1 suspects", values: diagnostics.nPlusOneSuspects },
        { label: "Recursion suspects", values: diagnostics.recursionSuspects },
        { label: "Probable hot path", values: diagnostics.hotPath },
    ];

    return (
        <Flex gap={3} wrap="wrap" mb={4}>
            {cards.map((card) => (
                <Box key={card.label} borderWidth="1px" borderRadius="md" p={3} minW="280px" bg="white">
                    <Text fontWeight="700" mb={1}>{card.label}</Text>
                    {card.values.length ? card.values.map((value) => (
                        <Text key={value} color="gray.700" fontSize="sm">• {value}</Text>
                    )) : <Text color="gray.500" fontSize="sm">No strong signal.</Text>}
                </Box>
            ))}
        </Flex>
    );
}

export default function TimelineTab({ payload }) {
    const [hoveredId, setHoveredId] = useState(null);
    const [kindFilter, setKindFilter] = useState("all");
    const [query, setQuery] = useState("");

    const { rows, usesAbsoluteTime, diagnostics } = useMemo(() => {
        const tree = buildTree(payload);
        const enriched = (() => {
            const full = flattenTree(tree);
            const byId = new Map(full.map((node) => [node.id, node]));

            function merge(node) {
                const meta = byId.get(node.id);
                return {
                    ...node,
                    path: meta?.path || [node.name],
                    kind: meta?.kind || "other",
                    children: node.children.map(merge),
                };
            }

            return merge(tree);
        })();

        const timeline = buildTimelineRows(enriched);
        return {
            rows: timeline.rows,
            usesAbsoluteTime: timeline.usesAbsoluteTime,
            diagnostics: buildDiagnostics(enriched),
        };
    }, [payload]);

    const visibleRows = useMemo(
        () => filterRows(rows, { kind: kindFilter, query }),
        [rows, kindFilter, query],
    );

    const totalDuration = visibleRows[0]?.end ? Math.max(visibleRows[0].end - visibleRows[0].start, 1) : 1;
    const origin = visibleRows[0]?.start || 0;

    return (
        <Box>
            <DiagnosticsPanel diagnostics={diagnostics} />

            <Flex gap={3} mb={3}>
                <Box as="select" maxW="220px" value={kindFilter} onChange={(e) => setKindFilter(e.target.value)} borderWidth="1px" borderRadius="md" p={2} bg="white">
                    <option value="all">All nodes</option>
                    <option value="endpoint">Endpoint</option>
                    <option value="transaction">Transaction</option>
                    <option value="class">Class</option>
                    <option value="package">Package</option>
                    <option value="sql">SQL only</option>
                </Box>
                <Input
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    placeholder="Filter by endpoint / class / package / SQL pattern"
                    bg="white"
                />
            </Flex>

            {!usesAbsoluteTime && (
                <Alert.Root status="warning" mb={4}>
                    <Alert.Indicator />
                    <Alert.Content>
                        <Alert.Title>Synthetic timeline mode</Alert.Title>
                        <Alert.Description>
                            This payload has no absolute events (`t_start`, `t_end`). Child spans are distributed by relative cost.
                        </Alert.Description>
                    </Alert.Content>
                </Alert.Root>
            )}

            <Flex gap={4} align="stretch" minH="70vh">
                <Box
                    flex="0 0 40%"
                    borderWidth="1px"
                    borderRadius="md"
                    overflow="auto"
                    maxH="75vh"
                    p={2}
                    bg="white"
                >
                    {visibleRows.map((row) => {
                        const isActive = hoveredId === row.id;

                        return (
                            <Flex
                                key={`list-${row.id}`}
                                align="center"
                                py={1.5}
                                px={2}
                                borderRadius="md"
                                bg={isActive ? "blue.50" : "transparent"}
                                cursor="pointer"
                                ml={`${row.depth * 14}px`}
                                onMouseEnter={() => setHoveredId(row.id)}
                                onMouseLeave={() => setHoveredId(null)}
                            >
                                <Text fontWeight={isActive ? "700" : "500"} flex="1" overflow="hidden" whiteSpace="nowrap" textOverflow="ellipsis">
                                    {row.name}
                                </Text>
                                <Text color="gray.600" fontSize="sm" ml={3} whiteSpace="nowrap">
                                    {formatDuration(row.cost)}
                                </Text>
                            </Flex>
                        );
                    })}
                </Box>

                <Box
                    flex="1"
                    borderWidth="1px"
                    borderRadius="md"
                    overflow="auto"
                    maxH="75vh"
                    bg="white"
                    p={3}
                >
                    <Box position="relative" minW="640px" h={`${visibleRows.length * ROW_HEIGHT + 20}px`}>
                        {visibleRows.map((row, index) => {
                            const isActive = hoveredId === row.id;
                            const left = ((row.start - origin) / totalDuration) * 100;
                            const width = Math.max(((row.end - row.start) / totalDuration) * 100, 0.6);

                            return (
                                <Box
                                    key={`timeline-${row.id}`}
                                    position="absolute"
                                    top={`${index * ROW_HEIGHT}px`}
                                    left={`${left}%`}
                                    width={`${width}%`}
                                    height="22px"
                                    borderRadius="sm"
                                    bg={isActive ? "blue.500" : "purple.500"}
                                    color="white"
                                    px={2}
                                    fontSize="xs"
                                    lineHeight="22px"
                                    overflow="hidden"
                                    whiteSpace="nowrap"
                                    textOverflow="ellipsis"
                                    cursor="pointer"
                                    onMouseEnter={() => setHoveredId(row.id)}
                                    onMouseLeave={() => setHoveredId(null)}
                                    title={`${row.name} — ${formatDuration(row.cost)}`}
                                >
                                    {row.name}
                                </Box>
                            );
                        })}
                    </Box>
                </Box>
            </Flex>
        </Box>
    );
}
