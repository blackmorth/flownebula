export default function collapseSmallNodes(tree, totalCost, threshold = 0.01) {
    const children = [];
    let collapsed = { name: "(other)", cost: 0, children: [] };

    for (const child of tree.children) {
        const ratio = child.cost / totalCost;

        if (ratio < threshold) {
            collapsed.cost += child.cost;
            collapsed.children.push(child);
        } else {
            children.push(collapseSmallNodes(child, totalCost, threshold));
        }
    }

    if (collapsed.cost > 0) {
        children.push(collapsed);
    }

    return { ...tree, children };
}
