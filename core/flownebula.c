#include "php_flownebula.h"
#include <stdio.h>
#include <stdlib.h>

static FILE *trace_file = NULL;

void (*original_execute_ex)(zend_execute_data *execute_data);

static nebula_frame stack[NEBULA_MAX_STACK];
static int stack_top = 0;

ZEND_DECLARE_MODULE_GLOBALS(flownebula)

static void php_flownebula_init_globals(zend_flownebula_globals *flownebula_globals)
{
    flownebula_globals->trace_path = NULL;
}

PHP_INI_BEGIN()
    STD_PHP_INI_ENTRY(
        "flownebula.trace_path",
        "",
        PHP_INI_ALL,
        OnUpdateString,
        trace_path,
        zend_flownebula_globals,
        flownebula_globals
    )
PHP_INI_END()


/* -----------------------------
   High resolution timer
----------------------------- */

uint64_t nebula_time(void)
{
    return (uint64_t) zend_hrtime(true);
}


/* -----------------------------
   Trace file management
----------------------------- */

void nebula_trace_open()
{
    const char *path = NULL;

    if (FLOWNEBULA_G(trace_path) && ZSTR_LEN(FLOWNEBULA_G(trace_path)) > 0) {
        path = ZSTR_VAL(FLOWNEBULA_G(trace_path));
    } else {
        const char *env = getenv("FLOWNEBULA_TRACE");

        if (env && env[0] != '\0') {
            path = env;
        } else {
#ifdef _WIN32
            path = "C:\\tmp\\nebula.trace";
#else
            path = "/tmp/nebula.trace";
#endif
        }
    }

    trace_file = fopen(path, "w");

    if (!trace_file) {
        php_error_docref(NULL, E_WARNING,
            "FlowNebula: failed to open trace file '%s' for writing",
            path
        );
    }
}

void nebula_trace_close()
{
    if (trace_file) {
        fclose(trace_file);
        trace_file = NULL;
    }
}


/* -----------------------------
   Executor hook
----------------------------- */

void nebula_execute_ex(zend_execute_data *execute_data)
{
    if (stack_top >= NEBULA_MAX_STACK) {
        original_execute_ex(execute_data);
        return;
    }

    const zend_function *func = execute_data->func;

    const char *fname = "main";

    if (func->common.function_name) {
        fname = ZSTR_VAL(func->common.function_name);
    }

    uint64_t start = nebula_time();
    size_t   mem_before = zend_memory_usage(0);

    stack[stack_top].function      = fname;
    stack[stack_top].start_time    = start;
    stack[stack_top].start_memory  = mem_before;
    stack_top++;

    original_execute_ex(execute_data);

    uint64_t end = nebula_time();
    uint64_t duration = end - start;
    size_t   mem_after = zend_memory_usage(0);

    if (stack_top > 0) {
        stack_top--;
    } else {
        stack_top = 0;
    }

    size_t mem_delta = 0;
    if (mem_after > mem_before) {
        mem_delta = mem_after - mem_before;
    }

    const char *caller = "main";

    if (stack_top > 0) {
        caller = stack[stack_top - 1].function;
    }

    if (trace_file) {
        fprintf(
            trace_file,
            "%s %s %llu %zu\n",
            caller,
            fname,
            (unsigned long long) duration,
            (size_t) mem_delta
        );
    }
}


/* -----------------------------
   PHP lifecycle
----------------------------- */

PHP_MINIT_FUNCTION(flownebula)
{
    ZEND_INIT_MODULE_GLOBALS(flownebula, php_flownebula_init_globals, NULL);
    REGISTER_INI_ENTRIES();

    original_execute_ex = zend_execute_ex;
    zend_execute_ex = nebula_execute_ex;

    return SUCCESS;
}

PHP_MSHUTDOWN_FUNCTION(flownebula)
{
    zend_execute_ex = original_execute_ex;
    UNREGISTER_INI_ENTRIES();

    return SUCCESS;
}

PHP_RINIT_FUNCTION(flownebula)
{
#if defined(ZTS) && defined(COMPILE_DL_FLOWNEBULA)
    ZEND_TSRMLS_CACHE_UPDATE();
#endif

    nebula_trace_open();

    return SUCCESS;
}

PHP_RSHUTDOWN_FUNCTION(flownebula)
{
    nebula_trace_close();

    return SUCCESS;
}

PHP_MINFO_FUNCTION(flownebula)
{
    php_info_print_table_start();
    php_info_print_table_row(2, "FlowNebula profiler", "enabled");
    php_info_print_table_row(2, "version", PHP_FLOWNEBULA_VERSION);
    php_info_print_table_end();

    DISPLAY_INI_ENTRIES();
}


/* -----------------------------
   Module entry
----------------------------- */

zend_module_entry flownebula_module_entry = {
    STANDARD_MODULE_HEADER,
    PHP_FLOWNEBULA_EXTNAME,
    NULL,
    PHP_MINIT(flownebula),
    PHP_MSHUTDOWN(flownebula),
    PHP_RINIT(flownebula),
    PHP_RSHUTDOWN(flownebula),
    PHP_MINFO(flownebula),
    PHP_FLOWNEBULA_VERSION,
    STANDARD_MODULE_PROPERTIES
};

#ifdef COMPILE_DL_FLOWNEBULA
# ifdef ZTS
ZEND_TSRMLS_CACHE_DEFINE()
# endif
ZEND_GET_MODULE(flownebula)
#endif