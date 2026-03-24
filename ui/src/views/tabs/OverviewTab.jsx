import { useState } from "react";
import { Table, Box, Flex } from "@chakra-ui/react";
import { ChevronUp, ChevronDown } from "lucide-react";
import { Icon } from "@chakra-ui/react"


export default function OverviewTab({ payload }) {
    const rows = Object.entries(payload.nodes).map(([id, node]) => ({
        id,
        name: node.nodeId,
        wt: node.inclusive_cost?.wt || 0,
        cpu: node.inclusive_cost?.cpu || 0,
        mu: node.inclusive_cost?.mu || 0,
        pmu: node.inclusive_cost?.pmu || 0,
    }));

    const [sortKey, setSortKey] = useState("wt");
    const [sortDir, setSortDir] = useState("desc");

    function sortBy(key) {
        if (sortKey === key) {
            setSortDir(sortDir === "asc" ? "desc" : "asc");
        } else {
            setSortKey(key);
            setSortDir("desc");
        }
    }

    const sorted = [...rows].sort((a, b) => {
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
        <Box overflowX="auto">
            <Table.Root size="sm" variant="outline">
                <Table.Header>
                    <Table.Row>
                        <HeaderCell column="name">Function</HeaderCell>
                        <HeaderCell column="wt">Wall Time</HeaderCell>
                        <HeaderCell column="cpu">CPU</HeaderCell>
                        <HeaderCell column="mu">Memory</HeaderCell>
                        <HeaderCell column="pmu">Peak Mem</HeaderCell>
                    </Table.Row>
                </Table.Header>

                <Table.Body>
                    {sorted.map(row => (
                        <Table.Row key={row.id}>
                            <Table.Cell>{row.name}</Table.Cell>
                            <Table.Cell>{row.wt}</Table.Cell>
                            <Table.Cell>{row.cpu}</Table.Cell>
                            <Table.Cell>{row.mu}</Table.Cell>
                            <Table.Cell>{row.pmu}</Table.Cell>
                        </Table.Row>
                    ))}
                </Table.Body>
            </Table.Root>
        </Box>
    );
}
