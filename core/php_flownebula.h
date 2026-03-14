/*
 * FlowNebula – PHP profiler header
 * Depends only on the public PHP/Zend extension API.
 * Target: PHP 8.1+
 */
#ifndef PHP_FLOWNEBULA_H
#define PHP_FLOWNEBULA_H

#ifdef HAVE_CONFIG_H
# include "config.h"
#endif

#include <php.h>
#include <Zend/zend_types.h>
#include <Zend/zend_execute.h>
#include <Zend/zend_alloc.h>
#include <Zend/zend_API.h>
#include <Zend/zend_ini.h>

#if PHP_VERSION_ID < 80100
# error "FlowNebula requires PHP 8.1 or later"
#endif

/* -------------------------------------------------------------------------
 * Compatibility macros (insulate from minor API renames)
 * ------------------------------------------------------------------------- */
#ifndef FLOWNEBULA_ZSTR_VAL
# define FLOWNEBULA_ZSTR_VAL(s)  ZSTR_VAL(s)
#endif
#ifndef FLOWNEBULA_ZSTR_LEN
# define FLOWNEBULA_ZSTR_LEN(s)   ZSTR_LEN(s)
#endif

/* -------------------------------------------------------------------------
 * Extension identity
 * ------------------------------------------------------------------------- */
#define PHP_FLOWNEBULA_EXTNAME   "flownebula"
#define PHP_FLOWNEBULA_VERSION  "0.1"

/* -------------------------------------------------------------------------
 * Module globals (single struct, no internal layout dependency)
 * ------------------------------------------------------------------------- */
ZEND_BEGIN_MODULE_GLOBALS(flownebula)
    zend_string *trace_path;
ZEND_END_MODULE_GLOBALS(flownebula)

ZEND_EXTERN_MODULE_GLOBALS(flownebula)
#define FLOWNEBULA_G(v)  ZEND_MODULE_GLOBALS_ACCESSOR(flownebula, v)

/* -------------------------------------------------------------------------
 * Module entry (required by PHP)
 * ------------------------------------------------------------------------- */
extern zend_module_entry flownebula_module_entry;
#define phpext_flownebula_ptr  (&flownebula_module_entry)

/* -------------------------------------------------------------------------
 * Lifecycle hooks
 * ------------------------------------------------------------------------- */
PHP_MINIT_FUNCTION(flownebula);
PHP_MSHUTDOWN_FUNCTION(flownebula);
PHP_RINIT_FUNCTION(flownebula);
PHP_RSHUTDOWN_FUNCTION(flownebula);
PHP_MINFO_FUNCTION(flownebula);

/* -------------------------------------------------------------------------
 * Executor hook (single callback, no internal executor layout)
 * ------------------------------------------------------------------------- */
typedef void (*nebula_execute_ex_fn)(zend_execute_data *execute_data);
extern nebula_execute_ex_fn original_execute_ex;
void nebula_execute_ex(zend_execute_data *execute_data);

/* -------------------------------------------------------------------------
 * Trace I/O (opaque; implementation in .c)
 * ------------------------------------------------------------------------- */
void nebula_trace_open(void);
void nebula_trace_close(void);

/* -------------------------------------------------------------------------
 * High-resolution timer (nanoseconds, monotonic)
 * ------------------------------------------------------------------------- */
uint64_t nebula_time(void);

/* -------------------------------------------------------------------------
 * Stack frame (plain struct, fixed-size types only)
 * ------------------------------------------------------------------------- */
#define NEBULA_MAX_STACK  1024

typedef struct _nebula_frame {
    const char *function;   /* name only, no pointer to Zend internals */
    uint64_t    start_time; /* nanoseconds */
    size_t      start_memory;
} nebula_frame;

#endif /* PHP_FLOWNEBULA_H */
