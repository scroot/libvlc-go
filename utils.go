package vlc

// #cgo CFLAGS: -I.
// #cgo LDFLAGS: -L.
// #cgo linux LDFLAGS: -lvlc
// #cgo windows LDFLAGS: -lvlc.x64.dll
// #include <vlc/vlc.h>
// #include <stdlib.h>
import "C"
import "errors"

func getError() error {
	msg := C.libvlc_errmsg()
	if msg != nil {
		return errors.New(C.GoString(msg))
	}

	return nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}

	return 0
}
