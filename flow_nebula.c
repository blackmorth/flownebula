#include "php.h"
#include "zend_compile.h"
#include "zend_extensions.h"
#include "zend_API.h"
#include "zend_exceptions.h"

#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <arpa/inet.h>
#include <unistd.h>
#include <time.h>
#include <sys/time.h>
#include <stdatomic.h>

#define NEBULA_STACK_SIZE 256
#define NEBULA_RING_SIZE  65536
#define SESSION_ID_SIZE   8

typedef struct frame_t {
    const zend_function *func;
    uint64_t start_time;
    uint64_t child_time;
    size_t   start_mem;
} frame_t;

typedef struct __attribute__((packed)) call_event_t {
    char     session_id[SESSION_ID_SIZE];
    uint8_t  event_type;   /* 0 = enter, 1 = exit */
    uint32_t func_id;
    uint64_t inclusive;
    uint64_t exclusive;
    int64_t  mem_delta;
} call_event_t;

ZEND_BEGIN_MODULE_GLOBALS(nebula_probe)
    zend_bool enabled;
    uint64_t  threshold_ns;

    int       depth;
    frame_t   stack[NEBULA_STACK_SIZE];

    call_event_t         buffer[NEBULA_RING_SIZE];
    atomic_uint_fast32_t write_pos;

    int                  udp_fd;
    struct sockaddr_in   agent_addr;

    uint32_t             next_func_id;
    HashTable            func_map;
    char                *func_names[65536];

    char                 session_id[SESSION_ID_SIZE];
ZEND_END_MODULE_GLOBALS(nebula_probe)

ZEND_DECLARE_MODULE_GLOBALS(nebula_probe)
#define NEBULA_G(v) ZEND_MODULE_GLOBALS_ACCESSOR(nebula_probe, v)

static void php_nebula_probe_init_globals(zend_nebula_probe_globals *g)
{
    memset(g, 0, sizeof(*g));
    g->enabled      = 0;
    g->threshold_ns = 0;
}

PHP_INI_BEGIN()
    STD_PHP_INI_ENTRY("nebula_probe.enabled",      "0",    PHP_INI_ALL, OnUpdateBool,
                      enabled, zend_nebula_probe_globals, nebula_probe_globals)
    STD_PHP_INI_ENTRY("nebula_probe.threshold_ns", "0",    PHP_INI_ALL, OnUpdateLong,
                      threshold_ns, zend_nebula_probe_globals, nebula_probe_globals)
    STD_PHP_INI_ENTRY("nebula_probe.agent_host",   "127.0.0.1", PHP_INI_SYSTEM, OnUpdateString,
                      agent_host, zend_nebula_probe_globals, nebula_probe_globals)
    STD_PHP_INI_ENTRY("nebula_probe.agent_port",   "8135", PHP_INI_SYSTEM, OnUpdateLong,
                      agent_port, zend_nebula_probe_globals, nebula_probe_globals)
PHP_INI_END()

static void generate_session_id(char out[SESSION_ID_SIZE])
{
    struct timeval tv;
    gettimeofday(&tv, NULL);
    uint64_t val = ((uint64_t)tv.tv_sec << 32) | (uint32_t)tv.tv_usec;
    memcpy(out, &val, SESSION_ID_SIZE);
}

static uint32_t get_func_id(const zend_function *func)
{
    zend_ulong key = (zend_ulong)func;
    zval *zid = zend_hash_index_find(&NEBULA_G(func_map), key);
    if (zid) {
        return (uint32_t)Z_LVAL_P(zid);
    }

    if (NEBULA_G.next_func_id >= 65536) {
        return 0;
    }

    uint32_t id = NEBULA_G.next_func_id++;

    zval zv;
    ZVAL_LONG(&zv, (zend_long)id);
    zend_hash_index_add_new(&NEBULA_G(func_map), key, &zv);

    const char *func_name  = func->common.function_name ? ZSTR_VAL(func->common.function_name) : NULL;
    const char *class_name = (func->common.scope && func->common.scope->name)
                             ? ZSTR_VAL(func->common.scope->name)
                             : NULL;

    char tmp[256];
    if (class_name && func_name) {
        snprintf(tmp, sizeof(tmp), "%s::%s", class_name, func_name);
    } else if (func_name) {
        snprintf(tmp, sizeof(tmp), "%s", func_name);
    } else {
        snprintf(tmp, sizeof(tmp), "Closure@%p", (void *)func);
    }

    NEBULA_G.func_names[id] = estrdup(tmp);
    return id;
}

