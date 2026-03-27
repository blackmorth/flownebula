<?php
declare(strict_types=1);

/**
 * Script de validation perf/fonctionnement
 * Exécution:
 *   php -d xdebug.mode=off validate.php
 */

$barCalls = 0;
$maxDepth = 0;

function foo(): void
{
    // 10 000 microsecondes = 10 ms
    usleep(10000);
}

function bar(int $n, int $depth = 1): void
{
    global $barCalls, $maxDepth;
    $barCalls++;
    if ($depth > $maxDepth) {
        $maxDepth = $depth;
    }

    if ($n <= 0) {
        return;
    }
    bar($n - 1, $depth + 1);
}

function nsToMs(float $ns): float
{
    return $ns / 1_000_000.0;
}

// ------------------------------------------------------------
// 1) Mesure de foo/usleep
// ------------------------------------------------------------
$t0 = hrtime(true);
foo();
$t1 = hrtime(true);

$fooNs = $t1 - $t0;
$fooMs = nsToMs($fooNs);

// ------------------------------------------------------------
// 2) Mesure de bar(125)
// ------------------------------------------------------------
$n = 125;
$expectedCalls = $n + 1; // de 125 à 0 inclus

$t2 = hrtime(true);
bar($n);
$t3 = hrtime(true);

$barNs = $t3 - $t2;
$barMs = nsToMs($barNs);

// ------------------------------------------------------------
// 3) (Optionnel) Profil xhprof si dispo
// ------------------------------------------------------------
$xhprofAvailable = function_exists('xhprof_enable') && function_exists('xhprof_disable');
$xhprofSummary = null;

if ($xhprofAvailable) {
    xhprof_enable(XHPROF_FLAGS_CPU | XHPROF_FLAGS_MEMORY);

    foo();
    bar($n);

    $profile = xhprof_disable();

    // Cherche les clés utiles (format souvent "caller==>callee")
    $keys = array_keys($profile);
    $barSelfCt = 0;
    $barFromRootCt = 0;
    $fooFromRootWt = null;
    $sleepWt = null;

    foreach ($profile as $edge => $cost) {
        // Exemples de clés:
        // "main()==>foo", "foo==>usleep", "main()==>bar", "bar==>bar"
        if (strpos($edge, 'bar==>bar') !== false) {
            $barSelfCt = (int)($cost['ct'] ?? 0);
        }
        if (strpos($edge, 'main()==>bar') !== false || strpos($edge, 'test.php==>bar') !== false) {
            $barFromRootCt = (int)($cost['ct'] ?? 0);
        }
        if (strpos($edge, 'main()==>foo') !== false || strpos($edge, 'test.php==>foo') !== false) {
            $fooFromRootWt = $cost['wt'] ?? null;
        }
        if (strpos($edge, 'foo==>usleep') !== false) {
            $sleepWt = $cost['wt'] ?? null;
        }
    }

    $xhprofSummary = [
        'edges_count' => count($keys),
        'bar_total_ct_approx' => $barSelfCt + $barFromRootCt,
        'bar_self_ct' => $barSelfCt,
        'bar_root_ct' => $barFromRootCt,
        'foo_wt_raw' => $fooFromRootWt,
        'usleep_wt_raw' => $sleepWt,
        'note' => 'Selon l’outil/UI, wt peut être en µs ou en ns (vérifier doc de ta stack).',
    ];
}

// ------------------------------------------------------------
// 4) Rapport
// ------------------------------------------------------------
echo "=== Validation locale ===\n";
echo "foo(usleep 10000) : " . number_format($fooMs, 3) . " ms\n";
echo "bar($n) : " . number_format($barMs, 3) . " ms\n";
echo "bar calls : $barCalls (attendu: $expectedCalls)\n";
echo "max depth : $maxDepth (attendu: $expectedCalls)\n";

echo "\n=== Verdict ===\n";
echo ($barCalls === $expectedCalls ? "[OK] Nombre d'appels bar correct\n" : "[KO] Nombre d'appels bar incorrect\n");
echo ($maxDepth === $expectedCalls ? "[OK] Profondeur max correcte\n" : "[KO] Profondeur max incorrecte\n");
echo ($fooMs >= 8.0 && $fooMs <= 30.0
    ? "[OK] usleep ~10ms (tolérance large)\n"
    : "[WARN] usleep hors plage attendue (charge machine possible)\n");

if ($xhprofSummary !== null) {
    echo "\n=== xhprof (optionnel) ===\n";
    echo json_encode($xhprofSummary, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES) . "\n";
} else {
    echo "\n[xhprof] Extension non disponible (normal si non installée).\n";
}

echo "\nOK\n";