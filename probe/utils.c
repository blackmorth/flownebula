#include "nebula_probe.h"
#include <string.h>
#include <stdlib.h>
#include <sys/time.h>
#include <unistd.h>
#include <errno.h>
#include <fcntl.h>

zend_bool nebula_should_sample(uint32_t func_id, uint64_t start_time)
{
    double rate = NEBULA_G(sample_rate);
    if (rate >= 1.0) return 1;
    if (rate <= 0.0) return 0;

    uint64_t mixed = start_time ^ ((uint64_t)func_id * 11400714819323198485ull);
    mixed ^= mixed >> 33;
    mixed *= 0xff51afd7ed558ccdull;
    mixed ^= mixed >> 33;
    mixed *= 0xc4ceb9fe1a85ec53ull;
    mixed ^= mixed >> 33;

    double normalized = (double)(mixed % 1000000u) / 1000000.0;
    return normalized < rate;
}

void generate_session_id(char out[SESSION_ID_SIZE])
{
    static _Atomic uint64_t counter = 0;
    uint64_t val = 0;
    int fd = open("/dev/urandom", O_RDONLY);
    if (fd >= 0) {
        ssize_t n = read(fd, &val, sizeof(val));
        close(fd);
        if (n == (ssize_t)sizeof(val)) {
            memcpy(out, &val, SESSION_ID_SIZE);
            return;
        }
    }
    struct timeval tv;
    gettimeofday(&tv, NULL);
    val = ((uint64_t)tv.tv_usec << 32) | (uint32_t)tv.tv_sec;
    val ^= ((uint64_t)getpid() << 16);
    val ^= ++counter;
    memcpy(out, &val, SESSION_ID_SIZE);
}

void send_func_name(uint32_t func_id, const char *name)
{
    if (!name || NEBULA_G(socket_fd) <= 0) return;
    size_t name_len = strlen(name);
    if (name_len > 255) name_len = 255;
    nebula_name_t pkg;
    memcpy(pkg.session_id, NEBULA_G(session_id_ptr), SESSION_ID_SIZE);
    pkg.event_type = NEBULA_EVENT_NAME;
    pkg.func_id    = func_id;
    pkg.name_len   = (uint32_t)name_len;
    memcpy(pkg.name, name, name_len);
    // Prefer reliability over non-blocking loss for short metadata packets.
    (void)sendto(NEBULA_G(socket_fd), &pkg, 17 + name_len, 0,
                 (struct sockaddr *)&NEBULA_G(agent_addr_un), sizeof(struct sockaddr_un));
}

uint32_t get_func_id(const zend_function *func)
{
    if (UNEXPECTED(!func)) return 0;

    /* Frame interne Zend sans filename → ignorer */
    if (func->type == ZEND_USER_FUNCTION && func->op_array.filename == NULL) {
        return 0;
    }

    zend_ulong key = (zend_ulong)func;
    zval *cached = zend_hash_index_find(&NEBULA_G(func_map), key);
    if (cached) return (uint32_t)Z_LVAL_P(cached);

    char name[256];
    name[0] = '\0';

    const zend_op_array *op =
        (func->type == ZEND_USER_FUNCTION) ? &func->op_array : NULL;

    const char *func_name =
        func->common.function_name ? ZSTR_VAL(func->common.function_name) : NULL;

    const char *class_name =
        (func->common.scope && func->common.scope->name)
            ? ZSTR_VAL(func->common.scope->name)
            : NULL;

    if (!func_name && op && op->filename) {
        const char *file = ZSTR_VAL(op->filename);
        const char *base = strrchr(file, '/');
        base = base ? base + 1 : file;
        snprintf(name, sizeof(name), "%s", base);
    }
    else if (op && (func->common.fn_flags & ZEND_ACC_CLOSURE)) {
        const char *file = op->filename ? ZSTR_VAL(op->filename) : "unknown";
        const char *base = strrchr(file, '/');
        base = base ? base + 1 : file;
        snprintf(name, sizeof(name), "{closure}::%s/%u",
                 base, (unsigned)op->line_start);
    }
    else if (class_name && func_name) {
        snprintf(name, sizeof(name), "%s::%s", class_name, func_name);
    }
    else if (func_name) {
        snprintf(name, sizeof(name), "%s", func_name);
    }
    else {
        snprintf(name, sizeof(name), "internal::unknown");
    }

    uint32_t id = NEBULA_G(next_func_id)++;
    send_func_name(id, name);

    zval zv;
    ZVAL_LONG(&zv, (zend_long)id);
    zend_hash_index_add_new(&NEBULA_G(func_map), key, &zv);

    return id;
}

