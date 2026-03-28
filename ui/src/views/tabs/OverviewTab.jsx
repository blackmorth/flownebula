import { useMemo, useState } from "react";
import { Table, Box, Flex, SimpleGrid, Stat, Text, Badge, VStack, HStack } from "@chakra-ui/react";
import { ChevronUp, ChevronDown } from "lucide-react";
import { Icon } from "@chakra-ui/react";
import buildTree from "../../utils/buildTree";
import { buildDiagnostics } from "../../utils/profileInsights";
import { compareProfiles } from "../../utils/profileCompare";
import { formatBytes, formatDeltaPercent, formatDuration, formatPercent } from "../../utils/formatters";

function percentile(values, p) {
    if (values.length === 0) return 0;
    const sorted = [...values].sort((a, b) => a - b);
    const idx = Math.min(sorted.length - 1, Math.floor((p / 100) * sorted.length));
    return sorted[idx];
}

export default function OverviewTab({ payload, baselinePayload = null }) {
    const rows = Object.entries(payload?.nodes || {}).map(([id, node]) => ({
        id,
        name: node.nodeId,
        wt: node.inclusive_cost?.wt || 0,
        cpu: node.inclusive_cost?.cpu || 0,
        mu: node.inclusive_cost?.mu || 0,
        pmu: node.inclusive_cost?.pmu || 0,
    }));

    const [sortKey, setSortKey] = useState("wt");
    const [sortDir, setSortDir] = useState("desc");

    const totalWt = rows.reduce((acc, item) => acc + item.wt, 0);
    const wtValues = rows.map((row) => row.wt);
    const percentiles = {
        p50: percentile(wtValues, 50),
        p90: percentile(wtValues, 90),
        p99: percentile(wtValues, 99),
    };

    const tree = useMemo(() => buildTree(payload), [payload]);
    const diagnostics = useMemo(() => buildDiagnostics(tree), [tree]);

    const comparisonRows = useMemo(() => {
        if (!baselinePayload) return [];
        return compareProfiles(payload, baselinePayload).slice(0, 8);
    }, [baselinePayload, payload]);

    function sortBy(key) {
        if (sortKey === key) {
            setSortDir(sortDir === "asc" ? "desc" : "asc");
        } else {
            setSortKey(key);
            setSortDir("desc");
        }
    }

    const sorted = [...rows].sort((a, b) => {
        if (sortKey === "name") {
            const cmp = String(a.name).localeCompare(String(b.name));
            return sortDir === "asc" ? cmp : -cmp;
        }

        const diff = a[sortKey] - b[sortKey];
        return sortDir === "asc" ? diff : -diff;
    });

    const SortIcon = ({ column }) => {
        if (sortKey !== column) return null;
        return sortDir === "asc" ? (
            <Icon size="lg" color="tomato">
                <ChevronUp size={14} style={{ marginLeft: 4 }} />
            </Icon>
        ) : (
            <Icon size="lg" color="tomato">
                <ChevronDown size={14} style={{ marginLeft: 4 }} />
            </Icon>
        );
    };

    const HeaderCell = ({ column, children }) => (
        <Table.ColumnHeader
            onClick={() => sortBy(column)}
            cursor="pointer"
            userSelect="none"
            _hover={{ color: "purple.400" }}
            whiteSpace="nowrap"
        >
            <Flex align="center">
                {children}
                <SortIcon column={column} />
            </Flex>
        </Table.ColumnHeader>
    );

    return (
        <VStack align="stretch" gap={5}>
            <SimpleGrid columns={{ base: 1, md: 3 }} gap={4}>
                <Stat.Root p={4} rounded="md" borderWidth="1px">
                    <Stat.Label>Total wall time</Stat.Label>
                    <Stat.ValueText>{formatDuration(totalWt)}</Stat.ValueText>
                    <Stat.HelpText>{rows.length} functions profilées</Stat.HelpText>
                </Stat.Root>

                <Stat.Root p={4} rounded="md" borderWidth="1px">
                    <Stat.Label>Distribution (wt)</Stat.Label>
                    <Stat.ValueText>
                        p50 {formatDuration(percentiles.p50)} · p90 {formatDuration(percentiles.p90)}
                    </Stat.ValueText>
                    <Stat.HelpText>p99 {formatDuration(percentiles.p99)}</Stat.HelpText>
                </Stat.Root>

                <Stat.Root p={4} rounded="md" borderWidth="1px">
                    <Stat.Label>Concentration</Stat.Label>
                    <Stat.ValueText>
                        {rows.length > 0 ? formatPercent((Math.max(...wtValues, 0) / Math.max(totalWt, 1)) * 100) : "0%"}
                    </Stat.ValueText>
                    <Stat.HelpText>Part de la fonction la plus coûteuse</Stat.HelpText>
                </Stat.Root>
            </SimpleGrid>

            <Box borderWidth="1px" rounded="md" p={4}>
                <Text fontWeight="bold" mb={3}>Diagnostic automatique</Text>
                <SimpleGrid columns={{ base: 1, md: 2 }} gap={3}>
                    {[
                        ["Hotspots SQL", diagnostics.sqlHotspots],
                        ["Suspects N+1", diagnostics.nPlusOneSuspects],
                        ["Contention de lock", diagnostics.lockContentionSuspects],
                        ["Appels externes lents", diagnostics.slowExternalCalls],
                        ["Récursions / répétitions", diagnostics.recursionSuspects],
                        ["Hot path", diagnostics.hotPath],
                    ].map(([title, values]) => (
                        <Box key={title} borderWidth="1px" rounded="md" p={3}>
                            <Text fontSize="sm" fontWeight="semibold" mb={2}>{title}</Text>
                            {values.length === 0 ? (
                                <Badge colorPalette="gray">Aucun signal</Badge>
                            ) : (
                                <VStack align="start" gap={1}>
                                    {values.map((item) => <Text fontSize="sm" key={item}>• {item}</Text>)}
                                </VStack>
                            )}
                        </Box>
                    ))}
                </SimpleGrid>
            </Box>

            {baselinePayload && (
                <Box borderWidth="1px" rounded="md" p={4}>
                    <Text fontWeight="bold" mb={3}>Diff Baseline vs Candidate</Text>
                    <Table.Root size="sm" variant="outline">
                        <Table.Header>
                            <Table.Row>
                                <Table.ColumnHeader>Function</Table.ColumnHeader>
                                <Table.ColumnHeader>Δ Wall Time</Table.ColumnHeader>
                                <Table.ColumnHeader>Candidate</Table.ColumnHeader>
                                <Table.ColumnHeader>Baseline</Table.ColumnHeader>
                            </Table.Row>
                        </Table.Header>
                        <Table.Body>
                            {comparisonRows.map((row) => (
                                <Table.Row key={row.name}>
                                    <Table.Cell>{row.name}</Table.Cell>
                                    <Table.Cell>
                                        <HStack>
                                            <Text>{`${row.wtDelta >= 0 ? "+" : "-"}${formatDuration(Math.abs(row.wtDelta))}`}</Text>
                                            <Badge colorPalette={row.wtDelta > 0 ? "red" : "green"}>
                                                {formatDeltaPercent(row.wt, row.baselineWt)}
                                            </Badge>
                                        </HStack>
                                    </Table.Cell>
                                    <Table.Cell>{formatDuration(row.wt)}</Table.Cell>
                                    <Table.Cell>{formatDuration(row.baselineWt)}</Table.Cell>
                                </Table.Row>
                            ))}
                        </Table.Body>
                    </Table.Root>
                </Box>
            )}

            <Box overflowX="auto">
                <Table.Root size="sm" variant="outline">
                    <Table.Header>
                        <Table.Row>
                            <HeaderCell column="name">Function</HeaderCell>
                            <HeaderCell column="wt">Wall Time</HeaderCell>
                            <HeaderCell column="cpu">CPU</HeaderCell>
                            <HeaderCell column="mu">Memory</HeaderCell>
                            <HeaderCell column="pmu">Peak Mem</HeaderCell>
                            <Table.ColumnHeader>Part wt</Table.ColumnHeader>
                        </Table.Row>
                    </Table.Header>

                    <Table.Body>
                        {sorted.map((row) => (
                            <Table.Row key={row.id}>
                                <Table.Cell>{row.name}</Table.Cell>
                                <Table.Cell>{formatDuration(row.wt)}</Table.Cell>
                                <Table.Cell>{formatDuration(row.cpu)}</Table.Cell>
                                <Table.Cell>{formatBytes(row.mu)}</Table.Cell>
                                <Table.Cell>{formatBytes(row.pmu)}</Table.Cell>
                                <Table.Cell>{formatPercent((row.wt / Math.max(totalWt, 1)) * 100)}</Table.Cell>
                            </Table.Row>
                        ))}
                    </Table.Body>
                </Table.Root>
            </Box>
        </VStack>
    );
}
