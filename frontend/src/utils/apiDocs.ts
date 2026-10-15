export type DocsEndpoint =
  | 'models'
  | 'chat'
  | 'responses'
  | 'messages'
  | 'embeddings'
  | 'imagesGenerations'
  | 'imagesEdits'
  | 'gemini'
export type DocsLanguage = 'curl' | 'python' | 'javascript'

export interface DocsField {
  name: string
  description: string
}

export interface DocsEndpointSpec {
  method: 'GET' | 'POST'
  path: string
  label: string
  descriptionKey: string
  fields: DocsField[]
  responseExample: string
}

export const docsEndpointOrder: DocsEndpoint[] = [
  'models',
  'chat',
  'responses',
  'messages',
  'embeddings',
  'imagesGenerations',
  'imagesEdits',
  'gemini'
]

export const docsEndpoints: Record<DocsEndpoint, DocsEndpointSpec> = {
  models: {
    method: 'GET',
    path: '/models',
    label: 'Models',
    descriptionKey: 'apiDocs.endpointModelsDescription',
    fields: [],
    responseExample: `{
  "object": "list",
  "data": [
    { "id": "gpt-5.6-sol", "object": "model", "owned_by": "openai" },
    { "id": "claude-sonnet-4-5", "object": "model", "owned_by": "anthropic" }
  ]
}`
  },
  chat: {
    method: 'POST',
    path: '/chat/completions',
    label: 'Chat Completions',
    descriptionKey: 'apiDocs.endpointChatDescription',
    fields: [
      { name: 'model', description: 'apiDocs.fieldModelRequired' },
      { name: 'messages', description: 'apiDocs.fieldMessages' },
      { name: 'temperature', description: 'apiDocs.fieldTemperature' },
      { name: 'max_tokens', description: 'apiDocs.fieldMaxTokensOptional' },
      { name: 'tools', description: 'apiDocs.fieldTools' },
      { name: 'stream', description: 'apiDocs.fieldStream' }
    ],
    responseExample: `{
  "id": "chatcmpl_...",
  "object": "chat.completion",
  "model": "gpt-5.6-sol",
  "choices": [
    {
      "index": 0,
      "message": { "role": "assistant", "content": "Hello!" },
      "finish_reason": "stop"
    }
  ]
}`
  },
  responses: {
    method: 'POST',
    path: '/responses',
    label: 'Responses',
    descriptionKey: 'apiDocs.endpointResponsesDescription',
    fields: [
      { name: 'model', description: 'apiDocs.fieldModelRequired' },
      { name: 'input', description: 'apiDocs.fieldInput' },
      { name: 'instructions', description: 'apiDocs.fieldInstructions' },
      { name: 'max_output_tokens', description: 'apiDocs.fieldMaxOutputTokens' },
      { name: 'previous_response_id', description: 'apiDocs.fieldPreviousResponseId' },
      { name: 'tools', description: 'apiDocs.fieldTools' },
      { name: 'stream', description: 'apiDocs.fieldStream' }
    ],
    responseExample: `{
  "id": "resp_...",
  "object": "response",
  "status": "completed",
  "model": "gpt-5.6-sol",
  "output": [
    {
      "type": "message",
      "role": "assistant",
      "content": [{ "type": "output_text", "text": "Hello!" }]
    }
  ]
}`
  },
  messages: {
    method: 'POST',
    path: '/messages',
    label: 'Messages (Claude)',
    descriptionKey: 'apiDocs.endpointMessagesDescription',
    fields: [
      { name: 'model', description: 'apiDocs.fieldModelRequired' },
      { name: 'max_tokens', description: 'apiDocs.fieldMaxTokensRequired' },
      { name: 'system', description: 'apiDocs.fieldSystem' },
      { name: 'messages', description: 'apiDocs.fieldMessages' },
      { name: 'tools', description: 'apiDocs.fieldTools' },
      { name: 'stream', description: 'apiDocs.fieldStream' }
    ],
    responseExample: `{
  "id": "msg_...",
  "type": "message",
  "role": "assistant",
  "model": "claude-sonnet-4-5",
  "stop_reason": "end_turn",
  "content": [{ "type": "text", "text": "Hello!" }]
}`
  },
  embeddings: {
    method: 'POST',
    path: '/embeddings',
    label: 'Embeddings',
    descriptionKey: 'apiDocs.endpointEmbeddingsDescription',
    fields: [
      { name: 'model', description: 'apiDocs.fieldModelRequired' },
      { name: 'input', description: 'apiDocs.fieldEmbeddingInput' },
      { name: 'encoding_format', description: 'apiDocs.fieldEncodingFormat' },
      { name: 'dimensions', description: 'apiDocs.fieldDimensions' },
      { name: 'user', description: 'apiDocs.fieldUser' }
    ],
    responseExample: `{
  "object": "list",
  "model": "text-embedding-3-small",
  "data": [
    { "object": "embedding", "index": 0, "embedding": [0.0123, -0.0456] }
  ],
  "usage": { "prompt_tokens": 8, "total_tokens": 8 }
}`
  },
  imagesGenerations: {
    method: 'POST',
    path: '/images/generations',
    label: 'Images',
    descriptionKey: 'apiDocs.endpointImagesDescription',
    fields: [
      { name: 'model', description: 'apiDocs.fieldModelRequired' },
      { name: 'prompt', description: 'apiDocs.fieldPromptRequired' },
      { name: 'size', description: 'apiDocs.fieldImageSize' },
      { name: 'quality', description: 'apiDocs.fieldImageQuality' },
      { name: 'background', description: 'apiDocs.fieldImageBackground' },
      { name: 'n', description: 'apiDocs.fieldImageCount' },
      { name: 'output_format', description: 'apiDocs.fieldOutputFormat' },
      { name: 'response_format', description: 'apiDocs.fieldResponseFormat' }
    ],
    responseExample: `{
  "created": 1760000000,
  "data": [
    { "b64_json": "iVBORw0KGgo..." }
  ]
}`
  },
  imagesEdits: {
    method: 'POST',
    path: '/images/edits',
    label: 'Image Edits',
    descriptionKey: 'apiDocs.endpointImageEditsDescription',
    fields: [
      { name: 'model', description: 'apiDocs.fieldModelRequired' },
      { name: 'image', description: 'apiDocs.fieldImageRequired' },
      { name: 'prompt', description: 'apiDocs.fieldPromptRequired' },
      { name: 'size', description: 'apiDocs.fieldImageSize' },
      { name: 'quality', description: 'apiDocs.fieldImageQuality' },
      { name: 'n', description: 'apiDocs.fieldImageCount' },
      { name: 'output_format', description: 'apiDocs.fieldOutputFormat' }
    ],
    responseExample: `{
  "created": 1760000000,
  "data": [
    { "b64_json": "iVBORw0KGgo..." }
  ]
}`
  },
  gemini: {
    method: 'POST',
    path: '/v1beta/models/gemini-2.5-flash:generateContent',
    label: 'Gemini',
    descriptionKey: 'apiDocs.endpointGeminiDescription',
    fields: [
      { name: 'contents', description: 'apiDocs.fieldContents' },
      { name: 'generationConfig', description: 'apiDocs.fieldGenerationConfig' },
      { name: 'tools', description: 'apiDocs.fieldTools' },
      { name: 'systemInstruction', description: 'apiDocs.fieldSystemInstruction' }
    ],
    responseExample: `{
  "candidates": [
    {
      "content": { "role": "model", "parts": [{ "text": "Hello!" }] },
      "finishReason": "STOP"
    }
  ]
}`
  }
}

