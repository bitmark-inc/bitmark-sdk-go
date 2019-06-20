# Bitmark SDK for Golang
The official Bitmark SDK for Golang

[![Build Status](https://travis-ci.org/bitmark-inc/bitmark-sdk-go.svg?branch=master)](https://travis-ci.org/bitmark-inc/bitmark-sdk-go)

## Setting Up

### Prerequisites

- Golang version 1.7+

### Installing

#### Go Module

In `go.mod` file:
```sh
require (
	github.com/bitmark-inc/bitmark-sdk-go
)
```

#### Go Vendor
```sh
govendor fetch github.com/bitmark-inc/bitmark-sdk-go
```

#### Manually
```sh
go get github.com/bitmark-inc/bitmark-sdk-go
```

## Documentation

Please refer to our [SDK Document](https://sdk-docs.bitmark.com/).


## Sample code
This is a [sample project](sample/). It shows how to use Bitmark SDK for Golang.

## Opening Issues
If you encounter a bug with the Bitmark SDK for Golang we would like to hear from you. Search the existing issues and try to make sure your problem doesn’t exist yet before opening a new issue. It’s helpful if you could provide the version of the SDK, Golang and OS you’re using. Please include a stack trace and reproducible case if possible.


## License

Copyright (c) 2014-2019 Bitmark Inc (support@bitmark.com).

Permission to use, copy, modify, and distribute this software for any
purpose with or without fee is hereby granted, provided that the above
copyright notice and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
