package extensions

import (
	_ "unsafe"

	"github.com/mattn/go-sqlite3"
)

// #include "native.h"
// #include <sqlite3.h>
import "C"

// RegisterNativeExtension registers two native (C language)
// extensions to dqlite.
func RegisterNativeExtension() error {
	err := C.register_native_extensions()
	if err != C.SQLITE_OK {
		return sqlite3.ErrNo(err)
	}
	return nil
}
