let globalSession = null;

document.addEventListener("DOMContentLoaded", () => {
    const file = document.body.dataset.file;
    loadSession(`/api/json?file=${encodeURIComponent(file)}`);
});

function loadSession(url) {
    fetch(url)
        .then(r => r.json())
        .then(session => {
            globalSession = session;

            renderTable(session);      // table.js
            //renderGraph(session);      // elk-graph.js

            attachTableHandlers();     // table.js
           // attachGraphHandlers();     // elk-graph.js
        });
}
