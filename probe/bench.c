#include "nebula_probe.h"

PHP_FUNCTION(nebula_probe_bench)
{
    zend_long iterations = 10000000; // 10M par défaut
    if (zend_parse_parameters(ZEND_NUM_ARGS(), "|l", &iterations) == FAILURE) {
        RETURN_FALSE;
    }

    // On force un session_id pour éviter les memcpy aléatoires
    if (!NEBULA_G(session_id_ptr)) {
        NEBULA_G(session_id_ptr) = calloc(1, SESSION_ID_SIZE);
    }

    // Reset du ring
    atomic_store(&NEBULA_G(write_pos), 0);
    atomic_store(&NEBULA_G(overflow_count), 0);
    atomic_store(&NEBULA_G(high_watermark), 0);

    uint64_t start = zend_hrtime_nebula();

    for (zend_long i = 0; i < iterations; i++) {

        // ENTER
        emit_call(
            0,          // event_type
            42,         // func_id
            i,          // inclusive -> ts_ns
            0,          // exclusive -> depth
            0,          // cpu_time
            0,          // mem_delta
            0,          // peak_memory
            0,          // io_wait
            0,          // network
            0          // event_time_unix_ns
        );

        // EXIT
        emit_call(
            1,
            42,
            i,
            0,
            0,0,0,0,0,
            0
        );
    }

    uint64_t end = zend_hrtime_nebula();
    double seconds = (double)(end - start) / 1e9;

    double ev_per_sec = (iterations * 2) / seconds;

    array_init(return_value);
    add_assoc_double(return_value, "events_per_second", ev_per_sec);
    add_assoc_double(return_value, "seconds", seconds);
    add_assoc_long(return_value, "iterations", iterations);
    add_assoc_long(return_value, "overflow", atomic_load(&NEBULA_G(overflow_count)));
    add_assoc_long(return_value, "high_watermark", atomic_load(&NEBULA_G(high_watermark)));
}
