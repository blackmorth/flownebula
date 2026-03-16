#include "php_flownebula.h"

#if defined(_WIN32) || defined(_WIN64)
# include <windows.h>
# include <winsock2.h>
# include <ws2tcpip.h>
#else
# include <time.h>
# include <sys/types.h>
# include <sys/socket.h>
# include <netdb.h>
# include <unistd.h>   // close(), write()
# include <string.h>   // strncpy()
#endif

/* php_info_* may be in main/php_main.h; declare locally for portability */
PHPAPI void php_info_print_table_start(void);
PHPAPI void php_info_print_table_row(int num_cols, ...);
PHPAPI void php_info_print_table_end(void);

static FILE *trace_file = NULL;

static int agent_socket = -1;

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


// helper pour écrire ligne dans fichier ou agent
static void nebula_write_trace_line(const char *caller, const char *fname, uint64_t duration, size_t mem_delta)
{
    if (trace_file) {
        fprintf(trace_file, "%s %s %llu %zu\n", caller, fname,
            (unsigned long long)duration, mem_delta);
    } else if (agent_socket >= 0) {
        char buf[512];
        int n = snprintf(buf, sizeof(buf), "%s %s %llu %zu\n", caller, fname,
            (unsigned long long)duration, mem_delta);
#ifdef _WIN32
        send(agent_socket, buf, n, 0);
#else
        write(agent_socket, buf, n);
#endif
    }
}

/* -----------------------------
   High resolution timer (portable, no zend_hrtime dependency)
----------------------------- */

uint64_t nebula_time(void)
{
#if defined(_WIN32) || defined(_WIN64)
    LARGE_INTEGER freq, count;
    if (QueryPerformanceFrequency(&freq) && QueryPerformanceCounter(&count))
        return (uint64_t)((count.QuadPart * 1000000000ULL) / freq.QuadPart);
    return (uint64_t)GetTickCount64() * 1000000ULL;
#else
    struct timespec ts;
    if (clock_gettime(CLOCK_MONOTONIC, &ts) == 0)
        return (uint64_t)ts.tv_sec * 1000000000ULL + (uint64_t)ts.tv_nsec;
    return 0;
#endif
}


/* -----------------------------
   Trace file management
----------------------------- */

void nebula_trace_open()
{
    const char *agent_addr = getenv("FLOWNEBULA_AGENT_ADDR");
    php_error_docref(NULL, E_NOTICE, "FlowNebula: Starting trace_open. Agent addr: %s", agent_addr ? agent_addr : "NULL");
    php_error_docref(NULL, E_NOTICE, "FlowNebula: trace_path length: %zu, value: '%s'",
        FLOWNEBULA_G(trace_path) ? ZSTR_LEN(FLOWNEBULA_G(trace_path)) : 0,
        FLOWNEBULA_G(trace_path) ? ZSTR_VAL(FLOWNEBULA_G(trace_path)) : "NULL");

    // Priorité à l'agent si configuré
    if (agent_addr) {
        php_error_docref(NULL, E_NOTICE, "FlowNebula: Using agent at: %s", agent_addr);
        char host[256], port[16];
        const char *sep = strchr(agent_addr, ':');
        if (!sep) {
            php_error_docref(NULL, E_WARNING, "FlowNebula: Invalid agent address format (expected host:port)");
            return;
        }
        size_t hlen = sep - agent_addr;
        strncpy(host, agent_addr, hlen);
        host[hlen] = '\0';
        strncpy(port, sep+1, sizeof(port)-1);
        port[sizeof(port)-1] = '\0';

        struct addrinfo hints = {0}, *res;
        hints.ai_family = AF_UNSPEC;
        hints.ai_socktype = SOCK_STREAM;
        if (getaddrinfo(host, port, &hints, &res) != 0) {
            php_error_docref(NULL, E_WARNING, "FlowNebula: Failed to resolve agent address");
            return;
        }

        agent_socket = socket(res->ai_family, res->ai_socktype, res->ai_protocol);
        if (agent_socket < 0) {
            php_error_docref(NULL, E_WARNING, "FlowNebula: Failed to create socket (errno=%d)", errno);
            freeaddrinfo(res);
            return;
        }

        if (connect(agent_socket, res->ai_addr, res->ai_addrlen) < 0) {
            php_error_docref(NULL, E_WARNING, "FlowNebula: Failed to connect to agent (errno=%d)", errno);
            close(agent_socket);
            agent_socket = -1;
            // Si la connexion échoue, on pourrait basculer vers un fichier ici (optionnel)
        } else {
            php_error_docref(NULL, E_NOTICE, "FlowNebula: Connected to agent successfully!");
        }
        freeaddrinfo(res);
    }
    // Sinon, si un chemin de fichier NON VIDE est configuré, utilise-le
    else if (FLOWNEBULA_G(trace_path) && ZSTR_LEN(FLOWNEBULA_G(trace_path)) > 0) {
        const char *path = ZSTR_VAL(FLOWNEBULA_G(trace_path));
        php_error_docref(NULL, E_NOTICE, "FlowNebula: Using file path: %s", path);
        trace_file = fopen(path, "w");
        if (!trace_file) {
            php_error_docref(NULL, E_WARNING, "FlowNebula: Cannot open trace file '%s'", path);
        } else {
            setvbuf(trace_file, NULL, _IOLBF, 0);
        }
    }
    // Aucune configuration valide
    else {
        php_error_docref(NULL, E_WARNING, "FlowNebula: No valid trace configuration. Traces will be lost.");
    }
}



void nebula_trace_close()
{
    if (trace_file) { fflush(trace_file); fclose(trace_file); trace_file = NULL; }
    if (agent_socket >= 0) {
#ifdef _WIN32
        closesocket(agent_socket);
#else
        close(agent_socket);
#endif
        agent_socket = -1;
    }
}


/* -----------------------------
   Executor hook
----------------------------- */

void nebula_execute_ex(zend_execute_data *execute_data)
{
    if (stack_top >= NEBULA_MAX_STACK) {
        php_error_docref(NULL, E_WARNING, "FlowNebula: stack overflow, increase NEBULA_MAX_STACK");
        original_execute_ex(execute_data);
        return;
    }

    const zend_function *func = execute_data->func;

    const char *fname = "main";

    if (func->common.scope) {
    // Cas d'une méthode de classe
    fname = "class_method";
    } else if (func->common.function_name) {
        fname = ZSTR_VAL(func->common.function_name);
    } else {
        fname = "main";
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

    nebula_write_trace_line(caller, fname, duration, mem_delta);
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