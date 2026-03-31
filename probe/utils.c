// utils.c
#include <stdint.h>
#include <time.h>
#include <stdatomic.h>
#include "nebula_protocol.h"
#include "nebula_probe.h"


#include <stdint.h>

void generate_session_id(unsigned char *buf);
void send_func_name(uint32_t func_id, const char *name);
void nebula_send_session_end(unsigned char *session_id);
void flush_buffer(void);
uint64_t zend_hrtime_nebula(void);

static uint64_t get_unix_time_ns(void)
{
    struct timespec ts;
    clock_gettime(CLOCK_REALTIME, &ts);
    return (uint64_t)ts.tv_sec * 1000000000ULL + (uint64_t)ts.tv_nsec;
}

void emit_call(uint8_t kind,
               uint32_t func_id,
               uint64_t ts_ns,
               uint32_t depth,
               uint8_t  func_type,
               uint8_t  flags,
               uint8_t  arg_count,
               uint8_t  has_exception,
               uint8_t  jit_flag,
               uint64_t unix_time_ns)
{
    if (!func_id && kind == 1) return;

    uint_fast32_t pos = atomic_fetch_add(&NEBULA_G(write_pos), 1);
    if (pos >= NEBULA_RING_SIZE) {
        atomic_fetch_add(&NEBULA_G(overflow_count), 1);
        return;
    }

    uint_fast32_t in_use = pos + 1;
    uint_fast32_t hw = atomic_load(&NEBULA_G(high_watermark));
    while (in_use > hw && !atomic_compare_exchange_weak(&NEBULA_G(high_watermark), &hw, in_use)) {
    }

    nebula_event_t *e = &NEBULA_G(buffer)[pos];
    memcpy(e->session_id, NEBULA_G(session_id_ptr), SESSION_ID_SIZE);
    e->kind        = kind;
    e->func_id     = func_id;
    e->depth       = depth;
    e->ts_ns       = ts_ns;
    e->func_type   = func_type;
    e->flags       = flags;
    e->arg_count   = arg_count;
    e->has_exception = has_exception;
    e->jit_flag    = jit_flag;
    e->unix_time_ns = unix_time_ns ? unix_time_ns : get_unix_time_ns();
}
