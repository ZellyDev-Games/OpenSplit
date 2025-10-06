#pragma once
#include <stdint.h>
typedef struct {
    uint8_t  type;     // 1=press, 2=release
    uint16_t keycode;  // raw X keycode
    char     name[64]; // keysym name (UTF-8) or "(unknown)"
} xi2_event;

// Returns 0 on success, nonzero on failure (e.g., Wayland / XI2 missing).
int xi2_open(char *err, int err_len);

// Blocks until a RawKeyPress/RawKeyRelease and fills 'out'. Returns 0 on success.
// Returns nonzero on fatal error (display closed).
int xi2_next(xi2_event *out);

// Close display and free resources.
void xi2_close(void);
