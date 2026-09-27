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
  em.deviceid = XIAllMasterDevices;
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

      memset(&st, 0, sizeof(st));

      if (XkbGetState(g_dpy, XkbUseCoreKbd, &st) != Success) {
        st.group = 0;
      }

      KeySym ks = XkbKeycodeToKeysym(g_dpy, kc, 0, 0);

      if (ks == NoSymbol) {
        ks = XkbKeycodeToKeysym(g_dpy, kc, 0, 1);
      }

      out->type = (cookie->evtype == XI_RawKeyPress) ? 1 : 2;

      out->keycode = (uint16_t)kc;
      out->dom_keycode = (uint16_t)keysym_to_dom_keycode(ks);

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

int keysym_to_dom_keycode(uint64_t keysym) {
  switch ((KeySym)keysym) {
  /* Letters */
  case XK_a:
  case XK_A:
    return 65;
  case XK_b:
  case XK_B:
    return 66;
  case XK_c:
  case XK_C:
    return 67;
  case XK_d:
  case XK_D:
    return 68;
  case XK_e:
  case XK_E:
    return 69;
  case XK_f:
  case XK_F:
    return 70;
  case XK_g:
  case XK_G:
    return 71;
  case XK_h:
  case XK_H:
    return 72;
  case XK_i:
  case XK_I:
    return 73;
  case XK_j:
  case XK_J:
    return 74;
  case XK_k:
  case XK_K:
    return 75;
  case XK_l:
  case XK_L:
    return 76;
  case XK_m:
  case XK_M:
    return 77;
  case XK_n:
  case XK_N:
    return 78;
  case XK_o:
  case XK_O:
    return 79;
  case XK_p:
  case XK_P:
    return 80;
  case XK_q:
  case XK_Q:
    return 81;
  case XK_r:
  case XK_R:
    return 82;
  case XK_s:
  case XK_S:
    return 83;
  case XK_t:
  case XK_T:
    return 84;
  case XK_u:
  case XK_U:
    return 85;
  case XK_v:
  case XK_V:
    return 86;
  case XK_w:
  case XK_W:
    return 87;
  case XK_x:
  case XK_X:
    return 88;
  case XK_y:
  case XK_Y:
    return 89;
  case XK_z:
  case XK_Z:
    return 90;

  /* Number row */
  case XK_0:
    return 48;
  case XK_1:
    return 49;
  case XK_2:
    return 50;
  case XK_3:
    return 51;
  case XK_4:
    return 52;
  case XK_5:
    return 53;
  case XK_6:
    return 54;
  case XK_7:
    return 55;
  case XK_8:
    return 56;
  case XK_9:
    return 57;

  /* Function keys */
  case XK_F1:
    return 112;
  case XK_F2:
    return 113;
  case XK_F3:
    return 114;
  case XK_F4:
    return 115;
  case XK_F5:
    return 116;
  case XK_F6:
    return 117;
  case XK_F7:
    return 118;
  case XK_F8:
    return 119;
  case XK_F9:
    return 120;
  case XK_F10:
    return 121;
  case XK_F11:
    return 122;
  case XK_F12:
    return 123;

  /* Navigation */
  case XK_Home:
    return 36;
  case XK_End:
    return 35;
  case XK_Left:
    return 37;
  case XK_Up:
    return 38;
  case XK_Right:
    return 39;
  case XK_Down:
    return 40;
  case XK_Prior:
    return 33;
  case XK_Next:
    return 34;
  case XK_Insert:
    return 45;
  case XK_Delete:
    return 46;

  /* Keypad navigation */
  case XK_KP_Home:
    return 36;
  case XK_KP_End:
    return 35;
  case XK_KP_Left:
    return 37;
  case XK_KP_Up:
    return 38;
  case XK_KP_Right:
    return 39;
  case XK_KP_Down:
    return 40;
  case XK_KP_Prior:
    return 33;
  case XK_KP_Next:
    return 34;
  case XK_KP_Insert:
    return 45;
  case XK_KP_Delete:
    return 46;

  /* Editing/control */
  case XK_BackSpace:
    return 8;
  case XK_Tab:
    return 9;
  case XK_Return:
    return 13;
  case XK_Escape:
    return 27;
  case XK_space:
    return 32;

  default:
    return 0;
  }
}
