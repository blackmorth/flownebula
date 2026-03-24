export default function buildTree(payload) {
    const nodes = payload.nodes;
    const edges = payload.edges;

    const children = {};
    for (const e of Object.values(edges)) {
        if (!children[e.caller]) children[e.caller] = [];
        children[e.caller].push(e.callee);
    }

    function build(id) {
        const node = nodes[id];
        return {
            id,
            name: node.nodeId,
            cost: node.inclusive_cost?.wt || 0,
            children: (children[id] || []).map(build)
        };
    }

    return build(payload.root);
}
