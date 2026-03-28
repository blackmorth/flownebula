export default function buildTree(payload) {
    const nodes = payload?.nodes || {};
    const edges = payload?.edges || {};

    const children = {};
    for (const edge of Object.values(edges)) {
        if (!children[edge.caller]) children[edge.caller] = [];
        children[edge.caller].push(edge);
    }

    function build(id, path = new Set(), viaEdge = null) {
        const node = nodes[id];
        const displayName = node?.nodeId || id;
        const inCurrentPath = path.has(id);
        const meta = {
            t_start: viaEdge?.t_start ?? node?.t_start ?? null,
            t_end: viaEdge?.t_end ?? node?.t_end ?? null,
        };

        if (inCurrentPath) {
            return {
                id: viaEdge?.edgeId ? `${id}#${viaEdge.edgeId}` : id,
                name: `${displayName} (recursion)`,
                cost: viaEdge?.cost?.wt || 0,
                meta,
                children: []
            };
        }

        path.add(id);

        const built = {
            id: viaEdge?.edgeId ? `${id}#${viaEdge.edgeId}` : id,
            name: displayName,
            cost: viaEdge?.cost?.wt ?? node?.inclusive_cost?.wt ?? 0,
            meta,
            children: (children[id] || []).map((edge) => build(edge.callee, path, edge))
        };

        path.delete(id);
        return built;
    }

    return build(payload?.root || "root");
}
