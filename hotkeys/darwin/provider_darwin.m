#import <ApplicationServices/ApplicationServices.h>  // Quartz + CGEvent APIs
#include <pthread.h>                                  // Threads, mutex, condvar
#include <stdbool.h>                                  // bool type
#include <stdint.h>                                   // fixed-width ints
#include <string.h>                                   // memset, strncpy, etc.
#include <stdio.h>                                    // snprintf

void hk_start(void);
void hk_stop(void);
int  hk_wait_next(uint16_t* out_keycode, char* out_name, size_t out_name_cap);

static CFMachPortRef      gEventTap     = NULL;  // The tap itself
static CFRunLoopSourceRef gRunLoopSrc   = NULL;  // RunLoop source wrapping the tap
static CFRunLoopRef       gRunLoop      = NULL;  // The runloop that owns the tap
static pthread_t          gThread       = 0;     // Background thread id

static pthread_mutex_t gMu = PTHREAD_MUTEX_INITIALIZER;
static pthread_cond_t  gCv = PTHREAD_COND_INITIALIZER;

static bool         gRunning = false;   // tap/thread alive
static bool         gHasMsg  = false;   // mailbox has an unread event
static uint16_t     gKeycode = 0;       // mailbox fields
static char         gName[32];

static void CGEventKeyDisplayName(CGEventRef event, char *out, size_t cap) {
    if (!out || cap == 0 || !event) return;
    out[0] = '\0';

    CGEventRef ev = event;
    CGEventRef temp = NULL;
    temp = CGEventCreateCopy(event);
    if (temp) {
        CGEventFlags f = CGEventGetFlags(temp);
        f &= ~(kCGEventFlagMaskShift | kCGEventFlagMaskAlternate |
               kCGEventFlagMaskCommand | kCGEventFlagMaskControl);
        CGEventSetFlags(temp, f);
        ev = temp;
    }

    // Ask CoreGraphics for the Unicode string this key would generate.
    UniChar buf[8] = {0};
    UniCharCount outLen = 0;
    CGEventKeyboardGetUnicodeString(ev, 8, &outLen, buf);

    if (temp) CFRelease(temp);

    if (outLen > 0) {
        // Convert UTF-16 UniChar[] to UTF-8.
        CFStringRef s = CFStringCreateWithCharactersNoCopy(kCFAllocatorDefault, buf, outLen, kCFAllocatorNull);
        if (s) {
            if (CFStringGetCString(s, out, (CFIndex)cap, kCFStringEncodingUTF8)) {
                // If it's a single ASCII letter, upper-case for a keycap look.
                if (out[1] == '\0' && out[0] >= 'a' && out[0] <= 'z') {
                    out[0] = (char)(out[0]-'a'+'A');
                }
                // Show a visible name for space
                if (out[0] == ' ' && out[1] == '\0') {
                    snprintf(out, cap, "Space");
                }
                CFRelease(s);
                return;
            }
            CFRelease(s);
        }
    }

    // Non-printing fallback: use the hardware keycode for a stable name.
    // (These codes come from kCGKeyboardEventKeycode.)
    uint16_t kc = (uint16_t)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
    switch (kc) {
        case 36:  snprintf(out, cap, "Return");  return;
        case 48:  snprintf(out, cap, "Tab");     return;
        case 49:  snprintf(out, cap, "Space");   return;
        case 51:  snprintf(out, cap, "Delete");  return; // Backspace
        case 53:  snprintf(out, cap, "Escape");  return;

        case 123: snprintf(out, cap, "Left");    return;
        case 124: snprintf(out, cap, "Right");   return;
        case 125: snprintf(out, cap, "Down");    return;
        case 126: snprintf(out, cap, "Up");      return;

        case 122: snprintf(out, cap, "F1");      return;
        case 120: snprintf(out, cap, "F2");      return;
        case 99:  snprintf(out, cap, "F3");      return;
        case 118: snprintf(out, cap, "F4");      return;
        case 96:  snprintf(out, cap, "F5");      return;
        case 97:  snprintf(out, cap, "F6");      return;
        case 98:  snprintf(out, cap, "F7");      return;
        case 100: snprintf(out, cap, "F8");      return;
        case 101: snprintf(out, cap, "F9");      return;
        case 109: snprintf(out, cap, "F10");     return;
        case 103: snprintf(out, cap, "F11");     return;
        case 111: snprintf(out, cap, "F12");     return;
    }

    // Last resort
    snprintf(out, cap, "Keycode:%u", (unsigned)kc);
}