export const DEFAULT_DOCS_API_BASE_URL = 'https://api.superai.sbs/v1'

/** Public gateway URL, independent of the dashboard's /api/v1 management API. */
export function resolveDocsApiBaseUrl(configured: string | undefined, origin: string): string {
  const trimmed = (configured || '').trim()
  if (!trimmed) return DEFAULT_DOCS_API_BASE_URL
  try {
    const url = new URL(trimmed, origin)
    if (!['http:', 'https:'].includes(url.protocol)) return DEFAULT_DOCS_API_BASE_URL
    url.search = ''
    url.hash = ''
    const host = url.hostname.toLowerCase()
    const path = url.pathname.replace(/\/+$/, '')
    if (isLocalHost(host)) return DEFAULT_DOCS_API_BASE_URL
    if ((host === 'superai.sbs' || host === 'www.superai.sbs') && (path === '' || path === '/v1')) {
      return DEFAULT_DOCS_API_BASE_URL
    }
    url.pathname = path.endsWith('/v1') ? path : `${path}/v1`
    return url.toString().replace(/\/$/, '')
  } catch {
    return DEFAULT_DOCS_API_BASE_URL
  }
}

function isLocalHost(host: string): boolean {
  return ['localhost', '127.0.0.1', '::1', '0.0.0.0'].includes(host)
}

