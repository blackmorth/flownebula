#include "nebula_probe.h"
#include <sys/socket.h>

ZEND_DECLARE_MODULE_GLOBALS(nebula_probe)

static void php_nebula_probe_init_globals(zend_nebula_probe_globals *g)
{
    memset(g, 0, sizeof(*g));
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

PHP_MINIT_FUNCTION(nebula_probe)
{
    REGISTER_INI_ENTRIES();
    #include <sys/un.h>

    NEBULA_G(udp_fd) = socket(AF_UNIX, SOCK_DGRAM, 0);
    if (NEBULA_G(udp_fd) > 0) {
        struct sockaddr_un *addr = &NEBULA_G(agent_addr_un);
        memset(addr, 0, sizeof(*addr));
        addr->sun_family = AF_UNIX;
        strncpy(addr->sun_path, "/var/run/nebula.sock", sizeof(addr->sun_path)-1);
    }
    if (NEBULA_G(udp_fd) > 0) {
        memset(&NEBULA_G(agent_addr), 0, sizeof(NEBULA_G(agent_addr)));
        NEBULA_G(agent_addr).sin_family = AF_INET;
        NEBULA_G(agent_addr).sin_port = htons(NEBULA_G(agent_port));
        inet_pton(AF_INET, NEBULA_G(agent_host), &NEBULA_G(agent_addr).sin_addr);
    }
    zend_hash_init(&NEBULA_G(func_map), 1024, NULL, NULL, 1);
    NEBULA_G(next_func_id) = 1;
    atomic_store(&NEBULA_G(write_pos), 0);
    old_execute_ex = zend_execute_ex;
    zend_execute_ex = nebula_execute_ex;
    old_execute_internal = zend_execute_internal;
    zend_execute_internal = nebula_execute_internal;
    NEBULA_G(session_id_ptr) = malloc(SESSION_ID_SIZE);
    return SUCCESS;
}

PHP_MSHUTDOWN_FUNCTION(nebula_probe)
{
    UNREGISTER_INI_ENTRIES();
    if (NEBULA_G(udp_fd) > 0) {
        close(NEBULA_G(udp_fd));
    }
    zend_hash_destroy(&NEBULA_G(func_map));
    zend_execute_ex = old_execute_ex;
    zend_execute_internal = old_execute_internal;
    if (NEBULA_G(session_id_ptr)) free(NEBULA_G(session_id_ptr));
    return SUCCESS;
}

PHP_RINIT_FUNCTION(nebula_probe)
{
#if defined(ZTS) && defined(COMPILE_DL_FLOW_NEBULA)
    ZEND_TSRMLS_CACHE_UPDATE();
#endif
    NEBULA_G(depth) = 0;
    atomic_store(&NEBULA_G(write_pos), 0);
    generate_session_id(NEBULA_G(session_id_ptr));
    return SUCCESS;
}

PHP_RSHUTDOWN_FUNCTION(nebula_probe)
{
    flush_buffer();
    return SUCCESS;
}

zend_module_entry nebula_probe_module_entry = {
    STANDARD_MODULE_HEADER,
    "flow_nebula",
    NULL,
    PHP_MINIT(nebula_probe),
    PHP_MSHUTDOWN(nebula_probe),
    PHP_RINIT(nebula_probe),
    PHP_RSHUTDOWN(nebula_probe),
    NULL,
    "0.1.0",
    STANDARD_MODULE_PROPERTIES
};

#ifdef ZTS
#ifdef COMPILE_DL_FLOW_NEBULA
ZEND_TSRMLS_CACHE_DEFINE()
#endif
#endif
ZEND_GET_MODULE(nebula_probe)
