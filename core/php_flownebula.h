#ifndef PHP_FLOWNEBULA_H
#define PHP_FLOWNEBULA_H

#ifdef HAVE_CONFIG_H
#include "config.h"
#endif

#include <php.h>
#include <Zend/zend_extensions.h>
#include <Zend/zend_API.h>
#include <Zend/zend_execute.h>
#include <Zend/zend_alloc.h>
#include <Zend/zend_hrtime.h>

/* Module globals */
ZEND_BEGIN_MODULE_GLOBALS(flownebula)
    zend_string *trace_path;
ZEND_END_MODULE_GLOBALS(flownebula)

ZEND_EXTERN_MODULE_GLOBALS(flownebula)

#define FLOWNEBULA_G(v) ZEND_MODULE_GLOBALS_ACCESSOR(flownebula, v)

/* Extension name */
#define PHP_FLOWNEBULA_EXTNAME "flownebula"
#define PHP_FLOWNEBULA_VERSION "0.1"

/* Module entry */
extern zend_module_entry flownebula_module_entry;
#define phpext_flownebula_ptr &flownebula_module_entry

/* Lifecycle hooks */
PHP_MINIT_FUNCTION(flownebula);
PHP_MSHUTDOWN_FUNCTION(flownebula);
PHP_RINIT_FUNCTION(flownebula);
PHP_RSHUTDOWN_FUNCTION(flownebula);
PHP_MINFO_FUNCTION(flownebula);

/* Hooked executor */
extern void (*original_execute_ex)(zend_execute_data *execute_data);
void nebula_execute_ex(zend_execute_data *execute_data);

/* Trace management */
void nebula_trace_open();
void nebula_trace_close();

/* High resolution timer */
uint64_t nebula_time();

/* Stack frame structure */
typedef struct _nebula_frame {
    const char *function;
    uint64_t start_time;
    size_t start_memory;
} nebula_frame;

/* Stack limits */
#define NEBULA_MAX_STACK 1024

#endif