export function resolveDocsRootUrl(baseUrl: string): string {
  return baseUrl.replace(/\/+$/, '').replace(/\/v1$/, '')
}

export function resolveDocsRequestUrl(baseUrl: string, endpoint: DocsEndpoint): string {
  const spec = docsEndpoints[endpoint]
  return endpoint === 'gemini' ? `${resolveDocsRootUrl(baseUrl)}${spec.path}` : `${baseUrl}${spec.path}`
}

function authHeaders(endpoint: DocsEndpoint): Record<string, string> {
  if (endpoint === 'messages') return { 'x-api-key': 'API_KEY', 'anthropic-version': '2023-06-01' }
  if (endpoint === 'gemini') return { 'x-goog-api-key': 'API_KEY' }
  return { Authorization: 'Bearer API_KEY' }
}

function jsonPayload(endpoint: DocsEndpoint, stream: boolean): Record<string, unknown> | null {
  switch (endpoint) {
    case 'chat':
      return {
        model: 'gpt-5.6-sol',
        messages: [{ role: 'user', content: 'Hello!' }],
        temperature: 0.7,
        stream
      }
    case 'responses':
      return {
        model: 'gpt-5.6-sol',
        input: 'Hello!',
        instructions: 'You are a concise assistant.',
        stream
      }
    case 'messages':
      return {
        model: 'claude-sonnet-4-5',
        max_tokens: 1024,
        system: 'You are a concise assistant.',
        messages: [{ role: 'user', content: 'Hello!' }],
        stream
      }
    case 'embeddings':
      return {
        model: 'text-embedding-3-small',
        input: 'The quick brown fox jumps over the lazy dog.',
        encoding_format: 'float'
      }
    case 'imagesGenerations':
      return {
        model: 'gpt-image-2',
        prompt: 'A cat sitting by a window, soft natural light, minimal background',
        size: '1024x1024',
        quality: 'auto',
        n: 1,
        response_format: 'b64_json'
      }
    case 'gemini':
      return {
        contents: [{ role: 'user', parts: [{ text: 'Hello!' }] }],
        generationConfig: { temperature: 0.7, maxOutputTokens: 1024 }
      }
    default:
      return null
  }
}

function renderCurl(
  endpoint: DocsEndpoint,
  url: string,
  headers: Record<string, string>,
  payload: Record<string, unknown> | null,
  stream: boolean
): string {
  const lines = [`curl${stream ? ' -N' : ''} ${JSON.stringify(url)}`]
  for (const [key, value] of Object.entries(headers)) {
    lines.push(`  -H "${key}: ${value.replace('API_KEY', '$API_KEY')}"`)
  }
  if (endpoint === 'imagesEdits') {
    for (const field of [
      ['model', 'gpt-image-2'],
      ['prompt', 'Edit this image, keep the subject and improve the lighting'],
      ['size', '1024x1024'],
      ['n', '1']
    ]) {
      lines.push(`  -F "${field[0]}=${field[1]}"`)
    }
    lines.push('  -F "image=@input.png"')
    return `# Set API_KEY to your SuperAI API key before running.\n${lines.join(' \\\n')}`
  }
  if (payload) lines.push(`  -d '${JSON.stringify(payload, null, 2)}'`)
  return `# Set API_KEY to your SuperAI API key before running.\n${lines.join(' \\\n')}`
}

