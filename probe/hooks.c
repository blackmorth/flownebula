// hooks.c (version low-overhead)

#include "nebula_probe.h"
#include <time.h>

void (*old_execute_ex)(zend_execute_data *execute_data);
void (*old_execute_internal)(zend_execute_data *execute_data, zval *return_value);

static inline uint64_t get_cpu_time_thread(void)
{
    struct timespec ts;
    clock_gettime(CLOCK_THREAD_CPUTIME_ID, &ts);
    return (uint64_t)ts.tv_sec * 1000000000ULL + (uint64_t)ts.tv_nsec;
}

void nebula_execute_ex(zend_execute_data *execute_data)
{
    if (UNEXPECTED(!NEBULA_G(enabled) || !execute_data->func)) {
        old_execute_ex(execute_data);
        return;
    }
    if (UNEXPECTED(NEBULA_G(depth) >= NEBULA_STACK_SIZE)) {
        old_execute_ex(execute_data);
        return;
    }

    const zend_function *func = execute_data->func;
    uint32_t func_id = get_func_id(func);
    frame_t *f = &NEBULA_G(stack)[NEBULA_G(depth)++];

    f->func           = func;
    f->start_time     = zend_hrtime_nebula();
    f->child_time     = 0;
    f->cpu_start      = get_cpu_time_thread();
    f->cpu_child_time = 0;
    f->start_mem      = zend_memory_usage(0);
    f->peak_mem_start = zend_memory_peak_usage(0);

    // enter event minimal : juste pour stack agent
    emit_call(0, func_id, 0, 0, 0, 0, 0, 0, 0);

    old_execute_ex(execute_data);

    uint64_t end_time = zend_hrtime_nebula();
    uint64_t cpu_end  = get_cpu_time_thread();
    size_t   end_mem  = zend_memory_usage(0);
    size_t   peak_mem = zend_memory_peak_usage(0);

    NEBULA_G(depth)--;

    uint64_t inclusive = end_time - f->start_time;
    uint64_t exclusive = inclusive - f->child_time;

    uint64_t cpu_total = cpu_end - f->cpu_start;
    uint64_t cpu_excl  = cpu_total - f->cpu_child_time;

    int64_t  mem_delta = (int64_t)end_mem - (int64_t)f->start_mem;

    if (NEBULA_G(depth) > 0) {
        frame_t *parent = &NEBULA_G(stack)[NEBULA_G(depth) - 1];
        parent->child_time     += inclusive;
        parent->cpu_child_time += cpu_total;
    }

    // exit complet : wall, cpu_excl, mem, peak
    emit_call(1, func_id, inclusive, exclusive, cpu_excl, mem_delta, (uint64_t)peak_mem, 0, 0);
}

void nebula_execute_internal(zend_execute_data *execute_data, zval *return_value)
{
    if (UNEXPECTED(!NEBULA_G(enabled) || !execute_data->func)) {
        if (old_execute_internal) old_execute_internal(execute_data, return_value);
        else execute_internal(execute_data, return_value);
        return;
    }
    if (UNEXPECTED(NEBULA_G(depth) >= NEBULA_STACK_SIZE)) {
        if (old_execute_internal) old_execute_internal(execute_data, return_value);
        else execute_internal(execute_data, return_value);
        return;
    }

    const zend_function *func = execute_data->func;
    uint32_t func_id = get_func_id(func);
    frame_t *f = &NEBULA_G(stack)[NEBULA_G(depth)++];

    f->func           = func;
    f->start_time     = zend_hrtime_nebula();
    f->child_time     = 0;
    f->cpu_start      = get_cpu_time_thread();
    f->cpu_child_time = 0;
    f->start_mem      = zend_memory_usage(0);
    f->peak_mem_start = zend_memory_peak_usage(0);

    emit_call(0, func_id, 0, 0, 0, 0, 0, 0, 0);

    if (old_execute_internal) old_execute_internal(execute_data, return_value);
    else execute_internal(execute_data, return_value);

    uint64_t end_time = zend_hrtime_nebula();
    uint64_t cpu_end  = get_cpu_time_thread();
    size_t   end_mem  = zend_memory_usage(0);
    size_t   peak_mem = zend_memory_peak_usage(0);

    NEBULA_G(depth)--;

    uint64_t inclusive = end_time - f->start_time;
    uint64_t exclusive = inclusive - f->child_time;

    uint64_t cpu_total = cpu_end - f->cpu_start;
    uint64_t cpu_excl  = cpu_total - f->cpu_child_time;

    int64_t  mem_delta = (int64_t)end_mem - (int64_t)f->start_mem;

    if (NEBULA_G(depth) > 0) {
        frame_t *parent = &NEBULA_G(stack)[NEBULA_G(depth) - 1];
        parent->child_time     += inclusive;
        parent->cpu_child_time += cpu_total;
    }

    emit_call(1, func_id, inclusive, exclusive, cpu_excl, mem_delta, (uint64_t)peak_mem, 0, 0);
}
