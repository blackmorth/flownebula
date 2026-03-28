import { useMemo, useState } from "react";
import { Box, Button, Flex, Input, Slider, Text } from "@chakra-ui/react";
import buildTree from "../../utils/buildTree";
import collapseSmallNodes from "../../utils/collapseSmallNodes";
import { filterRows, flattenTree, normalizeText } from "../../utils/profileInsights";

const BAR_HEIGHT = 28;

function buildFlameRows(root) {
    const rows = [];

    function walk(node, depth, offset, parentCost, ancestry) {
        const widthRatio = parentCost > 0 ? node.cost / parentCost : 1;
        rows.push({
            id: node.id,
            name: node.name,
            cost: node.cost,
            depth,
            left: offset,
            width: Math.max(widthRatio, 0.01),
            kind: node.kind,
            path: ancestry,
        });

        let cursor = offset;
        const safeCost = Math.max(node.cost, 1);
        node.children.forEach((child) => {
            const childWidth = child.cost / safeCost;
            walk(child, depth + 1, cursor, safeCost, [...ancestry, child.name]);
            cursor += childWidth;
        });
    }

    walk(root, 0, 0, root.cost || 1, [root.name]);
    return rows;
}

export default function FlamegraphTab({ payload }) {
    const [zoom, setZoom] = useState(1);
    const [query, setQuery] = useState("");
    const [kindFilter, setKindFilter] = useState("all");
    const [focusPath, setFocusPath] = useState([]);

    const tree = useMemo(() => {
        const built = buildTree(payload);
        const collapsed = collapseSmallNodes(built, payload.peaks.inclusive.wt);
        const annotated = (() => {
            const flat = flattenTree(collapsed);
            const byId = new Map(flat.map((node) => [node.id, node]));

            function enrich(node) {
                const meta = byId.get(node.id);
                return {
                    ...node,
                    kind: meta?.kind || "other",
                    path: meta?.path || [node.name],
                    children: node.children.map(enrich),
                };
            }

            return enrich(collapsed);
        })();

        if (focusPath.length === 0) return annotated;

        let current = annotated;
        for (const name of focusPath) {
            const next = current.children.find((child) => child.name === name);
            if (!next) return annotated;
            current = next;
        }

        return current;
    }, [payload, focusPath]);

    const rows = useMemo(() => {
        const allRows = buildFlameRows(tree);
        return filterRows(allRows, { kind: kindFilter, query });
    }, [tree, kindFilter, query]);

    return (
        <Box>
            <Flex gap={3} mb={3} wrap="wrap" align="center">
                <Button size="sm" onClick={() => setFocusPath([])} variant="outline">Reset focus</Button>
                <Box>
                    <Text fontSize="xs" color="gray.500">Zoom</Text>
                    <Slider.Root width="200px" min={1} max={8} step={0.5} value={[zoom]} onValueChange={(e) => setZoom(e.value[0])}>
                        <Slider.Control>
                            <Slider.Track>
                                <Slider.Range />
                            </Slider.Track>
                            <Slider.Thumbs />
                        </Slider.Control>
                    </Slider.Root>
                </Box>
                <Box as="select" value={kindFilter} onChange={(e) => setKindFilter(e.target.value)} borderWidth="1px" borderRadius="md" p={2} bg="white">
                    <option value="all">All nodes</option>
                    <option value="endpoint">Endpoint</option>
                    <option value="transaction">Transaction</option>
                    <option value="class">Class</option>
                    <option value="package">Package</option>
                    <option value="sql">SQL only</option>
                </Box>
                <Input value={query} onChange={(e) => setQuery(e.target.value)} placeholder="Search node" maxW="280px" bg="white" />
            </Flex>

            <Box borderWidth="1px" borderRadius="md" p={3} bg="white" overflow="auto">
                <Box position="relative" minW={`${900 * zoom}px`} h={`${(Math.max(...rows.map((r) => r.depth), 0) + 1) * BAR_HEIGHT + 16}px`}>
                    {rows.map((row) => {
                        const highlight = query.length > 0 && normalizeText(row.name).includes(normalizeText(query));
                        return (
                            <Box
                                key={`${row.id}-${row.depth}-${row.left}`}
                                position="absolute"
                                top={`${row.depth * BAR_HEIGHT}px`}
                                left={`${row.left * 100}%`}
                                width={`${Math.max(row.width * 100, 0.8)}%`}
                                height="22px"
                                borderRadius="sm"
                                bg={highlight ? "orange.500" : "purple.500"}
                                color="white"
                                px={2}
                                fontSize="xs"
                                lineHeight="22px"
                                overflow="hidden"
                                whiteSpace="nowrap"
                                textOverflow="ellipsis"
                                cursor="pointer"
                                onClick={() => setFocusPath(row.path.slice(1))}
                                title={`${row.path.join(" → ")} (${row.cost})`}
                            >
                                {row.name}
                            </Box>
                        );
                    })}
                </Box>
            </Box>
        </Box>
    );
}
