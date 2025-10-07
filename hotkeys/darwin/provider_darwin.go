//go:build darwin

package darwin

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework ApplicationServices
void hk_start(void);
void hk_stop(void);
int  hk_wait_next(unsigned short* out_keycode,
                  char*           out_name,
                  unsigned long   out_name_cap);
*/
import "C"
import (
	"sync"
	"unsafe"

	"github.com/zellydev-games/opensplit/keyinfo"
)

type Event struct {
	KeyCode uint16
	Name    string
}

type Manager struct {
	callback func(data keyinfo.KeyData)
	mu       sync.Mutex
}

func (m *Manager) StartHook(callback func(data keyinfo.KeyData)) error {
	m.mu.Lock()
	m.callback = callback
	m.mu.Unlock()

	C.hk_start()
	go func() {
		for {
			var kc C.ushort
			buf := make([]byte, 32)
			ok := C.hk_wait_next(&kc, (*C.char)(unsafe.Pointer(&buf[0])), C.ulong(len(buf))) != 0
			if !ok {
				return
			}
			name := C.GoString((*C.char)(unsafe.Pointer(&buf[0])))
			m.mu.Lock()
			cb := m.callback
			m.mu.Unlock()
			if cb != nil {
				cb(keyinfo.KeyData{
					KeyCode:    int(kc),
					LocaleName: name,
				})
			}
		}
	}()
	return nil
}

func (m *Manager) Unhook() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callback = nil
	C.hk_stop()
	return nil
}
