<?php

function flownebula_build_graph(string $trace, string $out): void
{
    $edges = [];

    $fh = fopen($trace, "r");
    if ($fh === false) {
        fwrite(STDERR, "Unable to open trace file: {$trace}\n");
        exit(1);
    }

    while (($line = fgets($fh)) !== false) {

        $line = trim($line);
        if ($line === "") {
            continue;
        }

        $parts = explode(" ", $line);

        if (count($parts) < 3) {
            // ligne invalide, on ignore
            continue;
        }

        $caller = $parts[0];
        $callee = $parts[1];
        $time   = (int) $parts[2];
        $mem    = 0;

        if (count($parts) >= 4) {
            $mem = (int) $parts[3];
        }

        $key = $caller . "->" . $callee;

        if (!isset($edges[$key])) {
            $edges[$key] = [
                "caller" => $caller,
                "callee" => $callee,
                "calls"  => 0,
                "time"   => 0,
                "mem"    => 0,
            ];
        }

        $edges[$key]["calls"]++;
        $edges[$key]["time"] += $time;
        $edges[$key]["mem"]  += $mem;
    }

    fclose($fh);

    file_put_contents(
        $out,
        json_encode(array_values($edges), JSON_PRETTY_PRINT)
    );
}

if (PHP_SAPI === 'cli' && realpath($argv[0]) === __FILE__) {
    $trace = $argv[1] ?? "/tmp/nebula.trace";
    $out   = $argv[2] ?? "nebula.json";

    flownebula_build_graph($trace, $out);
    echo "Graph written to $out\n";
}