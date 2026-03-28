function buildNodeMap(payload) {
    const nodes = payload?.nodes || {};
    const map = new Map();

    for (const node of Object.values(nodes)) {
        const key = node.nodeId || "unknown";
        const current = map.get(key) || { wt: 0, cpu: 0, mu: 0, pmu: 0, samples: 0 };
        current.wt += node?.inclusive_cost?.wt || 0;
        current.cpu += node?.inclusive_cost?.cpu || 0;
        current.mu += node?.inclusive_cost?.mu || 0;
        current.pmu += node?.inclusive_cost?.pmu || 0;
        current.samples += 1;
        map.set(key, current);
    }

    return map;
}

export function compareProfiles(candidatePayload, baselinePayload) {
    const candidateMap = buildNodeMap(candidatePayload);
    const baselineMap = buildNodeMap(baselinePayload);
    const allKeys = new Set([...candidateMap.keys(), ...baselineMap.keys()]);

    const rows = [];
    for (const key of allKeys) {
        const current = candidateMap.get(key) || { wt: 0, cpu: 0, mu: 0, pmu: 0 };
        const baseline = baselineMap.get(key) || { wt: 0, cpu: 0, mu: 0, pmu: 0 };

        rows.push({
            name: key,
            wt: current.wt,
            baselineWt: baseline.wt,
            wtDelta: current.wt - baseline.wt,
            cpu: current.cpu,
            baselineCpu: baseline.cpu,
            cpuDelta: current.cpu - baseline.cpu,
            mu: current.mu,
            baselineMu: baseline.mu,
            muDelta: current.mu - baseline.mu,
            pmu: current.pmu,
            baselinePmu: baseline.pmu,
            pmuDelta: current.pmu - baseline.pmu,
        });
    }

    return rows.sort((a, b) => b.wtDelta - a.wtDelta);
}
