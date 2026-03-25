import { useMemo, useState } from "react";
import { Box, Flex, Text } from "@chakra-ui/react";
import buildTree from "../../utils/buildTree";

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

function buildTimelineRows(node) {
    const rows = [];

    function visit(current, start, end, depth, parentId = null) {
        const row = {
            id: `${current.id}-${depth}-${start}`,
            nodeId: current.id,
            name: current.name,
            cost: current.cost,
            start,
            end,
            depth,
            parentId,
        };

        rows.push(row);

        if (!current.children.length) return;

        const rawTotal = current.children.reduce((sum, child) => sum + Math.max(child.cost, 0), 0);
        const total = rawTotal > 0 ? rawTotal : current.children.length;
        let cursor = start;

        current.children.forEach((child, index) => {
            const ratio = rawTotal > 0 ? Math.max(child.cost, 0) / total : 1 / total;
            const remaining = end - cursor;
            const span = index === current.children.length - 1 ? remaining : (end - start) * ratio;
            const childStart = cursor;
            const childEnd = Math.max(childStart, childStart + span);

            visit(child, childStart, childEnd, depth + 1, row.id);
            cursor = childEnd;
        });
    }

    visit(node, 0, Math.max(node.cost, 1), 0);
    return rows;
}

export default function TimelineTab({ payload }) {
    const [hoveredId, setHoveredId] = useState(null);

    const rows = useMemo(() => {
        const tree = buildTree(payload);
        return buildTimelineRows(tree);
    }, [payload]);

    const totalDuration = rows[0]?.end || 1;

    return (
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
                {rows.map((row) => {
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
                            <Text
                                fontWeight={isActive ? "700" : "500"}
                                flex="1"
                                overflow="hidden"
                                whiteSpace="nowrap"
                                textOverflow="ellipsis"
                            >
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
                <Box position="relative" minW="640px" h={`${rows.length * ROW_HEIGHT + 20}px`}>
                    {rows.map((row, index) => {
                        const isActive = hoveredId === row.id;
                        const left = (row.start / totalDuration) * 100;
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
    );
}
