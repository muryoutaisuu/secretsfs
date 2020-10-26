# Enabling Debug Output

To enable debug output, just set `general.logging.level` to debug.
_secretsfs_ knows to different debug loggers:

* secretsfs: this logs the behaviour and values known to secretsfs
* fuse: this logs the events done by the [used fuse library by hanwen](https://github.com/hanwen/go-fuse#appendix-i-go-fuse-log-format).

The fuse library logging is only enabled on logging levels `trace` and `debug`.

```yaml
---
# General
general:
  # logging levels may be: {trace,debug,info,warn,error,fatal,panic}
  logging:
    level: debug
```
