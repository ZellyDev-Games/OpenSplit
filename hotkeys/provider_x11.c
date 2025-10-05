#include "provider_x11.h"
#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <X11/Xlib.h>
#include <X11/XKBlib.h>
#include <X11/keysym.h>
#include <X11/extensions/XInput2.h>

static Display *g_dpy = NULL;
static int g_xi_opcode = -1;
static unsigned char *g_mask = NULL;
static int g_kc_min = 8, g_kc_max = 255;

static int select_raw(Display *dpy, Window root) {
    int major = 2, minor = 0;
    if (XIQueryVersion(dpy, &major, &minor) != Success) return 1;

    int mlen = XIMaskLen(XI_LASTEVENT);
    g_mask = calloc(mlen, 1);
    if (!g_mask) return 2;

    XIEventMask em = {0};
    em.deviceid = XIAllDevices;
    em.mask_len = mlen;
    em.mask = g_mask;
    XISetMask(g_mask, XI_RawKeyPress);
    XISetMask(g_mask, XI_RawKeyRelease);

    if (XISelectEvents(dpy, root, &em, 1) != Success) return 3;
    XFlush(dpy);
    return 0;
}

int xi2_open(char *err, int err_len) {
    if (g_dpy) return 0;
    g_dpy = XOpenDisplay(NULL);
    if (!g_dpy) { if (err) snprintf(err, err_len, "XOpenDisplay failed"); return 1; }

    int ev, er;
    if (!XQueryExtension(g_dpy, "XInputExtension", &g_xi_opcode, &ev, &er)) {
        if (err) snprintf(err, err_len, "XInputExtension missing (Wayland?)");
        XCloseDisplay(g_dpy); g_dpy = NULL; return 2;
    }

    XDisplayKeycodes(g_dpy, &g_kc_min, &g_kc_max);
    if (select_raw(g_dpy, DefaultRootWindow(g_dpy)) != 0) {
        if (err) snprintf(err, err_len, "XI2 select failed");
        XCloseDisplay(g_dpy); g_dpy = NULL; return 3;
    }
    return 0;
}

int xi2_next(xi2_event *out) {
    if (!g_dpy) return 1;

    for (;;) {
        XEvent ev;
        XNextEvent(g_dpy, &ev); // blocks

        if (ev.type != GenericEvent) continue;
        XGenericEventCookie *cookie = &ev.xcookie;
        if (!XGetEventData(g_dpy, cookie)) continue;
        if (cookie->extension != g_xi_opcode) { XFreeEventData(g_dpy, cookie); continue; }

        if (cookie->evtype == XI_RawKeyPress || cookie->evtype == XI_RawKeyRelease) {
            XIRawEvent *raw = (XIRawEvent*)cookie->data;
            KeyCode kc = (KeyCode)raw->detail;

            if (kc < g_kc_min || kc > g_kc_max) {
                XFreeEventData(g_dpy, cookie); continue;
            }

            // simple group/level: group from XKB state; level 0/1 by Shift snapshot
            XkbStateRec st;
            if (XkbGetState(g_dpy, XkbUseCoreKbd, &st) != Success) st.group = 0;

            KeySym ks = XkbKeycodeToKeysym(g_dpy, kc, st.group, 0);
            if (ks == NoSymbol) ks = XkbKeycodeToKeysym(g_dpy, kc, st.group, 1);
            if (ks == NoSymbol) ks = XkbKeycodeToKeysym(g_dpy, kc, 0, 0);

            memset(out, 0, sizeof(*out));
            out->type = (cookie->evtype == XI_RawKeyPress) ? 1 : 2;
            out->keycode = (uint16_t)kc;

            const char *nm = (ks != NoSymbol) ? XKeysymToString(ks) : NULL;
            if (nm && *nm) {
                strncpy(out->name, nm, sizeof(out->name)-1);
            } else {
                strncpy(out->name, "(unknown)", sizeof(out->name)-1);
            }

            XFreeEventData(g_dpy, cookie);
            return 0;
        }

        XFreeEventData(g_dpy, cookie);
    }
}

void xi2_close(void) {
    if (g_mask) { free(g_mask); g_mask = NULL; }
    if (g_dpy)  { XCloseDisplay(g_dpy); g_dpy = NULL; }
}