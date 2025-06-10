#include "native.h"
#include <sqlite3.h>
#include <time.h>

static int dqlite_auto_extension(sqlite3 *connection, char **pzErrMsg,
                                 struct sqlite3_api_routines *pThunk);

int register_native_extensions() {
  return sqlite3_auto_extension((void (*)(void))dqlite_auto_extension);
}

int unregister_native_extensions() {
  return sqlite3_cancel_auto_extension((void (*)(void))dqlite_auto_extension);
}

static sqlite3_int64 gcd_impl(sqlite3_int64 a, sqlite3_int64 b) {
  if (b == 0)
    return a;

  return gcd_impl(b, a % b);
}

/* Greatest Common Denominator */
static void gcd(sqlite3_context *context, int argc, sqlite3_value **argv) {
  if (argc <= 0) {
    return sqlite3_result_error(context, "gcd requires at least one argument",
                                -1);
  }
  sqlite3_int64 result = sqlite3_value_int64(argv[0]);
  for (int i = 1; i < argc; i++) {
    result = gcd_impl(result, sqlite3_value_int64(argv[i]));
  }
  sqlite3_result_int64(context, result);
}

static int dqlite_auto_extension(sqlite3 *connection, char **pzErrMsg,
                                 struct sqlite3_api_routines *pThunk) {
  return sqlite3_create_function_v2(connection, "gcd", -1,
                                    SQLITE_UTF8 | SQLITE_DETERMINISTIC, NULL,
                                    gcd, NULL, NULL, NULL);
}