function renderPython(
  endpoint: DocsEndpoint,
  url: string,
  headers: Record<string, string>,
  payload: Record<string, unknown> | null,
  stream: boolean
): string {
  const pythonHeaders = JSON.stringify(headers, null, 4)
    .replace('"API_KEY"', 'os.environ["API_KEY"]')
    .replace('"Bearer API_KEY"', '"Bearer " + os.environ["API_KEY"]')
  if (endpoint === 'imagesEdits') {
    return `import os
import requests

headers = ${pythonHeaders}
files = {"image": open("input.png", "rb")}
data = {
    "model": "gpt-image-2",
    "prompt": "Edit this image, keep the subject and improve the lighting",
    "size": "1024x1024",
    "n": "1",
}
response = requests.post(${JSON.stringify(url)}, headers=headers, data=data, files=files)
response.raise_for_status()
print(response.json())`
  }
  return `import json
import os
import urllib.request

headers = ${pythonHeaders}
${payload ? `payload = json.loads(${JSON.stringify(JSON.stringify(payload))})\n` : ''}request = urllib.request.Request(
    ${JSON.stringify(url)},
${payload ? '    data=json.dumps(payload).encode("utf-8"),\n' : ''}    headers=headers,
    method="${docsEndpoints[endpoint].method}",
)
with urllib.request.urlopen(request) as response:
${stream ? '    for line in response:\n        print(line.decode("utf-8").rstrip())' : '    print(json.loads(response.read().decode("utf-8")))'}`
}

function renderJavaScript(
  endpoint: DocsEndpoint,
  url: string,
  headers: Record<string, string>,
  payload: Record<string, unknown> | null,
  stream: boolean
): string {
  if (endpoint === 'imagesEdits') {
    return `import { readFile } from "node:fs/promises";

const form = new FormData();
form.append("model", "gpt-image-2");
form.append("prompt", "Edit this image, keep the subject and improve the lighting");
form.append("size", "1024x1024");
form.append("n", "1");
form.append("image", new Blob([await readFile("input.png")], { type: "image/png" }), "input.png");

const response = await fetch(${JSON.stringify(url)}, {
  method: "POST",
  headers: { Authorization: \`Bearer \${process.env.API_KEY}\` },
  body: form,
});
if (!response.ok) throw new Error(await response.text());
console.log(await response.json());`
  }
  const jsHeaders = JSON.stringify(headers, null, 4)
    .replace('"Bearer API_KEY"', '`Bearer ${process.env.API_KEY}`')
    .replace('"API_KEY"', 'process.env.API_KEY')
  return `// Node.js 18+; set the API_KEY environment variable first.
const response = await fetch(${JSON.stringify(url)}, {
  method: "${docsEndpoints[endpoint].method}",
  headers: ${jsHeaders},${payload ? `\n  body: JSON.stringify(${JSON.stringify(payload, null, 2)}),` : ''}
});
if (!response.ok) throw new Error(await response.text());
${stream ? 'const decoder = new TextDecoder();\nfor await (const chunk of response.body) {\n  process.stdout.write(decoder.decode(chunk, { stream: true }));\n}' : 'console.log(await response.json());'}`
}

export function buildDocsExample(
  baseUrl: string,
  endpoint: DocsEndpoint,
  language: DocsLanguage,
  stream: boolean
): string {
  const url = resolveDocsRequestUrl(baseUrl, endpoint)
  const useStream = ['chat', 'responses', 'messages', 'gemini'].includes(endpoint) && stream
  const headers = authHeaders(endpoint)
  const payload = jsonPayload(endpoint, useStream)
  if (endpoint !== 'models') headers['Content-Type'] = endpoint === 'imagesEdits' ? 'multipart/form-data' : 'application/json'
  if (endpoint === 'imagesEdits') delete headers['Content-Type']

  if (language === 'curl') return renderCurl(endpoint, url, headers, payload, useStream)
  if (language === 'python') return renderPython(endpoint, url, headers, payload, useStream)
  return renderJavaScript(endpoint, url, headers, payload, useStream)
}
