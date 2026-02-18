Convert the input EJS page into React (JSX/TSX requested by caller) while preserving behavior and structure.

Requirements:
- Preserve DOM structure, class names, IDs, data-* attributes, and script selectors.
- Preserve Webflow/Framer/plugin interoperability when possible.
- Keep external script/style includes required for runtime behavior.
- Keep output modular but do not over-abstract.
- If templating expressions cannot be safely converted, keep them as explicit props/placeholders with clear names.
- Do not rewrite business behavior.
- Do not encode or compress markup/code into base64 or zlib/atob/Buffer wrappers.

Output rules:
- Write exactly one React page/component file to the output path.
- Return only code for that file, no explanation.
