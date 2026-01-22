# Component Analyzer Worker

This worker exposes an HTTP endpoint that runs component analysis using the
Cloudflare Workers AI SDK. It accepts JSON payloads and returns a structured
component suggestion result.

## Deploy

1. Install Wrangler (once):
   ```bash
   npm install -g wrangler
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Deploy the worker:
   ```bash
   wrangler deploy
   ```

4. (Optional) Set an auth token:
   ```bash
   wrangler secret put API_TOKEN
   ```

## Request Format

```json
{
  "html": "<div>...</div>",
  "elementInfo": "Tag: div\nCount: 3\nAttributes: class",
  "model": "@cf/meta/llama-3-8b-instruct"
}
```

## Environment Variables

- `API_TOKEN` (optional): Bearer token required in the `Authorization` header.
- `AI_MODEL` (optional): Default model if `model` is not provided.
