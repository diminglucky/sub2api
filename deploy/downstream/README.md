# Draw Downstream Pilot Deployment

This pilot keeps one Sub2API backend and one database. The backend embeds both
frontends:

- apex `superai.sbs` serves the existing main frontend.
- `draw.superai.sbs` serves `frontend-downstream`.
- `https://draw.superai.sbs/v1/*` goes to the existing OpenAI-compatible
  gateway without path rewriting.

## Build

The image build now runs both frontend builds and embeds them into the Go
binary:

```powershell
docker compose --env-file .env `
  -f deploy/docker-compose.local.yml `
  -f deploy/docker-compose.superai.yml `
  build sub2api
```

## Edge Routing

Use `cloudflared.example.yml` as the tunnel ingress template. The important
parts are:

- keep the public host `draw.superai.sbs`;
- proxy all paths for that host to `http://sub2api:8080`;
- do not strip or rewrite `/v1`;
- preserve `Host` with `httpHostHeader` so backend subsite resolution sees
  `draw.superai.sbs`.

If Cloudflare is configured manually instead of with a tunnel config, create a
proxied DNS record for `draw.superai.sbs` and route it to the same origin as the
apex domain. Do not create a path-specific rewrite for `/v1`.

## Checks

Run these after DNS, TLS, and the container are live:

```bash
curl -sS -o /tmp/draw-root.html -w '%{http_code} %{content_type}\n' https://draw.superai.sbs/
grep -qi '<div id="app"></div>' /tmp/draw-root.html
```

Expected: `200 text/html` and the downstream SPA shell. This must not be the
main-site HTML.

```bash
curl -sS -D - https://draw.superai.sbs/robots.txt -o /tmp/draw-robots.txt
grep -q 'User-agent: \*' /tmp/draw-robots.txt
```

Expected: `200` with plain text and the downstream `User-agent: *` rule.

```bash
curl -sS -D - https://draw.superai.sbs/v1/models -o /tmp/draw-v1.json
grep -q '"error"' /tmp/draw-v1.json
```

Expected: an OpenAI-compatible JSON error without an API key. The important
failure mode to catch is an HTML `404` or the downstream SPA HTML: that means
`/v1` was routed to the frontend instead of the gateway.

## Rollback

To roll back the pilot, remove the `draw.superai.sbs` DNS/tunnel route. The apex
domain and main backend stay unchanged.
