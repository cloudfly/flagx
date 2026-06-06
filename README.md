# Flagx

Golang command line tool with environment supproted, Command line arguments have higher priority than environment variables.


## Examples

```go
logFile    = flagx.NewString("log.file", "", "the log file")
logLevel   = flagx.NewBool("log.level", false, "the log level, default is bool")
serverHost = flagx.NewString("server.addr", ":7070", "the address http will serve on")
serverPort = flagx.NewInt("server.port", 8080, "the tcp port will listen on")
secs = flagx.NewDuration("secs", "30s", "the duration flag, 30 secodns")
secs2 = flagx.NewDuration("days", "30d", "the duration flag, 30 days")
secs3 = flagx.NewDuration("weeks", "2w", "the duration flag, 2 weeks")
bytes = flagx.NewBytes("bytes", 128, "the bytes flag, 128 Byte")
kbs = flagx.NewBytes("kbytes", 64, "the bytes flag, 64KB")
mbs = flagx.NewBytes("mbytes", 64, "the bytes flag, 64KB")
gbs = flagx.NewBytes("gbytes", 64, "the bytes flag, 64KB")
strs = flagx.NewArrayString("array.str", "the string array flag, default is empty")
ints = flagx.NewArrayInt("array.int", "the string array flag, default is empty")
```

## Options

### Env

``` go
// the default env will be LOG_FILE
logFile := flagx.NewString("log.file", "", "the name flag")

// run command with --env.prefix SERVER_, the env name will be SERVER_LOG_FILE
logFile := flagx.NewString("log.file", "", "the name flag")

// custom the env name, the env name will be CUSTOM_LOG_FILE
logFile := flagx.NewString("log.file", "", "the name flag", flagx.Env("CUSTOM_LOG_FILE"))

// custom multiple env names, flagx will try them in from left to right order
logFile := flagx.NewString("log.file", "", "the name flag", flagx.Env("CUSTOM_LOG_FILE"), flagx.Env("SERVER_LOG_PATH"), , flagx.Env("LOG_PATH"))
```

### Follow
```go
// define a parent flag
addr := flagx.NewString("addr", "", "the address flag")

// let the addr2 use the addr's value if it not set.
//
// if --addr2 is set or ADDR2 environment variable is set, Follow will no effect.
addr2 := flagx.NewString("addr2", "", "the address flag", flagx.Follow(addr))
```