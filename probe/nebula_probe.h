// nebula_probe.h

#ifndef NEBULA_PROBE_H
#define NEBULA_PROBE_H

#include <stdint.h>
#include "php.h"
#include "nebula_probe_globals.h"
#include "nebula_protocol.h"

void emit_call(uint8_t kind,
               uint32_t func_id,
               uint64_t ts_ns,
               uint32_t depth,
               uint8_t  func_type,
               uint8_t  flags,
               uint8_t  arg_count,
               uint8_t  has_exception,
               uint8_t  jit_flag,
               uint64_t unix_time_ns);

PHP_FUNCTION(nebula_probe_bench);

#endif

extern zend_nebula_probe_globals nebula_probe_globals;