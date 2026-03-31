#ifndef NEBULA_PROBE_GLOBALS_H
#define NEBULA_PROBE_GLOBALS_H

#include <stdint.h>
#include <stdatomic.h>
#include <sys/un.h>
#include "zend.h"

ZEND_BEGIN_MODULE_GLOBALS(nebula_probe)
    zend_bool enabled;
    long threshold_ns;
    double sample_rate;

    int socket_fd;
    struct sockaddr_un agent_addr_un;

    HashTable func_map;
    uint32_t next_func_id;

    _Atomic uint32_t write_pos;
    _Atomic uint32_t overflow_count;
    _Atomic uint32_t flush_error_count;
    _Atomic uint32_t high_watermark;

    unsigned char *session_id_ptr;
    uint64_t request_start;
    uint32_t depth;
ZEND_END_MODULE_GLOBALS(nebula_probe)

ZEND_EXTERN_MODULE_GLOBALS(nebula_probe)
#define NEBULA_G(v) ZEND_MODULE_GLOBALS_ACCESSOR(nebula_probe, v)

#endif
