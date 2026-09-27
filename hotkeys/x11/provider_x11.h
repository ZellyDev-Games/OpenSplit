#pragma once

#include <stdint.h>

typedef struct {
  uint8_t type;       // 1=press, 2=release
  uint16_t keycode;   // raw X keycode
  uint32_t modifiers; // X11 effective modifier mask
  char name[64];      // keysym name or "(unknown)"
} xi2_event;

// Returns 0 on success, nonzero on failure.
int xi2_open(char *err, int err_len);

// Blocks until a RawKeyPress/RawKeyRelease or shutdown.
// Returns:
//   0 = event returned
//   1 = fatal X11 error
//   2 = shutdown requested
int xi2_next(xi2_event *out);

// Requests xi2_next() to wake up and return 2.
void xi2_stop(void);

// Closes the X11 display and frees resources.
// Must be called after xi2_next() has returned.
void xi2_close(void);
