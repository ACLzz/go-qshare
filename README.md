# go-qshare
[![codecov](https://codecov.io/github/ACLzz/go-qshare/graph/badge.svg?token=S39GK295NH)](https://codecov.io/github/ACLzz/go-qshare)<br/>
GO library to communicate with devices via Quick Share protocol.

## Requirements
To work with android devices (which is the main goal of this project) you need to enable experimental features for bluez. It can be done in `bluetooth.service` systemd service file. Just add `--experimental` to Exec.

## Test
`make test`

## TODO
- Refactor all TODOs in codebase
- Add more tests
- Usage example in README
- Docs
- Security review
- Security bot to keep track of vulnerabilities in dependent libs
- Codebase statistics (lines of code / files / dependencies count)