// Event tap callback: write the *latest* event into the mailbox and signal.
static CGEventRef tapCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    if (type == kCGEventTapDisabledByTimeout || type == kCGEventTapDisabledByUserInput) {
        if (gEventTap) CGEventTapEnable(gEventTap, true);
        return event;
    }
    if (type != kCGEventKeyDown) return event;

    int64_t isRepeat = CGEventGetIntegerValueField(event, kCGKeyboardEventAutorepeat);
    if (isRepeat) return event;

    char name[32];
    CGEventKeyDisplayName(event, name, sizeof(name));
    uint16_t kc = (uint16_t)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);

    // Store into mailbox (overwrites any previous unread event).
    pthread_mutex_lock(&gMu);
    gKeycode = kc;
    strncpy(gName, name, sizeof(gName)-1);
    gName[sizeof(gName)-1] = '\0';
    gHasMsg = true;
    pthread_cond_signal(&gCv); // wake one waiter
    pthread_mutex_unlock(&gMu);

    return event;
}

static void* tapThread(void* _) {
    pthread_mutex_lock(&gMu);
    gRunning = true;
    pthread_mutex_unlock(&gMu);

    CGEventMask mask = (CGEventMaskBit(kCGEventKeyDown));
    gEventTap = CGEventTapCreate(kCGSessionEventTap, kCGHeadInsertEventTap,
                                 kCGEventTapOptionListenOnly, mask, tapCallback, NULL);
    if (!gEventTap) {
        pthread_mutex_lock(&gMu);
        gRunning = false;
        pthread_cond_broadcast(&gCv);
        pthread_mutex_unlock(&gMu);
        return NULL;
    }
    gRunLoopSrc = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, gEventTap, 0);
    gRunLoop = CFRunLoopGetCurrent();
    CFRunLoopAddSource(gRunLoop, gRunLoopSrc, kCFRunLoopCommonModes);
    CGEventTapEnable(gEventTap, true);
    CFRunLoopRun();

    if (gRunLoopSrc) { CFRelease(gRunLoopSrc); gRunLoopSrc = NULL; }
    if (gEventTap)   { CFRelease(gEventTap);   gEventTap   = NULL; }
    gRunLoop = NULL;

    pthread_mutex_lock(&gMu);
    gRunning = false;
    pthread_cond_broadcast(&gCv);
    pthread_mutex_unlock(&gMu);
    return NULL;
}

void hk_start(void) {
    pthread_mutex_lock(&gMu);
    bool already = gRunning;
    pthread_mutex_unlock(&gMu);
    if (already) return;
    pthread_create(&gThread, NULL, tapThread, NULL);
}

void hk_stop(void) {
    if (gRunLoop) CFRunLoopStop(gRunLoop);
    if (gThread) { pthread_join(gThread, NULL); gThread = 0; }
}

// Blocks until we either (a) have a mailbox message, or (b) the tap stops.
int hk_wait_next(uint16_t* out_keycode, char* out_name, size_t out_cap) {
    pthread_mutex_lock(&gMu);
    for (;;) {
        if (gHasMsg) {
            if (out_keycode) *out_keycode = gKeycode;
            if (out_name && out_cap) {
                size_t n = strnlen(gName, sizeof(gName));
                if (n >= out_cap) n = out_cap - 1;
                memcpy(out_name, gName, n);
                out_name[n] = '\0';
            }
            gHasMsg = false; // consume
            pthread_mutex_unlock(&gMu);
            return 1;
        }
        if (!gRunning) {
            pthread_mutex_unlock(&gMu);
            return 0; // tap closed
        }
        pthread_cond_wait(&gCv, &gMu);
    }
}
