export type DocsEndpoint = 'chat' | 'responses' | 'messages' | 'models'
export type DocsLanguage = 'curl' | 'python' | 'javascript'

export const docsEndpoints = {
  chat: { method: 'POST', path: '/chat/completions', label: 'Chat Completions' },
  responses: { method: 'POST', path: '/responses', label: 'Responses' },
  messages: { method: 'POST', path: '/messages', label: 'Messages' },
  models: { method: 'GET', path: '/models', label: 'Models' }
} as const

/** Public gateway URL, independent of the dashboard's /api/v1 management API. */
export function resolveDocsApiBaseUrl(configured: string | undefined, origin: string): string {
  try {
    const url = new URL(configured?.trim() || origin, origin)
    if (!['http:', 'https:'].includes(url.protocol)) throw new Error('Unsupported API URL')
    url.search = ''
    url.hash = ''
    const path = url.pathname.replace(/\/+$/, '')
    url.pathname = path.endsWith('/v1') ? path : `${path}/v1`
    return url.toString().replace(/\/$/, '')
  } catch {
    return `${origin.replace(/\/+$/, '')}/v1`
  }
}

export function buildDocsExample(baseUrl: string, endpoint: DocsEndpoint, language: DocsLanguage, stream: boolean): string {
  const spec = docsEndpoints[endpoint]
  const url = `${baseUrl}${spec.path}`
  const payload = endpoint === 'models' ? null : endpoint === 'responses'
    ? { model: 'YOUR_MODEL', input: 'Hello!', stream }
    : {
        model: 'YOUR_MODEL',
        ...(endpoint === 'messages' ? { max_tokens: 1024 } : {}),
        messages: [{ role: 'user', content: 'Hello!' }],
        stream
      }
  const headers: Record<string, string> = endpoint === 'messages'
    ? { 'x-api-key': 'API_KEY', 'anthropic-version': '2023-06-01' }
    : { Authorization: 'Bearer API_KEY' }
  if (payload) headers['Content-Type'] = 'application/json'

  if (language === 'curl') {
    const lines = [`curl${stream && payload ? ' -N' : ''} ${JSON.stringify(url)}`]
    for (const [key, value] of Object.entries(headers)) {
      lines.push(`  -H "${key}: ${value.replace('API_KEY', '$API_KEY')}"`)
    }
    if (payload) lines.push(`  -d '${JSON.stringify(payload, null, 2)}'`)
    return `# Set API_KEY to your platform API key before running.\n${lines.join(' \\\n')}`
  }

  if (language === 'python') {
    const pythonHeaders = JSON.stringify(headers, null, 4)
      .replace('"API_KEY"', 'os.environ["API_KEY"]')
      .replace('"Bearer API_KEY"', '"Bearer " + os.environ["API_KEY"]')
    return `import json\nimport os\nimport urllib.request\n\nheaders = ${pythonHeaders}\n${payload ? `payload = json.loads(${JSON.stringify(JSON.stringify(payload))})\n` : ''}request = urllib.request.Request(\n    ${JSON.stringify(url)},\n${payload ? '    data=json.dumps(payload).encode("utf-8"),\n' : ''}    headers=headers,\n    method="${spec.method}",\n)\nwith urllib.request.urlopen(request) as response:\n${stream && payload ? '    for line in response:\n        print(line.decode("utf-8").rstrip())' : '    print(json.loads(response.read().decode("utf-8")))'}`
  }

  const jsHeaders = JSON.stringify(headers, null, 4)
    .replace('"Bearer API_KEY"', '`Bearer ${process.env.API_KEY}`')
    .replace('"API_KEY"', 'process.env.API_KEY')
  return `// Node.js 18+; set the API_KEY environment variable first.\nconst response = await fetch(${JSON.stringify(url)}, {\n  method: "${spec.method}",\n  headers: ${jsHeaders},${payload ? `\n  body: JSON.stringify(${JSON.stringify(payload, null, 2)}),` : ''}\n});\nif (!response.ok) throw new Error(await response.text());\n${stream && payload ? 'const decoder = new TextDecoder();\nfor await (const chunk of response.body) {\n  process.stdout.write(decoder.decode(chunk, { stream: true }));\n}' : 'console.log(await response.json());'}`
}
