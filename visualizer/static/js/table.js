/**
 * Entry point
 */
function renderTable(session) {
    globalSession = session;

    enrichGraph(session);

    const list = document.getElementById("func-list");
    const count = document.getElementById("func-count");

    let nodes = Object.entries(session.nodes);

    // tri par coût
    nodes.sort((a, b) => (b[1].inclusive_cost?.wt || 0) - (a[1].inclusive_cost?.wt || 0));

    // limiter
    nodes = nodes.slice(0, 200);

    count.textContent = nodes.length + " fonctions";

    let html = "";

    for (const [id, node] of nodes) {
        const pctInc = node.inclusive_percentage?.wt ?? 0;
        const pctExc = node.exclusive_percentage?.wt ?? 0;
        const calls = node.inclusive_cost?.ct ?? 0;

        html += `
        <tr class="fn-row" data-id="${id}">
            <td>
                <div class="metric-pastille" style="background:${getPercentageColor(pctInc)}"></div>
            </td>

            <td class="fn-name">${escapeHtml(node.nodeId || id)}</td>

            <td>
                <div class="progress">
                    <div class="progress-inc" style="width:${pctInc}%"></div>
                    <div class="progress-exc" style="width:${pctExc}%"></div>
                </div>
            </td>

            <td class="fn-calls">${calls}</td>
        </tr>
        `;
    }

    list.innerHTML = html;

    attachTableHandlers();
}

/**
 * Enrich graph with callers / callees
 */
function enrichGraph(session) {
    const { nodes, edges } = session;

    for (const id in nodes) {
        nodes[id]._computed = {
            callers: [],
            callees: []
        };
    }

    for (const e of Object.values(edges || {})) {
        if (nodes[e.caller] && nodes[e.callee]) {
            nodes[e.caller]._computed.callees.push(e);
            nodes[e.callee]._computed.callers.push(e);
        }
    }
}

/**
 * Click handler
 */
function attachTableHandlers() {
    document.querySelectorAll(".fn-row").forEach(row => {
        row.addEventListener("click", () => {
            const id = row.dataset.id;

            const existing = document.querySelector(`.fn-details[data-id="${id}"]`);

            // toggle OFF
            if (existing) {
                existing.remove();
                row.classList.remove("active");
                return;
            }

            // reset
            document.querySelectorAll(".fn-row").forEach(r => r.classList.remove("active"));
            document.querySelectorAll(".fn-details").forEach(d => d.remove());

            row.classList.add("active");

            const node = globalSession.nodes[id];

            const detailsHtml = `
                <tr class="fn-details" data-id="${id}">
                    <td colspan="4">
                        ${buildDetailsFull(id, node)}
                    </td>
                </tr>
            `;

            row.insertAdjacentHTML("afterend", detailsHtml);

            // sync graph
            if (typeof zoomToNode === "function") {
                zoomToNode(id);
            }
        });
    });
}

/**
 * Build full details block
 */
function buildDetailsFull(id, node) {
    const { callers = [], callees = [] } = node._computed || {};

    return `
    <div class="details-wrapper">

        ${renderCallers(callers)}

        <div class="details-metrics">
            ${metricRow("Wall Time", node.inclusive_cost?.wt)}
            ${metricRow("CPU", node.inclusive_cost?.cpu)}
            ${metricRow("I/O", node.inclusive_cost?.io)}
            ${metricRow("Memory", node.inclusive_cost?.pmu)}
            ${metricRow("Network", node.inclusive_cost?.nw)}
        </div>

        ${renderCallees(callees)}

    </div>
    `;
}

/**
 * Callers (bar distribution)
 */
function renderCallers(callers) {
    if (!callers.length) return "";

    const total = callers.reduce((s, e) => s + (e.calls || 0), 0);

    let html = `
        <div class="calls-title">Callers (${callers.length} - ${total} calls)</div>
        <div class="call-bars">
    `;

    callers.forEach(e => {
        const pct = total ? (e.calls / total) * 100 : 0;

        html += `
            <div 
                class="call-bar" 
                title="${escapeHtml(globalSession.nodes[e.caller]?.nodeId || e.caller)} (${e.calls})"
                style="width:${pct}%">
            </div>
        `;
    });

    html += `</div>`;

    return html;
}

/**
 * Callees list
 */
function renderCallees(callees) {
    if (!callees.length) return "";

    let html = `<div class="calls-title">Callees (${callees.length})</div>`;

    callees.forEach(e => {
        const target = globalSession.nodes[e.callee];
        if (!target) return;

        const pct = target.inclusive_percentage?.wt ?? 0;

        html += `
        <div class="callee-row">
            <span class="callee-name">${escapeHtml(target.nodeId)}</span>
            <div class="mini-bar">
                <div style="width:${pct}%"></div>
            </div>
        </div>
        `;
    });

    return html;
}

/**
 * Metric row
 */
function metricRow(label, value) {
    return `
        <div class="metric-row">
            <span class="metric-label">${label}</span>
            <span class="metric-value">${formatMetric(label, value)}</span>
        </div>
    `;
}
