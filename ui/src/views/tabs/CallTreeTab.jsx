import { Box, Flex, Text } from "@chakra-ui/react";
import buildTree from "../../utils/buildTree";

function formatDuration(ns) {
    const value = Number.isFinite(ns) ? Math.max(ns, 0) : 0;

    const units = [
        { size: 3_600_000_000_000, label: "h" },
        { size: 60_000_000_000, label: "m" },
        { size: 1_000_000_000, label: "s" },
        { size: 1_000_000, label: "ms" },
        { size: 1_000, label: "µs" },
    ];

    for (const unit of units) {
        if (value >= unit.size) {
            return `${new Intl.NumberFormat("en-US", {
                minimumFractionDigits: value >= unit.size * 10 ? 0 : 2,
                maximumFractionDigits: 2,
            }).format(value / unit.size)} ${unit.label}`;
        }
    }

    return `${new Intl.NumberFormat("en-US").format(Math.round(value))} ns`;
}

function CallTree({ node, depth = 0, total = node.cost || 1 }) {
    const width = Math.max((node.cost / total) * 100, 1);

    return (
        <Box mt={2}>
            <Flex
                align="center"
                borderRadius="md"
                bg={depth === 0 ? "purple.600" : "purple.500"}
                color="white"
                py={1}
                px={2}
                ml={`${depth * 18}px`}
                minW="220px"
                width={`${width}%`}
            >
                <Text fontWeight="bold" flex="1" overflow="hidden" whiteSpace="nowrap" textOverflow="ellipsis">
                    {node.name}
                </Text>
                <Text fontSize="sm" opacity={0.95} ml={3} whiteSpace="nowrap">
                    {formatDuration(node.cost)}
                </Text>
            </Flex>

            {node.children.map((child) => (
                <CallTree key={child.id} node={child} depth={depth + 1} total={total} />
            ))}
        </Box>
    );
}

export default function CallTreeTab({ payload }) {
    const tree = buildTree(payload);

    return <CallTree node={tree} total={Math.max(tree.cost, 1)} />;
}
