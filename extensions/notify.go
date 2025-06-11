package extensions

import (
	"fmt"
	"os"
	"unsafe"

	_ "github.com/canonical/go-dqlite/v3"
	"github.com/mattn/go-sqlite3"
)

// #cgo LDFLAGS: -lsqlite3
// #include <stdlib.h>
// #include <sqlite3.h>
//
// typedef void (*sqlite3_func)(sqlite3_context*,int,sqlite3_value**);
//
// typedef void (*sqlite3_extension_init)(void);
//
// int notify_extension_trampoline(sqlite3 *connection, char **pzErrMsg, struct sqlite3_api_routines *pThunk);
// void notify_trampoline(sqlite3_context *context, int nargs, sqlite3_value **values);
// void watch_trampoline(sqlite3_context *context, int nargs, sqlite3_value **values);
// void unwatch_trampoline(sqlite3_context *context, int nargs, sqlite3_value **values);
import "C"

func RegisterNotifyExtension() error {
	err := C.sqlite3_auto_extension(C.sqlite3_extension_init(C.notify_extension_trampoline))
	if err != C.SQLITE_OK {
		return sqlite3.ErrNo(err)
	}

	return nil
}

//export notify_extension
func notify_extension(connection *C.sqlite3, pzErrMsg **C.char, pThunk *C.sqlite3_api_routines) C.int {
	err := create_function(connection, "notify", -1, C.sqlite3_func(C.notify_trampoline))
	if err != C.SQLITE_OK {
		return err
	}

	err = create_function(connection, "watch", 1, C.sqlite3_func(C.watch_trampoline))
	if err != C.SQLITE_OK {
		return err
	}

	err = create_function(connection, "unwatch", 1, C.sqlite3_func(C.unwatch_trampoline))
	if err != C.SQLITE_OK {
		return err
	}
	return 0
}

func create_function(connection *C.sqlite3, name string, nargs C.int, f C.sqlite3_func) C.int {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	return C.sqlite3_create_function_v2(
		connection,
		cname,
		nargs,
		C.SQLITE_UTF8|C.SQLITE_DIRECTONLY,
		nil,
		f, nil, nil, nil,
	)
}

var channels = make(map[string]chan any)

func setErr(context *C.sqlite3_context, err string) {
	str := C.CString(err)
	defer C.free(unsafe.Pointer(str))

	C.sqlite3_result_error(context, str, -1)
}

//export notify
func notify(context *C.sqlite3_context, nargs C.int, values **C.sqlite3_value) {
	args := unsafe.Slice(values, int(nargs))
	if len(args) < 1 {
		setErr(context, "expected at least one arguments")
		return
	} else if len(args) > 2 {
		setErr(context, "too many arguments")
		return
	}

	cname := (*C.char)(unsafe.Pointer(C.sqlite3_value_text(args[0])))
	name := C.GoString(cname)
	ch, ok := channels[name]
	if !ok {
		// Not an error
		C.sqlite3_result_int(context, 0)
		return
	}

	var value any
	if len(args) == 2 {
		switch C.sqlite3_value_type(args[1]) {
		case C.SQLITE_INTEGER:
			value = int64(C.sqlite3_value_int64(args[1]))
		case C.SQLITE_FLOAT:
			value = float64(C.sqlite3_value_double(args[1]))
		case C.SQLITE_TEXT:
			value = C.GoString((*C.char)(unsafe.Pointer(C.sqlite3_value_text(args[1]))))
		case C.SQLITE_BLOB:
			cblob := C.sqlite3_value_blob(args[1])
			cblobSize := C.sqlite3_value_bytes(args[1])
			blob := make([]byte, int(cblobSize))
			copy(blob, unsafe.Slice((*byte)(cblob), cblobSize))
			value = blob
		}
	}

	ch <- value
	C.sqlite3_result_int(context, 1)
}

//export watch
func watch(context *C.sqlite3_context, nargs C.int, values **C.sqlite3_value) {
	if nargs != 1 {
		panic("impossible")
	}

	args := unsafe.Slice(values, int(nargs))
	cname := (*C.char)(unsafe.Pointer(C.sqlite3_value_text(args[0])))
	name := C.GoString(cname)

	if _, ok := channels[name]; ok {
		setErr(context, fmt.Sprintf("channer %s already watched", name))
		return
	}

	ch := make(chan any)
	channels[name] = ch

	go func() {
		fmt.Fprintf(os.Stderr, "started watchng %s\n", name)
		for n := range ch {
			fmt.Fprintf(os.Stderr, "[%s]: %v\n", name, n)
		}
		fmt.Fprintf(os.Stderr, "stopped watchng %s\n", name)
	}()

	C.sqlite3_result_text(context, cname, -1, nil)
}

//export unwatch
func unwatch(context *C.sqlite3_context, nargs C.int, values **C.sqlite3_value) {
	if nargs != 1 {
		panic("impossible")
	}

	args := unsafe.Slice(values, int(nargs))
	cname := (*C.char)(unsafe.Pointer(C.sqlite3_value_text(args[0])))
	name := C.GoString(cname)

	ch, ok := channels[name]
	if !ok {
		setErr(context, fmt.Sprintf("channel %s is not being watched", name))
		return
	}

	close(ch)
	delete(channels, name)

	C.sqlite3_result_text(context, cname, -1, nil)
}
