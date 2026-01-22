import { Ai } from '@cloudflare/ai';

const DEFAULT_MODEL = '@cf/meta/llama-3-8b-instruct';

export default {
  async fetch(request, env) {
    if (request.method !== 'POST') {
      return jsonResponse({ success: false, error: 'Method not allowed' }, 405);
    }

    if (env.API_TOKEN) {
      const authHeader = request.headers.get('authorization') || '';
      const token = authHeader.startsWith('Bearer ') ? authHeader.slice(7) : '';
      if (token !== env.API_TOKEN) {
        return jsonResponse({ success: false, error: 'Unauthorized' }, 401);
      }
    }

    let payload;
    try {
      payload = await request.json();
    } catch (err) {
      return jsonResponse({ success: false, error: 'Invalid JSON' }, 400);
    }

    const html = typeof payload.html === 'string' ? payload.html : '';
    const elementInfo = typeof payload.elementInfo === 'string' ? payload.elementInfo : '';
    if (!html.trim() && !elementInfo.trim()) {
      return jsonResponse({ success: false, error: 'Missing html or elementInfo' }, 400);
    }

    const model = typeof payload.model === 'string' && payload.model
      ? payload.model
      : (env.AI_MODEL || DEFAULT_MODEL);

    const ai = new Ai(env.AI);
    const prompt = buildPrompt(html, elementInfo);
    const messages = [
      {
        role: 'system',
        content: [
          'You are an expert component architect.',
          'Analyze HTML elements and decide if they should become reusable components.',
          'Return JSON with fields: shouldBeComponent, reason, componentName, props, pattern, confidence.'
        ].join(' ')
      },
      {
        role: 'user',
        content: prompt
      }
    ];

    let responseText = '';
    try {
      const result = await ai.run(model, { messages });
      responseText = typeof result.response === 'string' ? result.response : JSON.stringify(result);
    } catch (err) {
      return jsonResponse({ success: false, error: 'AI request failed' }, 502);
    }

    const parsed = extractJson(responseText);
    if (!parsed) {
      return jsonResponse({ success: false, error: 'Unable to parse AI response', raw: responseText }, 502);
    }

    return jsonResponse({ success: true, result: parsed, raw: responseText });
  }
};

function buildPrompt(html, elementInfo) {
  const maxLength = 2000;
  const truncated = html.length > maxLength ? html.slice(0, maxLength) + '... [truncated]' : html;
  return [
    'Analyze this HTML element and determine if it should become a reusable component.',
    '',
    'Element Information:',
    elementInfo,
    '',
    'HTML Content:',
    truncated,
    '',
    'Return JSON only.'
  ].join('\n');
}

function extractJson(text) {
  let t = (text || '').trim();
  if (t.startsWith('```json')) {
    t = t.slice(7);
  } else if (t.startsWith('```')) {
    t = t.slice(3);
  }
  if (t.endsWith('```')) {
    t = t.slice(0, -3);
  }

  const start = t.indexOf('{');
  const end = t.lastIndexOf('}');
  if (start === -1 || end === -1 || start >= end) {
    return null;
  }

  const jsonText = t.slice(start, end + 1);
  try {
    return JSON.parse(jsonText);
  } catch (err) {
    return null;
  }
}

function jsonResponse(payload, status = 200) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: { 'content-type': 'application/json' }
  });
}
