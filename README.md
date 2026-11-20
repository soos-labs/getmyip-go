# getmyip-go

Tiny, dependency-free Go client **and CLI** for the [getmyip.pro](https://getmyip.pro) IP geolocation API — privacy-first (no logs, no trackers), with ip-api-compatible JSON.

> Anonymous lookups: 60 requests/min and 10,000/day, no key. Batch needs a key.
> Sign in at https://getmyip.pro/login for one key: 600/min, 100,000/day,
> batch up to 1,000 IPs. Revoke the current key to create another.

[![Test](https://github.com/soos-labs/getmyip-go/actions/workflows/test.yml/badge.svg)](https://github.com/soos-labs/getmyip-go/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/soos-labs/getmyip-go.svg)](https://pkg.go.dev/github.com/soos-labs/getmyip-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/soos-labs/getmyip-go)](https://goreportcard.com/report/github.com/soos-labs/getmyip-go)
![License: MIT](https://img.shields.io/badge/license-MIT-blue)

## CLI

```sh
go install github.com/soos-labs/getmyip-go/cmd/getmyip@latest
```

```console
$ getmyip
IP:       45.82.64.40
Location: Naaldwijk, Netherlands (NL)
Network:  WorldStream B.V. (AS49981)
Timezone: Europe/Amsterdam

$ getmyip 8.8.8.8
IP:       8.8.8.8
Location: United States (US)
Network:  Google LLC (AS15169)

$ getmyip -field org 8.8.8.8
Google LLC

$ getmyip -json 8.8.8.8
{ "ip": "8.8.8.8", "country": "United States", ... }
```

An API key is optional (the anonymous tier needs none); pass `-key` or set `GETMYIP_API_KEY` for higher limits. Point the CLI at a self-hosted instance with `GETMYIP_BASE_URL`.

## Library

```sh
go get github.com/soos-labs/getmyip-go
```

```go
package main

import (
	"context"
	"fmt"

	getmyip "github.com/soos-labs/getmyip-go"
)

func main() {
	c := getmyip.New() // getmyip.WithAPIKey("...") for higher limits

	me, _ := c.Me(context.Background())
	fmt.Println(me.IP, me.Country)

	r, _ := c.Lookup(context.Background(), "8.8.8.8")
	fmt.Println(r.Org, r.ASN) // Google LLC 15169

	batch, _ := c.Batch(context.Background(), []string{"8.8.8.8", "1.1.1.1"})
	fmt.Println(len(batch))
}
```

## API

Base URL `https://api.getmyip.pro` — see the [API docs](https://docs.getmyip.pro). Endpoints: `GET /v1/json` (your IP), `GET /v1/{ip}`, `POST /v1/batch`. The response is ip-api-compatible JSON, so migrating from `ip-api.com` is usually a one-line change.

## License

MIT © Soos Labs

## Related

- Python client/CLI: [soos-labs/getmyip-python](https://github.com/soos-labs/getmyip-python)
- Node.js client/CLI: [soos-labs/getmyip-js](https://github.com/soos-labs/getmyip-js)
- API docs: [docs.getmyip.pro](https://docs.getmyip.pro)
