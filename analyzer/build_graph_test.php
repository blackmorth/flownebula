<?php

require __DIR__ . "/build_graph.php";

$tmpTrace = tempnam(sys_get_temp_dir(), "nebula_trace_");
$tmpOut   = tempnam(sys_get_temp_dir(), "nebula_json_");

$traceContent = <<<TXT
main a 100 10
a b 50 5
main a 200 20
TXT;

file_put_contents($tmpTrace, $traceContent);

flownebula_build_graph($tmpTrace, $tmpOut);

$json = file_get_contents($tmpOut);

if ($json === false) {
    fwrite(STDERR, "Failed to read output JSON\n");
    exit(1);
}

$data = json_decode($json, true);

if (!is_array($data)) {
    fwrite(STDERR, "Output is not valid JSON array\n");
    exit(1);
}

// On attend 2 arêtes : main->a et a->b
if (count($data) !== 2) {
    fwrite(STDERR, "Expected 2 edges, got " . count($data) . "\n");
    exit(1);
}

// Reindex by key for easier checks
$edges = [];
foreach ($data as $e) {
    $edges[$e["caller"] . "->" . $e["callee"]] = $e;
}

assertEdge($edges, "main->a", 2, 300, 30);
assertEdge($edges, "a->b", 1, 50, 5);

echo "build_graph.php tests passed.\n";

function assertEdge(array $edges, string $key, int $calls, int $time, int $mem): void
{
    if (!isset($edges[$key])) {
        fwrite(STDERR, "Missing edge {$key}\n");
        exit(1);
    }

    $e = $edges[$key];

    if ($e["calls"] !== $calls) {
        fwrite(STDERR, "Edge {$key}: expected calls={$calls}, got {$e["calls"]}\n");
        exit(1);
    }

    if ($e["time"] !== $time) {
        fwrite(STDERR, "Edge {$key}: expected time={$time}, got {$e["time"]}\n");
        exit(1);
    }

    if ($e["mem"] !== $mem) {
        fwrite(STDERR, "Edge {$key}: expected mem={$mem}, got {$e["mem"]}\n");
        exit(1);
    }
}