static zend_always_inline void emit_call(
    uint8_t event_type,
    uint32_t func_id,
    uint64_t inclusive,
    uint64_t exclusive,
    int64_t mem_delta
){
    if (!func_id && event_type == 1) {
        return;
    }

    uint32_t pos = atomic_fetch_add(&NEBULA_G.write_pos, 1) % NEBULA_RING_SIZE;
    call_event_t *e = &NEBULA_G.buffer[pos];

    memcpy(e->session_id, NEBULA_G.session_id, SESSION_ID_SIZE);
    e->event_type = event_type;
    e->func_id    = func_id;
    e->inclusive  = inclusive;
    e->exclusive  = exclusive;
    e->mem_delta  = mem_delta;
}

static void flush_buffer(void)
{
    uint32_t n = atomic_load(&NEBULA_G.write_pos);
    if (n == 0 || NEBULA_G.udp_fd <= 0) {
        return;
    }
    if (n > NEBULA_RING_SIZE) {
        n = NEBULA_RING_SIZE;
    }

    for (uint32_t i = 0; i < n; i++) {
        (void)sendto(
            NEBULA_G.udp_fd,
            &NEBULA_G.buffer[i],
            sizeof(call_event_t),
            MSG_DONTWAIT,
            (struct sockaddr *)&NEBULA_G.agent_addr,
            sizeof(NEBULA_G.agent_addr)
        );
    }

    atomic_store(&NEBULA_G.write_pos, 0);
}

static void (*old_execute_ex)(zend_execute_data *execute_data);
static void (*old_execute_internal)(zend_execute_data *execute_data, zval *return_value);

static void nebula_execute_ex(zend_execute_data *execute_data)
{
    if (!NEBULA_G.enabled || !execute_data->func) {
        old_execute_ex(execute_data);
        return;
    }

    int depth = NEBULA_G.depth;
    if (depth >= NEBULA_STACK_SIZE) {
        old_execute_ex(execute_data);
        return;
    }

    frame_t *f = &NEBULA_G.stack[depth++];
    f->func       = execute_data->func;
    f->start_time = zend_hrtime();
    f->child_time = 0;
    f->start_mem  = zend_memory_usage(0);

    uint32_t func_id = get_func_id(f->func);
    emit_call(0 /* enter */, func_id, 0, 0, 0);

    NEBULA_G.depth = depth;

    old_execute_ex(execute_data);

    depth = NEBULA_G.depth;
    f     = &NEBULA_G.stack[--depth];

    uint64_t total     = zend_hrtime() - f->start_time;
    uint64_t exclusive = total - f->child_time;
    int64_t  mem_delta = (int64_t)zend_memory_usage(0) - (int64_t)f->start_mem;

    if (depth > 0) {
        NEBULA_G.stack[depth - 1].child_time += total;
    }
    NEBULA_G.depth = depth;

    if (exclusive < NEBULA_G.threshold_ns) {
        return;
    }

    emit_call(1 /* exit */, func_id, total, exclusive, mem_delta);
}

static void nebula_execute_internal(zend_execute_data *execute_data, zval *return_value)
{
    if (!NEBULA_G.enabled || !execute_data->func) {
        if (old_execute_internal) {
            old_execute_internal(execute_data, return_value);
        } else {
            execute_internal(execute_data, return_value);
        }
        return;
    }

    int depth = NEBULA_G.depth;
    if (depth >= NEBULA_STACK_SIZE) {
        if (old_execute_internal) {
            old_execute_internal(execute_data, return_value);
        } else {
            execute_internal(execute_data, return_value);
        }
        return;
    }

    frame_t *f = &NEBULA_G.stack[depth++];
    f->func       = execute_data->func;
    f->start_time = zend_hrtime();
    f->child_time = 0;
    f->start_mem  = zend_memory_usage(0);

    uint32_t func_id = get_func_id(f->func);
    emit_call(0 /* enter */, func_id, 0, 0, 0);

    NEBULA_G.depth = depth;

    if (old_execute_internal) {
        old_execute_internal(execute_data, return_value);
    } else {
        execute_internal(execute_data, return_value);
    }

    depth = NEBULA_G.depth;
    f     = &NEBULA_G.stack[--depth];

    uint64_t total     = zend_hrtime() - f->start_time;
    uint64_t exclusive = total - f->child_time;
    int64_t  mem_delta = (int64_t)zend_memory_usage(0) - (int64_t)f->start_mem;

    if (depth > 0) {
        NEBULA_G.stack[depth - 1].child_time += total;
    }
    NEBULA_G.depth = depth;

    if (exclusive < NEBULA_G.threshold_ns) {
        return;
    }

    emit_call(1 /* exit */, func_id, total, exclusive, mem_delta);
}

