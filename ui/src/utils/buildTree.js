export default function buildTree(payload) {
    const nodes = payload.nodes;
    const edges = payload.edges;

    const children = {};
    for (const e of Object.values(edges)) {
        if (!children[e.caller]) children[e.caller] = [];
        children[e.caller].push(e.callee);
    }

    function build(id, path = new Set()) {
        const node = nodes[id];

        if (!node) {
            return {
                id,
                name: id,
                cost: 0,
                children: []
            };
        }

        const inCurrentPath = path.has(id);

        if (inCurrentPath) {
            return {
                id,
                name: `${node.nodeId} (recursion)`,
                cost: 0,
                children: []
            };
        }

        path.add(id);

        const built = {
            id,
            name: node.nodeId,
            cost: node.inclusive_cost?.wt || 0,
            children: (children[id] || []).map(childId => build(childId, path))
        };

        path.delete(id);

        return built;
    }

    return build(payload.root);
}
