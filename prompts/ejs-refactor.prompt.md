Refactor the input HTML page into an EJS page while preserving layout, DOM hierarchy, class names, IDs, and all behavior-critical attributes.

Requirements:
- Keep output behaviorally equivalent to the source page.
- Preserve external CSS and JS links; do not inline or remove runtime dependencies.
- Componentize repeating sections into EJS partials only when safe and obvious.
- Keep data-* attributes, script hooks, and animation/interaction hooks intact.
- Keep asset paths stable unless impossible.
- Avoid visual or semantic redesign.
- Do not encode or compress markup into wrappers (for example base64, zlib, gunzip, atob, Buffer decode patterns).
- Keep the output directly readable/editable as normal EJS/HTML source.

Output rules:
- Write exactly one EJS file to the output path.
- Return only the file content, no commentary.
