#include <assert.h>
#include <stddef.h>
#include <stdint.h>

#include "nebula_protocol.h"

int main(void) {
    assert(SESSION_ID_SIZE == 8);
    assert(NEBULA_EVENT_NAME == 255);
    assert(NEBULA_EVENT_SESSION_END == 0xFE);

    assert(sizeof(nebula_event_t) == 69);
    assert(offsetof(nebula_event_t, event_type) == 8);
    assert(offsetof(nebula_event_t, func_id) == 9);
    assert(offsetof(nebula_event_t, network) == 61);

    assert(sizeof(nebula_name_t) == 273);
    assert(offsetof(nebula_name_t, name_len) == 13);
    assert(offsetof(nebula_name_t, name) == 17);

    return 0;
}
