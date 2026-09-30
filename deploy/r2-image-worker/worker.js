/**
 * Serve an R2 bucket over a custom hostname.
 *
 * Why this exists: R2 "Custom Domains" require the domain's zone to live in the
 * same Cloudflare account as the bucket. When the zone and the bucket are in
 * different accounts, this Worker (deployed in the zone's account) signs
 * requests to the bucket's S3 endpoint, so `img.dihappy.cfd` works without
 * moving either account.
 *
 * Required environment variables (Worker -> Settings -> Variables and Secrets):
 *   R2_ENDPOINT           https://<account_id>.r2.cloudflarestorage.com
 *   R2_BUCKET             superai
 *   R2_ACCESS_KEY_ID      R2 API token access key id (Object Read on the bucket)
 *   R2_SECRET_ACCESS_KEY  R2 API token secret (mark as "Secret")
 *
 * Route: connect the Worker to a custom domain, e.g. img.dihappy.cfd
 *
 * The app then sets public_base_url = https://img.dihappy.cfd and object keys
 * look like images/<user_id>/<uuid>.png, so a public URL becomes
 * https://img.dihappy.cfd/images/<user_id>/<uuid>.png
 */

const REGION = 'auto'
const SERVICE = 's3'
const UNSIGNED_PAYLOAD = 'UNSIGNED-PAYLOAD'
const CACHE_SECONDS = 86400

export default {
  async fetch(request, env) {
    if (request.method !== 'GET' && request.method !== 'HEAD') {
      return new Response('Method Not Allowed', {
        status: 405,
        headers: { allow: 'GET, HEAD' },
      })
    }

    const endpoint = String(env.R2_ENDPOINT || '').replace(/\/+$/, '')
    const bucket = String(env.R2_BUCKET || '').trim()
    const accessKeyId = String(env.R2_ACCESS_KEY_ID || '').trim()
    const secretAccessKey = String(env.R2_SECRET_ACCESS_KEY || '').trim()
    if (!endpoint || !bucket || !accessKeyId || !secretAccessKey) {
      return new Response('Worker is not configured', { status: 500 })
    }

    const key = decodeURIComponent(new URL(request.url).pathname).replace(/^\/+/, '')
    if (!key) {
      return new Response('Not Found', { status: 404 })
    }

    const target = new URL(`${endpoint}/${bucket}/${encodePath(key)}`)
    const headers = await signRequest({
      method: request.method,
      url: target,
      accessKeyId,
      secretAccessKey,
    })

    const upstream = await fetch(target.toString(), {
      method: request.method,
      headers,
      cf: { cacheEverything: true, cacheTtl: CACHE_SECONDS },
    })

    const out = new Headers()
    for (const name of ['content-type', 'content-length', 'etag', 'last-modified']) {
      const value = upstream.headers.get(name)
      if (value) out.set(name, value)
    }
    out.set('cache-control', `public, max-age=${CACHE_SECONDS}`)
    // Objects are images; allow them to be read from the panel origin.
    out.set('access-control-allow-origin', '*')

    return new Response(request.method === 'HEAD' ? null : upstream.body, {
      status: upstream.status,
      headers: out,
    })
  },
}

function encodePath(key) {
  return key
    .split('/')
    .map((segment) => encodeURIComponent(segment))
    .join('/')
}

async function signRequest({ method, url, accessKeyId, secretAccessKey }) {
  const amzDate = new Date().toISOString().replace(/[:-]|\.\d{3}/g, '')
  const dateStamp = amzDate.slice(0, 8)
  const canonicalHeaders =
    `host:${url.host}\n` +
    `x-amz-content-sha256:${UNSIGNED_PAYLOAD}\n` +
    `x-amz-date:${amzDate}\n`
  const signedHeaders = 'host;x-amz-content-sha256;x-amz-date'
  const canonicalRequest = [
    method,
    url.pathname,
    '',
    canonicalHeaders,
    signedHeaders,
    UNSIGNED_PAYLOAD,
  ].join('\n')

  const scope = `${dateStamp}/${REGION}/${SERVICE}/aws4_request`
  const stringToSign = [
    'AWS4-HMAC-SHA256',
    amzDate,
    scope,
    await sha256Hex(canonicalRequest),
  ].join('\n')

  const signature = toHex(await hmac(await deriveSigningKey(secretAccessKey, dateStamp), stringToSign))

  return {
    // Do not set `host` explicitly: Workers derives it from the URL, and the
    // signature above already covers that value.
    'x-amz-content-sha256': UNSIGNED_PAYLOAD,
    'x-amz-date': amzDate,
    authorization:
      `AWS4-HMAC-SHA256 Credential=${accessKeyId}/${scope}, ` +
      `SignedHeaders=${signedHeaders}, Signature=${signature}`,
  }
}

async function deriveSigningKey(secretAccessKey, dateStamp) {
  let key = new TextEncoder().encode(`AWS4${secretAccessKey}`)
  for (const part of [dateStamp, REGION, SERVICE, 'aws4_request']) {
    key = await hmac(key, part)
  }
  return key
}

async function hmac(keyBytes, value) {
  const key = await crypto.subtle.importKey(
    'raw',
    keyBytes,
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['sign'],
  )
  const signature = await crypto.subtle.sign('HMAC', key, new TextEncoder().encode(value))
  return new Uint8Array(signature)
}

async function sha256Hex(value) {
  const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(value))
  return toHex(new Uint8Array(digest))
}

function toHex(bytes) {
  return [...bytes].map((byte) => byte.toString(16).padStart(2, '0')).join('')
}
