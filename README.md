# LFTP server [![Build Status](https://travis-ci.org/Metalnem/lftp-server.svg?branch=master)](https://travis-ci.org/Metalnem/lftp-server) [![Go Report Card](https://goreportcard.com/badge/github.com/metalnem/lftp-server)](https://goreportcard.com/report/github.com/metalnem/lftp-server) [![license](https://img.shields.io/badge/license-MIT-blue.svg?style=flat)](https://raw.githubusercontent.com/metalnem/lftp-server/master/LICENSE)
JSON RPC server for LFTP.

## Installation

```
$ go get github.com/metalnem/lftp-server
```

## Binaries (x64)

[Linux](https://github.com/Metalnem/lftp-server/releases/download/v1.0.0/lftp-server-linux64-1.0.0.zip)  
[Mac OS X](https://github.com/Metalnem/lftp-server/releases/download/v1.0.0/lftp-server-darwin64-1.0.0.zip)

## Usage

```
$ ./lftp-server
Usage of ./lftp-server:
  -rpc-listen-port int
    	Specify a port number for JSON-RPC server to listen to. Possible values: 1024-65535 (default 7800)
  -rpc-secret string
    	Set RPC secret authorization token (required)
```

## Example

```
$ ./lftp-server --rpc-listen-port=7801 --rpc-secret=SECRET
```
