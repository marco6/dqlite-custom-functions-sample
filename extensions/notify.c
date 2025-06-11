
#include <sqlite3.h>

extern int notify_extension(sqlite3 *connection, char **pzErrMsg,
                            struct sqlite3_api_routines *pThunk);
int notify_extension_trampoline(sqlite3 *connection, char **pzErrMsg,
                                struct sqlite3_api_routines *pThunk) {
  return notify_extension(connection, pzErrMsg, pThunk);
}

extern void notify(sqlite3_context *context, int nargs, sqlite3_value **values);
void notify_trampoline(sqlite3_context *context, int nargs,
                       sqlite3_value **values) {
  return notify(context, nargs, values);
}

extern void watch(sqlite3_context *context, int nargs, sqlite3_value **values);
void watch_trampoline(sqlite3_context *context, int nargs,
                      sqlite3_value **values) {
  return watch(context, nargs, values);
}

extern void unwatch(sqlite3_context *context, int nargs,
                    sqlite3_value **values);
void unwatch_trampoline(sqlite3_context *context, int nargs,
                        sqlite3_value **values) {
  return unwatch(context, nargs, values);
}