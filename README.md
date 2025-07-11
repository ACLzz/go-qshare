# go-qshare
[![codecov](https://codecov.io/github/ACLzz/go-qshare/graph/badge.svg?token=S39GK295NH)](https://codecov.io/github/ACLzz/go-qshare)
GO library to communicate with devices via Quick Share protocol.

## Requirements
To work with android devices (which is the main goal of this project) you need to enable experimental features for bluez. It can be done in `bluetooth.service` systemd service file. Just add `--experimental` to Exec.

## Test
`make test`

## TODO
- CI/CD Pipeline
- CI/CD Widget
- Code coverage widget
- Refactor all TODOs in codebase
- Add more tests
- Better library incapsulation
- Usage example in README
- Docs
- Security review
