// fancy-hello: a small showcase of Element (HTML built in Go, no templates)
// served by RWeb.
//
//	GET /             -> "Hello, World!" on an animated glass card
//	GET /hello/:name  -> the same card, greeting :name
//
// Everything is plain Go with no template files, no JS build and no assets on
// disk. The "fancy" part is pure CSS (animated gradient, glassmorphism,
// shimmering text, staggered pop-in), so the page works with JS disabled
// apart from the small name form.
package main

import (
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/rohanthewiz/element"
	"github.com/rohanthewiz/rweb"
)

// "Hello" in several languages, shown as pills that pop in one after another.
var greetings = []string{
	"Hello", "Hola", "Bonjour", "Ciao", "Hallo", "Olá",
	"こんにちは", "नमस्ते", "Jambo", "Привет", "안녕하세요", "Merhaba",
}

func main() {
	s := rweb.NewServer(rweb.ServerOptions{Address: ":8080", Verbose: true})
	s.Use(rweb.RequestInfo) // one log line per request: method, path, status, duration

	s.Get("/", func(ctx rweb.Context) error {
		return ctx.WriteHTML(page("World"))
	})

	s.Get("/hello/:name", func(ctx rweb.Context) error {
		// RWeb v0.2.0 hands path params over still percent-encoded
		// (/hello/Ada%20Lovelace -> "Ada%20Lovelace"), so decode here. On a bad
		// escape ("50%") keep the raw value rather than failing the request.
		// NOTE: v0.2.1+ decodes params itself; drop this when upgrading, or a
		// literal "%2520" would be decoded twice.
		name := ctx.Request().PathParam("name")
		if decoded, err := url.PathUnescape(name); err == nil {
			name = decoded
		}
		return ctx.WriteHTML(page(name))
	})

	log.Println("Open http://localhost:8080 or http://localhost:8080/hello/Gopher")
	log.Fatal(s.Run()) // Run blocks; it shuts down gracefully on SIGINT/SIGTERM
}

// page renders the whole document for one request.
//
// A pooled builder avoids a fresh buffer allocation per request. The result is
// returned via b.String(), which copies, rather than b.Bytes(), which aliases
// the builder's buffer. That buffer goes back to the pool (and may be reused by
// another request) as soon as this function returns, so handing out the
// aliased slice would risk a corrupted response.
func page(name string) string {
	b := element.AcquireBuilder()
	defer element.ReleaseBuilder(b)

	b.Html().R( // Html() also emits <!DOCTYPE html>
		cachedHead.Render(b),
		b.Body().R(
			b.MainClass("card").R(
				// `name` comes from the URL, so it must go through TE() (HTML-escaped).
				// Plain T() would let /hello/<script>... inject markup.
				b.H1().R(
					b.T("Hello, "),
					b.SpanClass("name").TE(name),
					b.T("!"),
				),

				// Each pill gets its index as the CSS custom property --i; the stylesheet
				// turns that into animation-delay, giving a stagger with no JS.
				// element.ForEach has no index, so a plain loop runs inside Wrap().
				b.UlClass("greetings").R(
					b.Wrap(func() {
						for i, g := range greetings {
							b.Li("style", "--i:"+strconv.Itoa(i)).T(g)
						}
					}),
				),

				// A small progressive enhancement: build /hello/<name> on the client.
				// encodeURIComponent keeps names with spaces or slashes routable.
				// The JS uses single quotes because Element escapes `"` in attribute values.
				b.Form("onsubmit", "event.preventDefault();"+
					"location.href='/hello/'+encodeURIComponent(this.n.value.trim()||'World')").R(
					b.Input("name", "n", "placeholder", "Your name", "aria-label", "Your name",
						"autocomplete", "off").R(),
					b.Button("type", "submit").T("Greet me ✨"),
				),

				b.PClass("foot").F("Built with Element · Served by RWeb · %s",
					time.Now().Format("15:04:05")),
			),
		),
	)
	return b.String()
}

// head holds the static <head>: metadata and the stylesheet. It is identical
// for every request, which makes it a good fit for element.Cached below.
type head struct{}

func (head) Render(b *element.Builder) (x any) {
	b.Head().R(
		b.Meta("charset", "utf-8").R(),
		b.Meta("name", "viewport", "content", "width=device-width, initial-scale=1").R(),
		b.Title().T("Fancy Hello"),
		b.Style().T(css), // our own CSS literal, so verbatim T() is correct here
	)
	return
}

// cachedHead renders <head> once and replays the bytes on every later request.
// It is safe for concurrent handlers. It must stay static: Cached never invalidates.
var cachedHead = element.Cached(head{})

// css is the entire visual treatment:
//   - body:      a 4-colour gradient, 400% oversized, slowly panned by `drift`
//   - .card:     frosted glass via backdrop-filter, with a translucent border
//   - .name:     gradient-clipped text with a moving highlight (`shine`)
//   - li:        pills start invisible and `pop` in, delayed by --i * 90ms
//   - reduced-motion users get the final, static state with no animation
const css = `
* { box-sizing: border-box; margin: 0; }
body {
  min-height: 100vh; display: grid; place-items: center; padding: 1rem;
  font-family: ui-rounded, system-ui, -apple-system, sans-serif; color: #fff;
  background: linear-gradient(-45deg, #ee7752, #e73c7e, #23a6d5, #23d5ab);
  background-size: 400% 400%; animation: drift 14s ease infinite;
}
@keyframes drift { 50% { background-position: 100% 50%; } }

.card {
  max-width: 44rem; padding: 3rem 3.5rem; border-radius: 28px; text-align: center;
  background: rgba(255,255,255,.14); border: 1px solid rgba(255,255,255,.3);
  backdrop-filter: blur(18px); -webkit-backdrop-filter: blur(18px);
  box-shadow: 0 24px 70px rgba(0,0,0,.25);
}
h1 { font-size: clamp(2.4rem, 7vw, 4.8rem); font-weight: 800; letter-spacing: -.03em; }
.name {
  background: linear-gradient(90deg, #fff 20%, #ffe29a 50%, #fff 80%);
  background-size: 200% auto; -webkit-background-clip: text; background-clip: text;
  color: transparent; animation: shine 3s linear infinite;
}
@keyframes shine { to { background-position: -200% center; } }

.greetings {
  list-style: none; padding: 0; margin: 1.6rem 0 2rem;
  display: flex; flex-wrap: wrap; gap: .5rem; justify-content: center;
}
.greetings li {
  padding: .35rem .9rem; border-radius: 999px; background: rgba(255,255,255,.2);
  opacity: 0; animation: pop .5s cubic-bezier(.2,.9,.3,1.4) forwards;
  animation-delay: calc(var(--i) * 90ms);
}
@keyframes pop {
  from { opacity: 0; transform: translateY(14px) scale(.85); }
  to   { opacity: 1; transform: none; }
}

form { display: flex; gap: .5rem; justify-content: center; flex-wrap: wrap; }
input, button { font: inherit; padding: .65rem 1.1rem; border-radius: 14px; border: 0; }
input  { background: rgba(255,255,255,.92); color: #333; min-width: 14rem; }
button { background: #16161d; color: #fff; cursor: pointer; transition: transform .15s, box-shadow .15s; }
button:hover { transform: translateY(-2px); box-shadow: 0 8px 20px rgba(0,0,0,.3); }
.foot { margin-top: 1.8rem; font-size: .8rem; opacity: .75; }

@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after { animation: none !important; transition: none !important; }
  .greetings li { opacity: 1; }
}
`
