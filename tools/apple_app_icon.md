# Home-screen icon on iPhone (apple-touch-icon)

How to give a web app a real icon when it's saved to the iPhone home screen
(Safari → Share → Add to Home Screen). This is a **reusable pattern** that works the same way in
any app in the ecosystem: one PNG, one `<link>`, one route.

## The problem in one sentence

iOS ignores `<link rel="icon" type="image/svg+xml">`. For the home screen it only uses a **PNG**
from `<link rel="apple-touch-icon">`, and when that link is missing it requests
`/apple-touch-icon.png` on its own. In our apps that path falls through to `GET /` and comes back
as a 404 or HTML, so iOS shows a letter or a screenshot of the page instead of an icon.

## The fix (3 steps)

### 1. The PNG: `build/web/apple-touch-icon.png`

- **180×180**, PNG, **no transparency**.
- **Square, full-bleed, with no rounded corners of its own.** iOS applies its own corner mask.
  If you round the corners yourself (for example `rx="15"` from the favicon), a dark border
  shows up at the corners.
- Use the same artwork as `favicon.svg`, so the tab icon and the home-screen icon match.

### 2. The link in `build/web/index.html` (inside `<head>`, next to the favicon)

```html
<link rel="icon" href="/favicon.svg?v={{.Version}}" type="image/svg+xml" />
<link rel="apple-touch-icon" href="/apple-touch-icon.png?v={{.Version}}" />
```

`?v={{.Version}}` is the usual cache-bust (see `ui-update-delivery.md`).

### 3. The route in `build/main.go` (next to `favicon.svg`)

```go
mux.Handle("GET /favicon.svg", cacheStatic(fileServer))
mux.Handle("GET /apple-touch-icon.png", cacheStatic(fileServer))
```

The file is picked up by `//go:embed web` automatically. Without an explicit route the request
lands in `handleIndex` → 404.

Then the usual cycle: `prod-compose build release` → `run --rm release` → check `Built:` →
commit → deploy.

## Turning the SVG into a PNG

**Gotcha:** ImageMagick's built-in SVG renderer (`MSVG`, used when `rsvg-convert` isn't
installed) **drops strokes**. Rings drawn with `stroke` + `fill="none"` disappear and you get an
empty background square. Always look at the result before committing.

**Option A — `rsvg-convert` is available** (renders SVG correctly):

```bash
# remove rx on the background rect so the square is full-bleed
sed 's/ rx="[0-9.]*"//' build/web/favicon.svg > /tmp/touch.svg
rsvg-convert -w 180 -h 180 -b '#211C18' /tmp/touch.svg -o build/web/apple-touch-icon.png
```

**Option B — ImageMagick only:** draw the primitives directly with `-draw` at 4× size (720 px)
and downscale, which gives clean anti-aliasing. Multiply the coordinates from the 64×64
`viewBox` by `720/64 = 11.25`. This is the command used for nom-nom:

```bash
magick -size 720x720 xc:'#211C18' -fill none \
  -stroke '#B43E2C' -strokewidth 101.25 -draw "circle 360,371.25 573.75,371.25" \
  -stroke '#BE8A60' -strokewidth 56.25  -draw "circle 360,371.25 461.25,371.25" \
  -stroke none -fill '#211C18'          -draw "circle 191.25,191.25 286.875,191.25" \
  -resize 180x180 PNG24:build/web/apple-touch-icon.png
```

`circle cx,cy px,py` takes the centre and **a point on the circle** (`px = cx + r`), not the
radius. `PNG24:` forces RGB with no alpha channel.

Check: `identify build/web/apple-touch-icon.png` → `PNG 180x180`. Then open the file and look
at it.

## Checking on the iPhone

1. After the deploy: `curl -I https://<app>/apple-touch-icon.png` → `200`, `Content-Type: image/png`.
2. **Delete the old shortcut** from the home screen. iOS stores the icon when the shortcut is
   added and never refreshes it.
3. Safari → Share → **Add to Home Screen**. The icon should already show in the preview of that
   dialog.

## Related tags (already in `index.html`)

```html
<meta name="apple-mobile-web-app-capable" content="yes" />      <!-- opens without the Safari UI -->
<meta name="apple-mobile-web-app-title" content="nom-nom" />    <!-- name under the icon -->
<meta name="theme-color" content="#15110E" />
```
