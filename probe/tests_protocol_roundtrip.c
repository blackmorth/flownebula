#include <assert.h>
#include <stdint.h>
#include <string.h>

#include "nebula_protocol.h"

int main(void) {
    nebula_event_t ev = {0};
    uint8_t raw[sizeof(nebula_event_t)];

    memcpy(ev.session_id, "ABCDEFGH", SESSION_ID_SIZE);
    ev.event_type = 1;
    ev.func_id = 42;
    ev.inclusive = 1234;
    ev.exclusive = 1200;
    ev.cpu_time = 1000;
    ev.mem_delta = -64;
    ev.peak_memory = 2048;
    ev.io_wait = 3;
    ev.network = 4;
    ev.event_time_unix_ns = 555;
    ev.alloc_bytes = 1024;
    ev.free_bytes = 512;

    memcpy(raw, &ev, sizeof(ev));

    nebula_event_t decoded = {0};
    memcpy(&decoded, raw, sizeof(decoded));

    assert(memcmp(decoded.session_id, "ABCDEFGH", SESSION_ID_SIZE) == 0);
    assert(decoded.event_type == 1);
    assert(decoded.func_id == 42);
    assert(decoded.inclusive == 1234);
    assert(decoded.exclusive == 1200);
    assert(decoded.cpu_time == 1000);
    assert(decoded.mem_delta == -64);
    assert(decoded.peak_memory == 2048);
    assert(decoded.io_wait == 3);
    assert(decoded.network == 4);
    assert(decoded.event_time_unix_ns == 555);
    assert(decoded.alloc_bytes == 1024);
    assert(decoded.free_bytes == 512);

    nebula_name_t nameMsg = {0};
    memcpy(nameMsg.session_id, "12345678", SESSION_ID_SIZE);
    nameMsg.event_type = NEBULA_EVENT_NAME;
    nameMsg.func_id = 9;
    nameMsg.name_len = 4;
    memcpy(nameMsg.name, "main", 4);

    assert(nameMsg.event_type == 255);
    assert(nameMsg.name_len == 4);
    assert(memcmp(nameMsg.name, "main", 4) == 0);

    return 0;
}
