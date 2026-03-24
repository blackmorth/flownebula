import { Box, Text } from "@chakra-ui/react";
import buildTree from "../../utils/buildTree";
import collapseSmallNodes from "../../utils/collapseSmallNodes";

function CallTree({ node, depth = 0 }) {
    return (
        <Box ml={depth * 20} mt={2}>
            <Text fontWeight="bold">
                {node.name} — {node.cost} µs
            </Text>
            {node.children.map(child => (
                <CallTree key={child.id} node={child} depth={depth + 1} />
            ))}
        </Box>
    );
}

export default function CallTreeTab({ payload }) {
    const tree = buildTree(payload);
    const collapsed = collapseSmallNodes(tree, payload.peaks.inclusive.wt);

    return <CallTree node={collapsed} />;
}
