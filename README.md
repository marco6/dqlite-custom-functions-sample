This repository serves as an example of how to use SQLite3 extensions in dqlite. The main trick is to register some [statically linked extensions](https://www.sqlite.org/c3ref/auto_extension.html) so that any connection used internally by dqlite have access to those functions/vtables/whatever.

The example is split int two types of extension:
 - a Greatest Common Denominator one, entirely written in C (enable it with `--gcd`)
 - a notification API written in Go (enable it with `--notify`).

The gcd one only adds one function, `gcd` accepting any number of integers and returning the gcd.

The notify one adds 3 functions:
 - `notify(<channel>[, <value>])` which notifies a named channel with an optional `<value>`
 - `watch(<channel>)` which starts to watch a notification channel. Notificaitons are written to stderr.
 - `unwatch(<channel>)` which stops watching a channel.

Please note that this code is a proof of concept and is lacking in many ways. In particular, the notify API is not transactional, so a transaction like:

```SQL
BEGIN;
SELECT notify("my-channel", 1);
ROLLBACK;
```

Will still notify `my-channel` with value `1`. This is true even when using triggers.

Most of the code, except the extension one, has been copied and adapted/simplified from the [go-dqlite](https://github.com/canonical/dqlite) source code.