/*uint32_t get_func_id(const zend_function *func)
  {
      if (UNEXPECTED(!func)) return 0;
      zend_ulong key = (zend_ulong)func;
      zval *zid = zend_hash_index_find(&NEBULA_G(func_map), key);
      if (zid) return (uint32_t)Z_LVAL_P(zid);
      if (NEBULA_G(next_func_id) >= 1000000) return 0;
      uint32_t id = NEBULA_G(next_func_id)++;
      zval zv;
      ZVAL_LONG(&zv, (zend_long)id);
      zend_hash_index_add_new(&NEBULA_G(func_map), key, &zv);
      const char *func_name  = func->common.function_name ? ZSTR_VAL(func->common.function_name) : NULL;
      const char *class_name = (func->common.scope && func->common.scope->name)
                               ? ZSTR_VAL(func->common.scope->name) : NULL;
      char tmp[512];
      if (class_name && func_name) snprintf(tmp, sizeof(tmp), "%s::%s", class_name, func_name);
      else if (func_name) snprintf(tmp, sizeof(tmp), "%s", func_name);
      else snprintf(tmp, sizeof(tmp), "Closure@%p", (void *)func);
      send_func_name(id, tmp);
      return id;
  }*/

// --- fonctions neutres mais présentes pour compatibilité ---
uint64_t get_cpu_time(void) { return 0; }
uint64_t get_io_wait(void) { return 0; }
uint64_t get_nw_usage(void) { return 0; }

void emit_call(uint8_t event_type, uint32_t func_id, uint64_t inclusive, uint64_t exclusive,
               uint64_t cpu_time, int64_t mem_delta, uint64_t peak_memory, uint64_t io_wait, uint64_t network)
{
    if (!func_id && event_type == 1) return;
    uint32_t pos = atomic_fetch_add(&NEBULA_G(write_pos), 1);
    if (pos >= NEBULA_RING_SIZE) {
        atomic_fetch_add(&NEBULA_G(overflow_count), 1);
        return;
    }
    nebula_event_t *e = &NEBULA_G(buffer)[pos];
    memcpy(e->session_id, NEBULA_G(session_id_ptr), SESSION_ID_SIZE);
    e->event_type = event_type;
    e->func_id    = func_id;
    e->inclusive  = inclusive;
    e->exclusive  = exclusive;
    e->cpu_time   = cpu_time;
    e->mem_delta  = mem_delta;
    e->peak_memory = peak_memory;
    e->io_wait     = io_wait;
    e->network     = network;
}

void flush_buffer(void)
{
    uint32_t n = atomic_load(&NEBULA_G(write_pos));
    if (n == 0 || NEBULA_G(socket_fd) <= 0) {
        atomic_store(&NEBULA_G(write_pos), 0);
        return;
    }
    if (n > NEBULA_RING_SIZE) n = NEBULA_RING_SIZE;
    uint32_t sent = 0;
    while (sent < n) {
        uint32_t to_send = n - sent;
        if (to_send > NEBULA_BATCH_SIZE) to_send = NEBULA_BATCH_SIZE;
        size_t payload_size = to_send * sizeof(nebula_event_t);
        ssize_t res;
        do {
            res = sendto(NEBULA_G(socket_fd), &NEBULA_G(buffer)[sent], payload_size,
                         0, (struct sockaddr *)&NEBULA_G(agent_addr_un), sizeof(struct sockaddr_un));
        } while (res < 0 && errno == EINTR);

        if (res < 0) break;
        if ((size_t)res != payload_size) break;
        sent += to_send;
    }
    atomic_store(&NEBULA_G(write_pos), 0);
}

void nebula_send_session_end(unsigned char *session_id)
{
    emit_call(
        NEBULA_EVENT_SESSION_END, // event_type
        0,                        // func_id
        0,                        // inclusive
        0,                        // exclusive
        0,                        // cpu_time
        0,                        // mem_delta
        0,                        // peak_memory
        atomic_load(&NEBULA_G(overflow_count)), // io_wait = dropped events counter
        NEBULA_PROTOCOL_VERSION                 // network = protocol version
    );
}
