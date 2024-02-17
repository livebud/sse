# SSE

[![Go Reference](https://pkg.go.dev/badge/github.com/livebud/sse.svg)](https://pkg.go.dev/github.com/livebud/sse)

Simple, low-level server-sent event (SSE) handler and client.

## Features

- Easy to build live-reloading on top
- Low-level and customizable
- Comes with an SSE client

## Install

```sh
go get github.com/livebud/sse
```

## Example

On the server-side:

```go
package main

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/livebud/sse"
)

func main() {
	ctx := context.Background()
	log := slog.Default()
	handler := sse.New(log)
	go http.ListenAndServe(":8080", handler)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	count := 0
	for {
		<-ticker.C
		handler.Broadcast(ctx, &sse.Event{
			Data: []byte(strconv.Itoa(count)),
		})
		count++
	}
}
```

From the browser:

```js
const es = new EventSource("/")
// listen for messages
es.addEventListener("message", function (e) {
  console.log("got message", e.data)
})
// cleanup afterwards (too many zombie clients may cause server to hang)
window.addEventListener("beforeunload", function () {
  es && es.close()
})
```

## Contributors

- Matt Mueller ([@mattmueller](https://twitter.com/mattmueller))

## License

MIT
