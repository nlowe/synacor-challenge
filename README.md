# synacor-challenge

[![](https://github.com/nlowe/synacor-challenge/workflows/CI/badge.svg)](https://github.com/nlowe/synacor-challenge/actions) [![Coverage Status](https://coveralls.io/repos/github/nlowe/synacor-challenge/badge.svg?branch=master)](https://coveralls.io/github/nlowe/synacor-challenge?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/nlowe/synacor-challenge)](https://goreportcard.com/report/github.com/nlowe/synacor-challenge) [![License](https://img.shields.io/badge/license-MIT-brightgreen)](./LICENSE)

A Go implementation of the [Synacor challenge](https://challenge.synacor.com/)

Currently solves 3/8 flags:

* Sign-up (implied)
* Basic I/O
* Full POST

## Building

You need Go 1.18+. The VM can be run with `go run`:

```
Go implementation of the synacor challenge VM spec

Usage:
  synacor [flags]

Flags:
  -c, --challenge-file string   Challenge File to execute (default "challenge.bin")
  -d, --debug-log string        Record instructions to debug log
  -h, --help                    help for synacor
      --io-watchdog duration    Timeout for I/O operations (default 5m0s)
```
