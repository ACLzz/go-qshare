# go-qshare
[![codecov](https://codecov.io/github/ACLzz/go-qshare/graph/badge.svg?token=S39GK295NH)](https://codecov.io/github/ACLzz/go-qshare)<br/>
GO library to communicate with devices via Quick Share protocol.

## Requirements
To work with android devices (which is the main goal of this project) you need to enable experimental features for bluez. It can be done in `bluetooth.service` systemd service file. Just add `--experimental` to Exec.

## How to use
### Server
```go
func main() {
    server, err := qserver.NewBuilder().
        WithDeviceType(qshare.LaptopDevice).
        Build(authCallback, textCallback, filesCallback)
    if err != nil {
        // ...
    }

    if err = server.Listen(); err != nil {
        // ...
    }

    // decide when to stop server by yourself
    if err = server.Stop(); err != nil {
        // ...
    }
}

func authCallback(text *qshare.TextMeta, files []qshare.FileMeta, pin uint16) bool {
    // print out pin and upcoming payloads here
    return true
}

func textCallback(payload qshare.TextPayload) {
    // do what you want with transferred text payload
}

func filesCallback(payload qshare.FilePayload) {
    var (
        n int
        err error
        chunkBuf = make([]byte, 512 * 1024)
    )

    for {
        n, err = payload.Pr.Read(chunkBuf)
        if err != nil {
            if errors.Is(err, io.EOF) {
                // success transfer, no data arrived, exit loop.
                break
            }
            // ...
        }
        // write chunkBuf somewhere
    }
}
```
### Client
TODO

## Test
`make test`

## TODO
- Refactor all TODOs in codebase
- Add more tests
- Client usage example in README
- Security review
- Security bot to keep track of vulnerabilities in dependent libs
- Codebase statistics (lines of code / files / dependencies count)
