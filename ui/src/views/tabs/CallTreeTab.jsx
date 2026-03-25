import { Box, Flex, Text } from "@chakra-ui/react";
import buildTree from "../../utils/buildTree";

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

function collapseSmallNodes(tree, parentCost, threshold = 0.01) {
    const effectiveParentCost = Math.max(parentCost || tree.cost || 1, 1);
    const children = [];
    const collapsed = { id: `${tree.id}-other`, name: "(other)", cost: 0, children: [] };

    for (const child of tree.children) {
        const ratio = child.cost / effectiveParentCost;

        if (ratio < threshold) {
            collapsed.cost += child.cost;
            collapsed.children.push(child);
        } else {
            children.push(collapseSmallNodes(child, child.cost, threshold));
        }
    }

    if (collapsed.cost > 0) {
        children.push(collapsed);
    }

    return { ...tree, children };
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
    const collapsed = collapseSmallNodes(tree, tree.cost);

    return <CallTree node={collapsed} total={Math.max(collapsed.cost, 1)} />;
}
