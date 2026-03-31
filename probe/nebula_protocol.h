// nebula_protocol.h

#ifndef NEBULA_PROTOCOL_H
#define NEBULA_PROTOCOL_H

#include <stdint.h>   // ← indispensable
#include <stddef.h>   // ← pour size_t

#include "php.h"      // ← indispensable AVANT toute macro PHP
#include "zend.h"
#include "zend_API.h"
#include "zend_exceptions.h"
#include "zend_extensions.h"
#include "zend_compile.h"
#include "zend_execute.h"
#include "zend_interfaces.h"
#include "zend_types.h"


#define SESSION_ID_SIZE 8
#define NEBULA_EVENT_NAME 255
#define NEBULA_EVENT_SESSION_END 0xFE

#pragma pack(push, 1)

// 0 = ENTER, 1 = EXIT, 255 = NAME, 0xFE = SESSION_END
typedef struct {
    char     session_id[SESSION_ID_SIZE];
    uint8_t  kind;          // 0 = ENTER, 1 = EXIT

    uint64_t ts_ns;         // horloge monotone (zend_hrtime_nebula)
    uint32_t depth;         // profondeur d’appel

    uint32_t func_id;       // identifiant stable (catalogue côté agent)

    // Métadonnées gratuites
    uint8_t  func_type;     // func->type
    uint8_t  flags;         // bits: static, closure, generator, etc.
    uint8_t  arg_count;     // ZEND_CALL_NUM_ARGS(execute_data)
    uint8_t  has_exception; // EG(exception) != NULL
    uint8_t  jit_flag;      // JIT actif sur cette frame ?

    uint64_t unix_time_ns;  // timestamp absolu (corrélation)
} nebula_event_t;

/*
 * Message envoyé pour enregistrer le nom d'une fonction.
 * Identifié par kind == NEBULA_EVENT_NAME (255).
 */
typedef struct {
    char     session_id[SESSION_ID_SIZE];
    uint8_t  kind;       /* NEBULA_EVENT_NAME (255) */
    uint32_t func_id;
    uint32_t name_len;
    char     name[256];
} nebula_name_t;

#pragma pack(pop)

#endif /* NEBULA_PROTOCOL_H */
