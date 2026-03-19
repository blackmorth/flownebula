#ifndef NEBULA_PROTOCOL_H
#define NEBULA_PROTOCOL_H

#include <stdint.h>

#define SESSION_ID_SIZE 8
#define NEBULA_EVENT_NAME 255

#pragma pack(push, 1)

// event_type
// 0 = ENTER
// 1 = EXIT
// 255 = NAME
typedef struct {
    char     session_id[SESSION_ID_SIZE];
    uint8_t  event_type;   // 0 / 1
    uint32_t func_id;
    uint64_t inclusive;    // EXIT only
    uint64_t exclusive;    // EXIT only
    uint64_t cpu_time;     // EXIT only (exclusive)
    int64_t  mem_delta;    // EXIT only
    uint64_t peak_memory;  // EXIT only
    uint64_t io_wait;      // 0 (réservé)
    uint64_t network;      // 0 (réservé)
} nebula_event_t;

/*
 * Message envoyé pour enregistrer le nom d'une fonction.
 * Identifié par event_type == NEBULA_EVENT_NAME (255).
 */
typedef struct {
    char     session_id[SESSION_ID_SIZE];
    uint8_t  event_type;   /* Doit être NEBULA_EVENT_NAME (255) */
    uint32_t func_id;
    uint32_t name_len;     /* Longueur du nom qui suit */
    char     name[256];    /* Nom de la fonction (tronqué à 255 + \0 si nécessaire côté agent) */
} nebula_name_t;

#pragma pack(pop)

#endif /* NEBULA_PROTOCOL_H */
