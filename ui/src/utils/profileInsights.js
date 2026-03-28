export function normalizeText(value) {
    return String(value || "").toLowerCase();
}

export function detectNodeKind(name) {
    const lower = normalizeText(name);

    if (/\b(sql|select|insert|update|delete|from|join|repository|dao)\b/.test(lower)) {
        return "sql";
    }

    if (/\b(http|route|controller|endpoint|handler|api|grpc|graphql)\b/.test(lower)) {
        return "endpoint";
    }

    if (/\b(redis|memcache|kafka|sqs|sns|rabbitmq|amqp|queue)\b/.test(lower)) {
        return "external";
    }

    if (/\b(lock|mutex|semaphore|critical section|spinlock)\b/.test(lower)) {
        return "lock";
    }

    if (/transaction|tx\b/.test(lower)) {
        return "transaction";
    }

    if (name?.includes("::") || /^[A-Z][\w]+\.[A-Z][\w]+/.test(name || "")) {
        return "class";
    }

    if (name?.includes("/") || name?.includes(".")) {
        return "package";
    }

    return "other";
}

export function flattenTree(root) {
    const rows = [];

    function walk(node, depth = 0, lineage = []) {
        const kind = detectNodeKind(node.name);
        const path = [...lineage, node.name];
        rows.push({ ...node, depth, path, kind });

        node.children.forEach((child) => walk(child, depth + 1, path));
    }

    walk(root);
    return rows;
}

export function buildDiagnostics(tree) {
    const flat = flattenTree(tree);
    const byName = new Map();

    for (const node of flat) {
        byName.set(node.name, (byName.get(node.name) || 0) + 1);
    }

    const total = Math.max(tree.cost || 0, 1);

    const sqlHotspots = flat
        .filter((node) => node.kind === "sql")
        .sort((a, b) => b.cost - a.cost)
        .slice(0, 3)
        .map((node) => `${node.name} (${Math.round((node.cost / total) * 100)}%)`);

    const recursionSuspects = flat
        .filter((node) => node.name.includes("(recursion)") || byName.get(node.name) > 5)
        .slice(0, 3)
        .map((node) => node.name);

    const nPlusOneSuspects = [...byName.entries()]
        .filter(([name, count]) => count > 8 && /sql|query|select|find|repository/i.test(name))
        .sort((a, b) => b[1] - a[1])
        .slice(0, 3)
        .map(([name, count]) => `${name} (x${count})`);

    const lockContentionSuspects = flat
        .filter((node) => node.kind === "lock")
        .sort((a, b) => b.cost - a.cost)
        .slice(0, 3)
        .map((node) => `${node.name} (${Math.round((node.cost / total) * 100)}%)`);

    const slowExternalCalls = flat
        .filter((node) => node.kind === "external" || /http|grpc|api client|fetch/i.test(node.name))
        .sort((a, b) => b.cost - a.cost)
        .slice(0, 3)
        .map((node) => `${node.name} (${Math.round((node.cost / total) * 100)}%)`);

    const hotPath = flat
        .filter((node) => node.depth > 0)
        .sort((a, b) => b.cost - a.cost)
        .slice(0, 1)
        .map((node) => node.path.join(" → "));

    return {
        sqlHotspots,
        recursionSuspects,
        nPlusOneSuspects,
        lockContentionSuspects,
        slowExternalCalls,
        hotPath,
    };
}

export function filterRows(rows, { kind = "all", query = "" }) {
    const normalizedQuery = normalizeText(query).trim();

    return rows.filter((row) => {
        const kindMatches = kind === "all" || row.kind === kind;
        const queryMatches =
            normalizedQuery.length === 0 ||
            normalizeText(row.name).includes(normalizedQuery) ||
            normalizeText(row.path?.join(" ")).includes(normalizedQuery);

        return kindMatches && queryMatches;
    });
}
