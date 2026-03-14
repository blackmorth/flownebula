fetch("../nebula.json")
    .then(r => r.json())
    .then(data => {

        const nodes = {};
        const links = [];

        // agrégation edges -> nodes + liens enrichis
        data.forEach(e => {
            const caller = e.caller;
            const callee = e.callee;
            const calls  = e.calls || 0;
            const time   = e.time  || 0;
            const mem    = e.mem   || e.mem_total || 0;

            if (!nodes[caller]) {
                nodes[caller] = {
                    id: caller,
                    calls_out: 0,
                    time_out:  0,
                    mem_out:   0
                };
            }

            if (!nodes[callee]) {
                nodes[callee] = {
                    id: callee,
                    calls_out: 0,
                    time_out:  0,
                    mem_out:   0
                };
            }

            nodes[caller].calls_out += calls;
            nodes[caller].time_out  += time;
            nodes[caller].mem_out   += mem;

            links.push({
                source: caller,
                target: callee,
                calls:  calls,
                time:   time,
                mem:    mem
            });
        });

        const nodeList = Object.values(nodes);

        const maxTime = d3.max(nodeList, d => d.time_out) || 1;
        const maxMem  = d3.max(nodeList, d => d.mem_out)  || 1;
        const maxCall = d3.max(links,   d => d.calls)     || 1;

        const radiusScale = d3.scaleSqrt()
            .domain([0, maxTime])
            .range([5, 25]);

        const colorScale = d3.scaleSequential(d3.interpolateYlOrRd)
            .domain([0, maxMem]);

        const linkWidthScale = d3.scaleSqrt()
            .domain([0, maxCall])
            .range([0.5, 6]);

        const svg = d3.select("svg");

        const tooltip = d3.select("body")
            .append("div")
            .attr("class", "nebula-tooltip")
            .style("position", "absolute")
            .style("padding", "6px 10px")
            .style("background", "rgba(0,0,0,0.8)")
            .style("color", "#fff")
            .style("font-size", "12px")
            .style("border-radius", "4px")
            .style("pointer-events", "none")
            .style("opacity", 0);

        const simulation = d3.forceSimulation(nodeList)
            .force("link", d3.forceLink(links).id(d => d.id).distance(120))
            .force("charge", d3.forceManyBody().strength(-300))
            .force("center", d3.forceCenter(500, 350));

        const link = svg.selectAll("line")
            .data(links)
            .enter()
            .append("line")
            .attr("stroke-width", d => linkWidthScale(d.calls));

        const node = svg.selectAll("circle")
            .data(nodeList)
            .enter()
            .append("circle")
            .attr("r", d => radiusScale(d.time_out))
            .attr("fill", d => colorScale(d.mem_out))
            .call(
                d3.drag()
                    .on("start", dragstarted)
                    .on("drag", dragged)
                    .on("end", dragended)
            )
            .on("mouseover", (event, d) => {
                const avgTime = d.calls_out ? d.time_out / d.calls_out : 0;
                const avgMem  = d.calls_out ? d.mem_out  / d.calls_out : 0;

                tooltip
                    .style("opacity", 1)
                    .html(
                        `<strong>${d.id}</strong><br>` +
                        `Calls out: ${d.calls_out}<br>` +
                        `Time total: ${d.time_out} ns<br>` +
                        `Time avg: ${Math.round(avgTime)} ns<br>` +
                        `Mem total: ${d.mem_out} bytes<br>` +
                        `Mem avg: ${Math.round(avgMem)} bytes`
                    )
                    .style("left", (event.pageX + 10) + "px")
                    .style("top", (event.pageY + 10) + "px");
            })
            .on("mousemove", (event) => {
                tooltip
                    .style("left", (event.pageX + 10) + "px")
                    .style("top", (event.pageY + 10) + "px");
            })
            .on("mouseout", () => {
                tooltip.style("opacity", 0);
            });

        const label = svg.selectAll("text")
            .data(nodeList)
            .enter()
            .append("text")
            .text(d => d.id);

        simulation.on("tick", () => {

            link
                .attr("x1", d => d.source.x)
                .attr("y1", d => d.source.y)
                .attr("x2", d => d.target.x)
                .attr("y2", d => d.target.y);

            node
                .attr("cx", d => d.x)
                .attr("cy", d => d.y);

            label
                .attr("x", d => d.x + 10)
                .attr("y", d => d.y);
        });

        function dragstarted(event, d) {
            if (!event.active) simulation.alphaTarget(0.3).restart();
            d.fx = d.x;
            d.fy = d.y;
        }

        function dragged(event, d) {
            d.fx = event.x;
            d.fy = event.y;
        }

        function dragended(event, d) {
            if (!event.active) simulation.alphaTarget(0);
            d.fx = null;
            d.fy = null;
        }
    });