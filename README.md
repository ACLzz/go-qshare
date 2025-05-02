# go-qshare
GO library to communicate with devices via Quick Share protocol.

## Requirements
To work with android devices (which is the main goal of this project) you need to enable experimental features for bluez. It can be done in `bluetooth.service` systemd service file. Just add `--experimental` to Exec.

## Test
`make test`

## WIP
This project is still in WIP phase and cannot be used. I left it in public to be able to test my own projects with it and don't mess with the `.netrc` files.