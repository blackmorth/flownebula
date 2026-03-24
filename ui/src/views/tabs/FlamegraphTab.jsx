import { Box } from "@chakra-ui/react";
import buildTree from "../../utils/buildTree";
import collapseSmallNodes from "../../utils/collapseSmallNodes";

function FlameNode({ node, depth }) {
    const width = Math.max(node.cost / 1000, 2);

    return (
        <Box>
            <Box
                bg="purple.500"
                color="white"
                p={1}
                ml={depth * 20}
                width={`${width}px`}
                whiteSpace="nowrap"
                borderRadius="md"
            >
                {node.name} ({node.cost})
            </Box>

            {node.children.map(child => (
                <FlameNode key={child.id} node={child} depth={depth + 1} />
            ))}
        </Box>
    );
}

export default function FlamegraphTab({ payload }) {
    const tree = buildTree(payload);
    const collapsed = collapseSmallNodes(tree, payload.peaks.inclusive.wt);

    return <FlameNode node={collapsed} depth={0} />;
}
