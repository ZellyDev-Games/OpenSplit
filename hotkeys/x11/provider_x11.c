#include "provider_x11.h"

#include <errno.h>
#include <fcntl.h>
#include <poll.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#include <X11/XKBlib.h>
#include <X11/Xlib.h>
#include <X11/extensions/XInput2.h>
#include <X11/keysym.h>

static Display *g_dpy = NULL;
static int g_xi_opcode = -1;
static unsigned char *g_mask = NULL;
static int g_kc_min = 8;
static int g_kc_max = 255;

static int g_stop_pipe[2] = {-1, -1};

static int select_raw(Display *dpy, Window root) {
  int major = 2;
  int minor = 0;

  if (XIQueryVersion(dpy, &major, &minor) != Success) {
    return 1;
  }

  int mlen = XIMaskLen(XI_LASTEVENT);

  g_mask = calloc(mlen, 1);
  if (!g_mask) {
    return 2;
  }

  XIEventMask em = {0};
  em.deviceid = XIAllDevices;
  em.mask_len = mlen;
  em.mask = g_mask;

  XISetMask(g_mask, XI_RawKeyPress);
  XISetMask(g_mask, XI_RawKeyRelease);

  if (XISelectEvents(dpy, root, &em, 1) != Success) {
    return 3;
  }

  XFlush(dpy);

  return 0;
}

static int create_stop_pipe(void) {
  if (g_stop_pipe[0] >= 0 || g_stop_pipe[1] >= 0) {
    return 0;
  }

  if (pipe(g_stop_pipe) != 0) {
    return 1;
  }

  int flags = fcntl(g_stop_pipe[0], F_GETFL, 0);
  if (flags >= 0) {
    fcntl(g_stop_pipe[0], F_SETFL, flags | O_NONBLOCK);
  }

  flags = fcntl(g_stop_pipe[1], F_GETFL, 0);
  if (flags >= 0) {
    fcntl(g_stop_pipe[1], F_SETFL, flags | O_NONBLOCK);
  }

  return 0;
}

static void destroy_stop_pipe(void) {
  if (g_stop_pipe[0] >= 0) {
    close(g_stop_pipe[0]);
    g_stop_pipe[0] = -1;
  }

  if (g_stop_pipe[1] >= 0) {
    close(g_stop_pipe[1]);
    g_stop_pipe[1] = -1;
  }
}

int xi2_open(char *err, int err_len) {
  if (g_dpy) {
    return 0;
  }

  if (create_stop_pipe() != 0) {
    if (err) {
      snprintf(err, err_len, "failed to create X11 stop pipe");
    }
    return 1;
  }

  g_dpy = XOpenDisplay(NULL);

  if (!g_dpy) {
    if (err) {
      snprintf(err, err_len, "XOpenDisplay failed");
    }

    destroy_stop_pipe();
    return 2;
  }

  int ev;
  int er;

  if (!XQueryExtension(g_dpy, "XInputExtension", &g_xi_opcode, &ev, &er)) {
    if (err) {
      snprintf(err, err_len, "XInputExtension missing");
    }

    XCloseDisplay(g_dpy);
    g_dpy = NULL;
    destroy_stop_pipe();

    return 3;
  }

  XDisplayKeycodes(g_dpy, &g_kc_min, &g_kc_max);

  if (select_raw(g_dpy, DefaultRootWindow(g_dpy)) != 0) {
    if (err) {
      snprintf(err, err_len, "XI2 select failed");
    }

    XCloseDisplay(g_dpy);
    g_dpy = NULL;
    destroy_stop_pipe();

    return 4;
  }

  return 0;
}

int xi2_next(xi2_event *out) {
  if (!g_dpy) {
    return 1;
  }

  int xfd = ConnectionNumber(g_dpy);

  struct pollfd fds[2];

  fds[0].fd = xfd;
  fds[0].events = POLLIN;

  fds[1].fd = g_stop_pipe[0];
  fds[1].events = POLLIN;

  for (;;) {
    int rc = poll(fds, 2, -1);

    if (rc < 0) {
      if (errno == EINTR) {
        continue;
      }

      return 1;
    }

    if (fds[1].revents & POLLIN) {
      char buffer[32];

      while (read(g_stop_pipe[0], buffer, sizeof(buffer)) > 0) {
      }

      return 2;
    }

    if (!(fds[0].revents & POLLIN)) {
      continue;
    }

    while (XPending(g_dpy) > 0) {
      XEvent ev;

      XNextEvent(g_dpy, &ev);

      if (ev.type != GenericEvent) {
        continue;
      }

      XGenericEventCookie *cookie = &ev.xcookie;

      if (!XGetEventData(g_dpy, cookie)) {
        continue;
      }

      if (cookie->extension != g_xi_opcode) {
        XFreeEventData(g_dpy, cookie);
        continue;
      }

      if (cookie->evtype != XI_RawKeyPress &&
          cookie->evtype != XI_RawKeyRelease) {
        XFreeEventData(g_dpy, cookie);
        continue;
      }

      XIRawEvent *raw = (XIRawEvent *)cookie->data;
      KeyCode kc = (KeyCode)raw->detail;

      if (kc < g_kc_min || kc > g_kc_max) {
        XFreeEventData(g_dpy, cookie);
        continue;
      }

      XkbStateRec st;

      if (XkbGetState(g_dpy, XkbUseCoreKbd, &st) != Success) {
        st.group = 0;
      }

      KeySym ks = XkbKeycodeToKeysym(g_dpy, kc, st.group, 0);

      if (ks == NoSymbol) {
        ks = XkbKeycodeToKeysym(g_dpy, kc, st.group, 1);
      }

      if (ks == NoSymbol) {
        ks = XkbKeycodeToKeysym(g_dpy, kc, 0, 0);
      }

      memset(out, 0, sizeof(*out));

      out->type = (cookie->evtype == XI_RawKeyPress) ? 1 : 2;

      out->keycode = (uint16_t)kc;

      /*
       * X11 modifier masks:
       *
       * ShiftMask   = 1
       * LockMask    = 2
       * ControlMask = 4
       * Mod1Mask    = 8    usually Alt
       * Mod2Mask    = 16
       * Mod3Mask    = 32
       * Mod4Mask    = 64   usually Super/Meta
       * Mod5Mask    = 128
       *
       * Keep the complete effective mask here. Go performs
       * normalization when matching bindings.
       */
      out->modifiers = (uint32_t)st.mods;

      const char *nm = (ks != NoSymbol) ? XKeysymToString(ks) : NULL;

      if (nm && *nm) {
        strncpy(out->name, nm, sizeof(out->name) - 1);
      } else {
        strncpy(out->name, "(unknown)", sizeof(out->name) - 1);
      }

      XFreeEventData(g_dpy, cookie);

      return 0;
    }
  }
}

void xi2_stop(void) {
  if (g_stop_pipe[1] < 0) {
    return;
  }

  const char byte = 1;

  if (write(g_stop_pipe[1], &byte, 1) < 0) {
    // The reader may already have stopped.
  }
}

void xi2_close(void) {
  if (g_mask) {
    free(g_mask);
    g_mask = NULL;
  }

  if (g_dpy) {
    XCloseDisplay(g_dpy);
    g_dpy = NULL;
  }

  destroy_stop_pipe();
}
