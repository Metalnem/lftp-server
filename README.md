# LFTP server [![Build Status](https://travis-ci.org/Metalnem/lftp-server.svg?branch=master)](https://travis-ci.org/Metalnem/lftp-server) [![Go Report Card](https://goreportcard.com/badge/github.com/metalnem/lftp-server)](https://goreportcard.com/report/github.com/metalnem/lftp-server) [![license](https://img.shields.io/badge/license-MIT-blue.svg?style=flat)](https://raw.githubusercontent.com/metalnem/lftp-server/master/LICENSE)
JSON RPC server for [LFTP](https://lftp.yar.ru/).

## Description

LFTP server is a HTTP server that accepts FTP download requests in JSON format and forwards them to LFTP. Requests have the following format:

```json
{
	"path": "...",
	"username": "...",
	"password": "...",
	"secret": "..."
}
```

Path is FTP file or directory to be downloaded, username/password are optional FTP credentials, and secret must match --rpc-secret parameter used when server was started.

If the request was valid, server sends 200 OK response to the client, with the ID  of the created job. Response looks like this:

```json
{
	"id": "..."
}
```

In case of any error, server sends 400 Bad Request to the client, with the error message in the following format:

```json
{
	"message": "..."
}
```

If FTP credentials are missing or invalid, server responds to the client with 401 Unauthorized status code.

## Installation

```
$ go get github.com/metalnem/lftp-server
```

## Binaries (x64)

[Linux](https://github.com/Metalnem/lftp-server/releases/download/v1.2.0/lftp-server-linux64-1.2.0.zip)  
[Mac OS X](https://github.com/Metalnem/lftp-server/releases/download/v1.2.0/lftp-server-darwin64-1.2.0.zip)

## Usage

```
$ ./lftp-server
Usage of ./lftp-server:
  -n int
	  Number of connections to use when downloading single file. Possible values: 1-100 (default 4)
  -o string
	  Output directory (optional, default value is the current working directory)
  -p int
	  Number of files to download in parallel when mirroring directories. Possible values: 1-10 (default 1)
  -rpc-listen-port int
	  Specify a port number for JSON-RPC server to listen to. Possible values: 1024-65535 (default 7800)
  -rpc-secret string
	  Set RPC secret authorization token (required)
  -s string
	  Script to run after successful download
```

## Example

```
$ ./lftp-server -rpc-listen-port 7800 -rpc-secret SECRET -n 4 -p 1 -o ~/Downloads -s ./script.sh
```
