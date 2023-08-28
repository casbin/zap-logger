
<h1 align="center" style="border-bottom: none;">Zap Logger</h1>
<h3 align="center">Zap logger is the Zap logger for Casbin. With this library, Casbin can log information more powerful.</h3>
<div class="labels">
  <p  align="center">
    <a href="https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg">
      <img src="https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg" alt="semantic-release">
    </a>
    <a href="https://goreportcard.com/report/github.com/casbin/zap-logger">
      <img src="https://goreportcard.com/badge/github.com/casbin/zap-logger" alt="Go Report Card">
    </a>
    <a href="https://coveralls.io/github/casbin/zap-logger?branch=master">
      <img src="https://coveralls.io/repos/github/casbin/zap-logger/badge.svg?branch=master" alt="Coverage Status">
    </a>
    <a href="https://pkg.go.dev/github.com/casbin/zap-logger/v2">
      <img src="https://godoc.org/github.com/casbin/zap-logger?status.svg" alt="Godoc">
    </a>
    <a href="https://github.com/casbin/zap-logger/releases/latest">
      <img src="https://img.shields.io/github/release/casbin/zap-logger.svg" alt="Release">
    </a>
  </p>
    
  <p  align="center">
    <a href="https://github.com/casdoor/zap-logger/blob/master/LICENSE">
      <img src="https://img.shields.io/github/license/casbin/zap-logger?style=flat-square" alt="license">
    </a>
    <a href="https://github.com/casdoor/zap-logger/issues">
      <img src="https://img.shields.io/github/issues/casbin/zap-logger?style=flat-square" alt="GitHub issues">
    </a>
    <a href="#">
      <img src="https://img.shields.io/github/stars/casbin/zap-logger?style=flat-square" alt="GitHub stars">
    </a>
    <a href="https://github.com/casbin/zap-logger/network">
      <img src="https://img.shields.io/github/forks/casbin/zap-logger?style=flat-square" alt="GitHub forks">
    </a>
    <a href="https://discord.gg/5rPsrAzK7S">
      <img src="https://img.shields.io/discord/1022748306096537660?style=flat-square&logo=discord&label=discord&color=5865F2" alt="Discord">
    </a>
  </p>
</div>

## Installation

    go get github.com/casbin/zap-logger

## How to use it

You could import the zap-logger module like:
```
import (
    zaplogger "github.com/casbin/zap-logger/v2"
    "github.com/casbin/casbin/v2"
)
```
You could let your enforcer use this logger when you first initialize your enforcer like:
```go
e, _ := casbin.NewEnforcer("examples/rbac_model.conf", a)
e.EnableLog(true)
e.SetLogger(zaplogger.NewLogger(true, true))
```
or with and existing zap instance.
```go
logger := zaplogger.NewLoggerByZap(yourZapLogger, true)
e, _ := casbin.NewEnforcer("examples/rbac_model.conf", a)
e.EnableLog(true)
e.SetLogger(logger)
```

And the method `NewLogger` have two params: enabled and jsonEncode, you could initialize logger's log status and decide whether to output information with json encoded or not.

## Getting Help

- [Casbin](https://github.com/casbin/casbin)

## License

This project is under Apache 2.0 License. See the [LICENSE](LICENSE) file for the full license text.
