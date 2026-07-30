# Signal design system

Signal is the Tree Lot Scheduler's small, server-rendered design system. It
implements the approved light visual direction with semantic Tailwind tokens,
typed Go presentation models, and reusable `html/template` definitions. It is
an inbound web-adapter concern: components display application decisions but
never make authorization, eligibility, capacity, timing, or financial
decisions.

## Local use

Build the generated stylesheet after changing a template or source style:

```sh
make assets
```

Use `make assets-watch` during visual work. Run the web process with the
development gallery enabled:

```sh
APP_ENV=development go run ./cmd/web
```

Then open `http://localhost:8080/_dev/components`. The gallery uses fictional,
non-sensitive fixtures and the same typed models and templates as application
pages. `/_dev/components` and `/_dev/parity` are not registered unless
`APP_ENV=development`; production requests receive `404 Not Found`.

Node.js and npm are build-time tools only. Tailwind and its CLI are pinned in
`package-lock.json`. The generated `web/static/app.css` is committed and
embedded into the Go binary by `web/static`, so deployed processes do not need
Node or files from `node_modules`.

## Structure

- `web/styles/app.css` — Tailwind source, semantic theme tokens, component
  styles, breakpoints, safe-area behavior, and reduced-motion treatment.
- `web/static/app.css` — generated and minified CSS; do not edit by hand.
- `web/static/assets.go` — embeds generated assets and exposes their HTTP
  handler.
- `internal/web/views/models.go` — typed primitive and domain display models.
- `internal/web/views/templates/components.gohtml` — reusable semantic template
  definitions.
- `internal/web/views/templates/gallery.gohtml` — page shell, development
  gallery, and full-page/fragment parity examples.
- `internal/web/handlers` — normal HTTP routing, development gating, response
  selection, and browser security headers.

## Component API rules

Handlers and application presenters construct purpose-built view models. They
must not pass persistence records into templates. Variants use semantic names
such as `complete`, `warning`, `critical`, and `provisional`; literal color
names are not public component vocabulary.

Template components own their markup and accessibility contract:

- Native elements are preferred over ARIA substitutes.
- Fields have visible labels and associated hints or errors.
- Status always includes text and a shape or icon; color is supplemental.
- Icon-only controls require an accessible name. Decorative icons are hidden
  from assistive technology.
- Tables include captions and scoped headers, plus structured mobile rows when
  columns cannot fit.
- Full-page forms and links are the baseline. HTMX may request a fragment from
  the same handler and view model, but headers do not grant authority.
- Dialogs are enhancement surfaces only; application workflows also need a
  complete server-rendered confirmation page.

The domain patterns are display contracts. For example, a `PersonOption`
receives a precomputed disabled state and explanation; it does not determine
whether the person may sign up. Likewise, `ScoutBucksSummary` formats a
calculation supplied by the Reporting context and never calculates awards.

## Extension policy

Before adding a component:

1. Reuse a primitive or existing domain pattern.
2. Extend a semantic variant when behavior and accessibility are the same.
3. Add a component only for repeated behavior or a stable domain concept.
4. Add the smallest failing render or HTTP test, observe it fail, implement the
   behavior, and refactor only while tests are green.
5. Add all meaningful states to `GalleryData` using fictional content.
6. Verify narrow-phone, tablet, desktop, keyboard, visible-focus, 200% text
   zoom, reduced-motion, and status-without-color behavior.
7. Run `make assets`, focused tests, and finally `make ci`.

Do not copy markup into feature templates to create one-off visual variants.
Do not add a JavaScript application framework such as React or Angular. Focused
browser JavaScript is expected for WebAuthn passkeys and may be used for HTMX
or other browser APIs, provided server-side authorization and validation remain
authoritative.

## Current verification boundary

Go tests cover rendering, escaping, route gating, full-page/fragment parity,
embedded assets, landmarks, labels, table semantics, and representative
responsive/reduced-motion CSS contracts. Manual VoiceOver on Safari and
TalkBack on Chrome, automated browser accessibility tooling, focus trapping
for an interactive dialog controller, and visual-regression baselines remain
release-stage work when browser infrastructure is introduced.
