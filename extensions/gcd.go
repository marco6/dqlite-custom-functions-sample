package extensions

import (
	_ "unsafe"

	"github.com/mattn/go-sqlite3"
)

// #include <sqlite3.h>
//
// int register_native_extensions();
// int unregister_native_extensions();
import "C"

// RegisterGcdExtension registers two native (C language)
// extensions to dqlite.
func RegisterGcdExtension() error {
	err := C.register_native_extensions()
	if err != C.SQLITE_OK {
		return sqlite3.ErrNo(err)
	}
	return nil
}