PHP_MINIT_FUNCTION(nebula_probe)
{
    ZEND_INIT_MODULE_GLOBALS(nebula_probe, php_nebula_probe_init_globals, NULL);
    REGISTER_INI_ENTRIES();

    NEBULA_G.udp_fd = socket(AF_INET, SOCK_DGRAM, 0);
    if (NEBULA_G.udp_fd >= 0) {
        memset(&NEBULA_G.agent_addr, 0, sizeof(NEBULA_G.agent_addr));
        NEBULA_G.agent_addr.sin_family = AF_INET;

        const char *host = INI_STR("nebula_probe.agent_host");
        long port        = INI_INT("nebula_probe.agent_port");
        if (port <= 0 || port > 65535) {
            port = 8135;
        }
        NEBULA_G.agent_addr.sin_port = htons((uint16_t)port);

        if (!host || inet_pton(AF_INET, host, &NEBULA_G.agent_addr.sin_addr) != 1) {
            inet_pton(AF_INET, "127.0.0.1", &NEBULA_G.agent_addr.sin_addr);
        }
    }

    zend_hash_init(&NEBULA_G.func_map, 1024, NULL, NULL, 1);
    NEBULA_G.next_func_id = 1;
    atomic_store(&NEBULA_G.write_pos, 0);
    NEBULA_G.depth = 0;

    old_execute_ex       = zend_execute_ex;
    zend_execute_ex      = nebula_execute_ex;
    old_execute_internal = zend_execute_internal;
    zend_execute_internal = nebula_execute_internal;

    return SUCCESS;
}

PHP_MSHUTDOWN_FUNCTION(nebula_probe)
{
    zend_execute_ex       = old_execute_ex;
    zend_execute_internal = old_execute_internal;

    if (NEBULA_G.udp_fd > 0) {
        close(NEBULA_G.udp_fd);
        NEBULA_G.udp_fd = -1;
    }

    for (uint32_t i = 1; i < NEBULA_G.next_func_id; i++) {
        if (NEBULA_G.func_names[i]) {
            efree(NEBULA_G.func_names[i]);
            NEBULA_G.func_names[i] = NULL;
        }
    }

    zend_hash_destroy(&NEBULA_G.func_map);
    UNREGISTER_INI_ENTRIES();

    return SUCCESS;
}

PHP_RINIT_FUNCTION(nebula_probe)
{
#if defined(ZTS) && defined(COMPILE_DL_NEBULA_PROBE)
    ZEND_TSRMLS_CACHE_UPDATE();
#endif

    NEBULA_G.depth = 0;
    atomic_store(&NEBULA_G.write_pos, 0);
    generate_session_id(NEBULA_G.session_id);

    NEBULA_G.enabled      = INI_BOOL("nebula_probe.enabled");
    NEBULA_G.threshold_ns = (uint64_t)INI_INT("nebula_probe.threshold_ns");

    return SUCCESS;
}

PHP_RSHUTDOWN_FUNCTION(nebula_probe)
{
    flush_buffer();
    return SUCCESS;
}

zend_module_entry nebula_probe_module_entry = {
    STANDARD_MODULE_HEADER,
    "nebula_probe",
    NULL,
    PHP_MINIT(nebula_probe),
    PHP_MSHUTDOWN(nebula_probe),
    PHP_RINIT(nebula_probe),
    PHP_RSHUTDOWN(nebula_probe),
    NULL,
    "0.6.0",
    STANDARD_MODULE_PROPERTIES
};

#ifdef COMPILE_DL_NEBULA_PROBE
# ifdef ZTS
ZEND_TSRMLS_CACHE_DEFINE()
# endif
ZEND_GET_MODULE(nebula_probe)
#endif