Zap Logger
===

Zap logger is the [Zap](https://github.com/uber-go/zap) logger for [Casbin](https://github.com/casbin/casbin). With this library, Casbin can log information more powerful.

## Installation

    go get github.com/casbin/zap-logger

## How to use it

You could let your enforcer use this logger when you first initialize your enforcer like:
```go
e, _ := casbin.NewEnforcer("examples/rbac_model.conf", a, zaplogger.NewLogger(true, true), true)
```
or with and existing zap instance.
```go
logger := zaplogger.NewLoggerByZap(yourZapLogger, true)
e, _ := casbin.NewEnforcer("examples/rbac_model.conf", a, logger, true)
```

And the method `NewLogger` have two params: enabled and jsonEncode, you could initialize logger's log status and decide whether to output information with json encoded or not.

## Getting Help

- [Casbin](https://github.com/casbin/casbin)

## License

This project is under Apache 2.0 License. See the [LICENSE](LICENSE) file for the full license text.
