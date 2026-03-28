export function formatDuration(ns) {
    const value = Number.isFinite(ns) ? Math.max(ns, 0) : 0;

    const units = [
        { size: 3_600_000_000_000, label: "h" },
        { size: 60_000_000_000, label: "m" },
        { size: 1_000_000_000, label: "s" },
        { size: 1_000_000, label: "ms" },
        { size: 1_000, label: "µs" },
    ];

    for (const unit of units) {
        if (value >= unit.size) {
            return `${new Intl.NumberFormat("en-US", {
                minimumFractionDigits: value >= unit.size * 10 ? 0 : 2,
                maximumFractionDigits: 2,
            }).format(value / unit.size)} ${unit.label}`;
        }
    }

    return `${new Intl.NumberFormat("en-US").format(Math.round(value))} ns`;
}

export function formatBytes(bytes) {
    const value = Number.isFinite(bytes) ? Math.max(bytes, 0) : 0;
    const units = ["B", "KB", "MB", "GB", "TB"];
    let current = value;
    let unit = units[0];

    for (let i = 1; i < units.length && current >= 1024; i += 1) {
        current /= 1024;
        unit = units[i];
    }

    return `${new Intl.NumberFormat("en-US", {
        minimumFractionDigits: current >= 10 ? 0 : 2,
        maximumFractionDigits: 2,
    }).format(current)} ${unit}`;
}

export function formatPercent(value) {
    const safe = Number.isFinite(value) ? value : 0;
    return `${new Intl.NumberFormat("en-US", {
        minimumFractionDigits: 0,
        maximumFractionDigits: 2,
    }).format(safe)}%`;
}

export function formatDeltaPercent(current, baseline) {
    if (!Number.isFinite(current) || !Number.isFinite(baseline)) {
        return "—";
    }

    if (baseline === 0) {
        return current === 0 ? "0.00%" : "+∞";
    }

    const delta = ((current - baseline) / Math.abs(baseline)) * 100;
    const prefix = delta > 0 ? "+" : "";

    return `${prefix}${new Intl.NumberFormat("en-US", {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
    }).format(delta)}%`;
}
