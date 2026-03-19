#ifndef NEBULA_PROBE_H
#define NEBULA_PROBE_H

#include "php.h"
#include "zend_API.h"
#include "zend_extensions.h"
#include "zend_exceptions.h"
#include "zend_compile.h"
#include <stdint.h>
#include <stdatomic.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <time.h>
#include "nebula_protocol.h"

#define NEBULA_STACK_SIZE 256
#define NEBULA_RING_SIZE  65536
#define NEBULA_BATCH_SIZE 20

typedef struct frame_t {
    const zend_function *func;
    uint64_t start_time;
    uint64_t child_time;
    uint64_t cpu_start;
    uint64_t cpu_child_time;
    size_t   start_mem;
    size_t   peak_mem_start;
    uint64_t io_start;
    uint64_t nw_start;
} frame_t;

struct sockaddr_un agent_addr_un;

ZEND_BEGIN_MODULE_GLOBALS(nebula_probe)
    zend_bool enabled;
    uint64_t  threshold_ns;
    char     *agent_host;
    long      agent_port;
    int       depth;
    frame_t   stack[NEBULA_STACK_SIZE];
    nebula_event_t       buffer[NEBULA_RING_SIZE];
    atomic_uint_fast32_t write_pos;
    int                  udp_fd;
    struct sockaddr_in   agent_addr;
    uint32_t             next_func_id;
    HashTable            func_map;
    char                *session_id_ptr;
ZEND_END_MODULE_GLOBALS(nebula_probe)

extern zend_nebula_probe_globals nebula_probe_globals;

#ifdef ZTS
#define NEBULA_G(v) ZEND_TSRMG(nebula_probe_globals_id, zend_nebula_probe_globals *, v)
#else
#define NEBULA_G(v) (nebula_probe_globals.v)
#endif

#ifndef UNEXPECTED
#define UNEXPECTED(condition) (condition)
#endif

#if PHP_VERSION_ID < 80300
static inline uint64_t zend_hrtime_nebula(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (uint64_t)ts.tv_sec * 1000000000ULL + (uint64_t)ts.tv_nsec;
}
#else
#include "zend_hrtime.h"
#define zend_hrtime_nebula() zend_hrtime()
#endif

/* Utils */
uint64_t get_cpu_time(void);
uint64_t get_io_wait(void);
uint64_t get_nw_usage(void);
uint32_t get_func_id(const zend_function *func);
void generate_session_id(char out[SESSION_ID_SIZE]);
void send_func_name(uint32_t func_id, const char *name);
void emit_call(uint8_t event_type, uint32_t func_id, uint64_t inclusive, uint64_t exclusive, uint64_t cpu_time, int64_t mem_delta, uint64_t peak_memory, uint64_t io_wait, uint64_t network);
void flush_buffer(void);

/* Hooks */
extern void (*old_execute_ex)(zend_execute_data *execute_data);
extern void (*old_execute_internal)(zend_execute_data *execute_data, zval *return_value);
void nebula_execute_ex(zend_execute_data *execute_data);
void nebula_execute_internal(zend_execute_data *execute_data, zval *return_value);

#endif /* NEBULA_PROBE_H */
