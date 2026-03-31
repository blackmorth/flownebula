#include "nebula_probe.h"
#include <time.h>


/*
 * Mapping vers emit_call :
 *
 * emit_call(
 *   event_type,          // kind: 0=ENTER,1=EXIT
 *   func_id,             // func_id
 *   inclusive,           // ts_ns (zend_hrtime_nebula)
 *   exclusive,           // depth
 *   cpu_time,            // func_type
 *   mem_delta,           // flags
 *   peak_memory,         // arg_count
 *   io_wait,             // has_exception
 *   network,             // jit_flag
 *   event_time_unix_ns,  // 0 => rempli dans emit_call
 *   alloc_bytes,         // 0
 *   free_bytes           // 0
 * );
 */

static inline void nebula_emit_event(
    uint8_t kind,
    const zend_function *func,
    zend_execute_data *execute_data,
    uint32_t func_id,
    uint32_t depth,
    uint64_t ts_ns
) {
    uint8_t func_type   = func->type;
    uint8_t flags       = 0;
    uint8_t arg_count   = 0;
    uint8_t has_exc     = EG(exception) ? 1 : 0;
    uint8_t jit_flag    = 0;

    if (func->common.fn_flags & ZEND_ACC_STATIC)    flags |= 1;
    if (func->common.fn_flags & ZEND_ACC_CLOSURE)   flags |= 2;
    if (func->common.fn_flags & ZEND_ACC_GENERATOR) flags |= 4;

    if (execute_data) {
        arg_count = (uint8_t)ZEND_CALL_NUM_ARGS(execute_data);
    }

#ifdef ZEND_JIT_ENABLED
    if (func->type == ZEND_USER_FUNCTION &&
        (func->op_array.fn_flags & ZEND_ACC_JIT)) {
        jit_flag = 1;
    }
#endif
    emit_call(
        kind,              // event_type
        func_id,           // func_id
        ts_ns,             // inclusive -> ts_ns
        depth,             // exclusive -> depth
        func_type,         // cpu_time   -> func_type
        (int64_t)flags,    // mem_delta  -> flags
        arg_count,         // peak_memory-> arg_count
        has_exc,           // io_wait    -> has_exception
        jit_flag,          // network    -> jit_flag
        0                 // event_time_unix_ns (rempli dans emit_call)
    );
}

void nebula_execute_ex(zend_execute_data *execute_data)
{
    if (UNEXPECTED(!NEBULA_G(enabled) || !execute_data || !execute_data->func)) {
        old_execute_ex(execute_data);
        return;
    }
    if (UNEXPECTED(NEBULA_G(depth) >= NEBULA_STACK_SIZE)) {
        old_execute_ex(execute_data);
        return;
    }

    const zend_function *func = execute_data->func;
    uint32_t func_id = get_func_id(func);
    if (!func_id) {
        old_execute_ex(execute_data);
        return;
    }

    uint32_t depth = (uint32_t)NEBULA_G(depth)++;
    uint64_t ts_enter = zend_hrtime_nebula();

    nebula_emit_event(
        0,          // ENTER
        func,
        execute_data,
        func_id,
        depth,
        ts_enter
    );

    old_execute_ex(execute_data);

    uint64_t ts_exit = zend_hrtime_nebula();
    NEBULA_G(depth)--;

    nebula_emit_event(
        1,          // EXIT
        func,
        execute_data,
        func_id,
        depth,
        ts_exit
    );
}

void nebula_execute_internal(zend_execute_data *execute_data, zval *return_value)
{
    if (!execute_data || !execute_data->func || !NEBULA_G(enabled)) {
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
    if (!func_id) {
        if (old_execute_internal) old_execute_internal(execute_data, return_value);
        else execute_internal(execute_data, return_value);
        return;
    }

    uint32_t depth = (uint32_t)NEBULA_G(depth)++;
    uint64_t ts_enter = zend_hrtime_nebula();

    nebula_emit_event(
        0,          // ENTER
        func,
        execute_data,
        func_id,
        depth,
        ts_enter
    );

    if (old_execute_internal) old_execute_internal(execute_data, return_value);
    else execute_internal(execute_data, return_value);

    uint64_t ts_exit = zend_hrtime_nebula();
    NEBULA_G(depth)--;

    nebula_emit_event(
        1,          // EXIT
        func,
        execute_data,
        func_id,
        depth,
        ts_exit
    );
}
