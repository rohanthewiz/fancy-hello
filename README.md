# fancy-hello

A short but fancy Hello World in Go, built with
[Element](https://github.com/rohanthewiz/element) (HTML generated in Go, no templates)
and [RWeb](https://github.com/rohanthewiz/rweb) (a light web server).

A frosted-glass card sits on a background gradient that slowly shifts colour. It
greets you by name with a shimmer, and "Hello" pops in across 12 languages. The
effects are pure CSS, and they stay still for visitors who prefer reduced motion.

## Run

```sh
go run .
```

Then open:

- http://localhost:8080: Hello, World!
- http://localhost:8080/hello/Gopher: Hello, Gopher! (or type a name in the form)

Requires Go 1.25+.

## What it shows off

| Library | Feature |
| ------- | ------- |
| Element | Builds the page as nested Go calls that mirror the HTML tree |
| Element | Reuses pooled builders across requests (`AcquireBuilder` / `ReleaseBuilder`) |
| Element | Renders the static `<head>` once with `element.Cached`, then replays it |
| Element | HTML-escapes the name from the URL with `TE()`, so `/hello/<script>` stays text |
| RWeb    | A plain route plus a `/hello/:name` route with a path parameter |
| RWeb    | Logs one line per request with `rweb.RequestInfo` |

Everything lives in [`main.go`](main.go).
