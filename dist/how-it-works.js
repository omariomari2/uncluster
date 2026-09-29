(() => {
  'use strict';

  const root = document.documentElement;
  const themeButton = document.getElementById('theme-toggle');
  const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)');

  function setTheme(theme) {
    const nextTheme = theme === 'dark' ? 'dark' : 'light';
    root.dataset.theme = nextTheme;
    themeButton.textContent = nextTheme === 'dark' ? 'Light mode' : 'Dark mode';
    themeButton.setAttribute('aria-pressed', String(nextTheme === 'dark'));

    try {
      localStorage.setItem('uncluster-how-it-works-theme', nextTheme);
    } catch (error) {
      // The selected theme still applies when storage is unavailable.
    }
  }

  setTheme(root.dataset.theme);
  themeButton.addEventListener('click', () => {
    setTheme(root.dataset.theme === 'dark' ? 'light' : 'dark');
  });

  if (!reducedMotion.matches && 'IntersectionObserver' in window) {
    const revealGroups = [
      document.querySelectorAll('.docs-hero > *'),
      document.querySelectorAll('.docs-index'),
      document.querySelectorAll('.docs-section-header'),
      document.querySelectorAll('.flow-diagram figcaption'),
      document.querySelectorAll('.flow-inspector'),
      document.querySelectorAll('.capability-row'),
      document.querySelectorAll('.pipeline-list > li'),
      document.querySelectorAll('.code-compare'),
      document.querySelectorAll('.security-list > article'),
      document.querySelectorAll('.api-table-wrap'),
      document.querySelectorAll('.system-note'),
      document.querySelectorAll('.site-footer > *')
    ];

    const observer = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        if (!entry.isIntersecting) return;
        entry.target.classList.add('is-visible');
        observer.unobserve(entry.target);
      });
    }, {
      rootMargin: '0px 0px -8% 0px',
      threshold: 0.08
    });

    revealGroups.forEach((group) => {
      group.forEach((element, index) => {
        element.classList.add('scroll-reveal');
        element.style.setProperty('--reveal-delay', `${Math.min(index, 4) * 55}ms`);
        observer.observe(element);
      });
    });

    requestAnimationFrame(() => {
      root.dataset.motion = 'ready';
    });
  }

  const canvas = document.getElementById('flow-canvas');
  const context = canvas?.getContext('2d');
  if (!canvas || !context) return;

  const shell = canvas.closest('.flow-canvas-shell');
  const zoomOutput = document.getElementById('flow-zoom');
  const viewSelect = document.getElementById('flow-view');
  const nodeSelect = document.getElementById('flow-node-select');
  const fullscreenButton = shell.querySelector('[data-flow-action="fullscreen"]');
  const inspectorKicker = document.getElementById('flow-inspector-kicker');
  const inspectorTitle = document.getElementById('flow-inspector-title');
  const inspectorFile = document.getElementById('flow-inspector-file');
  const inspectorRole = document.getElementById('flow-inspector-role');
  const inspectorContract = document.getElementById('flow-inspector-contract');
  const inspectorInvariants = document.getElementById('flow-inspector-invariants');

  const nodeCatalog = {
    api: {
      group: 'Surfaces',
      label: 'HTTP SURFACE',
      title: 'Fiber API',
      detail: 'JSON, multipart, and ZIP responses',
      meta: 'main.go',
      kind: 'surface',
      order: 0,
      file: 'main.go - setupRoutes and handlers',
      role: 'Owns the browser-facing routes, request parsing, response envelopes, static UI, and downloadable ZIP headers.',
      contract: 'POST /api/export*, /api/scrape*, and /api/bundle-zip dispatch into capture and generation adapters. /api/format and /api/analyze return JSON instead of archives.',
      invariants: [
        'Fiber rejects request bodies above 50 MiB and handlers reject missing source fields with HTTP 400.',
        'ZIP responses set content type, length, and a sanitized attachment filename.',
        'Capture, conversion, and packaging failures are returned as JSON HTTP 500 responses.'
      ]
    },
    cli: {
      group: 'Surfaces',
      label: 'LOCAL SURFACE',
      title: 'CLI dispatch',
      detail: 'format, analyze, split, nodejs, ejs, bundle',
      meta: 'cmd/uncluster',
      kind: 'surface',
      order: 0,
      file: 'cmd/uncluster/main.go',
      role: 'Dispatches local files into formatting, analysis, extraction, project generation, or bundle workflows and writes results to disk.',
      contract: 'One input path plus a -to mode. Non-bundle modes read the complete file; bundle mode delegates HTML or ZIP detection to bundle.ProcessWithOptions.',
      invariants: [
        'There is no CLI URL-scrape mode and no CLI input-size limit.',
        'Usage errors exit 2; processing failures exit 1; help exits 0.',
        'CLI project modes treat local-asset localization errors as nonfatal and continue generation.'
      ]
    },
    request: {
      group: 'Intake paths',
      label: 'REQUEST GATE',
      title: 'Validate shape',
      detail: 'Required JSON fields or a .zip upload',
      meta: '50 MiB HTTP limit',
      kind: 'guard',
      order: 1,
      file: 'main.go - handleExport*, handleScrape*, handleBundleZip',
      role: 'Separates raw HTML, public URL, and uploaded ZIP requests before any transformation work starts.',
      contract: 'A nonblank html or url field, or a multipart file whose name ends in .zip, is passed to the matching adapter.',
      invariants: [
        'Malformed JSON, blank values, missing file fields, and non-.zip uploads stop with HTTP 400.',
        'The CLI bypasses this HTTP-only gate and validates its own flags and paths.',
        'A valid request shape does not imply the target URL or archive contents are safe; later guards own those checks.'
      ]
    },
    direct: {
      group: 'Intake paths',
      label: 'HTML ADAPTER',
      title: 'extractor.Extract',
      detail: 'Parse, split inline code, fetch links, rewrite',
      meta: 'string -> ExtractedContent',
      kind: 'adapter',
      order: 2,
      file: 'internal/extractor/extractor.go',
      role: 'Transforms a supplied HTML string into formatted rewritten HTML plus typed CSS, JavaScript, local-asset, and capture metadata collections.',
      contract: 'HTML string in; *extractor.ExtractedContent out. It parses, runs InlineCollector, fetches supported absolute CSS and JS, rewrites successful links, renders, and formats.',
      invariants: [
        'This narrow path only downloads literal absolute HTTP(S) stylesheets and script sources; relative URLs, images, and srcset are not localized here.',
        'Individual remote fetch failures are retained externally and recorded as failed; initial parse, render, or formatting failure is fatal.',
        'Google Fonts CSS and iframe sources intentionally remain external.'
      ]
    },
    url: {
      group: 'Intake paths',
      label: 'URL ADAPTER',
      title: 'scraper.ScrapeURL',
      detail: 'Fetch page and capture its render assets',
      meta: 'URL -> ExtractedContent',
      kind: 'adapter',
      order: 2,
      file: 'internal/scraper/scraper.go',
      role: 'Orchestrates a public webpage capture: guarded page fetch, DOM parse, asset graph discovery, recursive CSS capture, rewriting, inline extraction, and formatting.',
      contract: 'Public HTTP(S) URL in; *extractor.ExtractedContent out with rewritten HTML, text resources, binaries, and a source-aware capture manifest.',
      invariants: [
        'Page-fetch failure is fatal; individual CSS, JS, image, font, or media failures are best effort and remain external.',
        'Relative assets resolve against the submitted URL, not a redirect final URL or an HTML base element.',
        'Webflow pages receive local-host compatibility scripts around their runtime without hardcoding a site identifier.'
      ]
    },
    network: {
      group: 'Intake paths',
      label: 'NETWORK GUARD',
      title: 'safehttp.Client',
      detail: 'Public HTTP(S), guarded dial, bounded body',
      meta: '25 MiB / response',
      kind: 'guard',
      order: 2,
      file: 'internal/safehttp/safehttp.go and internal/fetcher/fetcher.go',
      role: 'Protects every remote page and asset fetch against unsafe schemes, non-public destinations, redirect abuse, oversized responses, and non-success status codes.',
      contract: 'Validated HTTP(S) request in; bounded 2xx response bytes and MIME type out, or an error before content enters the pipeline.',
      invariants: [
        'Resolved addresses are checked at connection time; loopback, private, link-local, multicast, reserved, and documentation ranges are blocked.',
        'Redirects are revalidated and stop after 10 hops; each response body is capped at 25 MiB.',
        'CSS and JS fetches time out after 10 seconds; page and binary fetches use 30 seconds.'
      ]
    },
    bundle: {
      group: 'Intake paths',
      label: 'BUNDLE ADAPTER',
      title: 'Bundle process',
      detail: 'bundle.ProcessWithOptions: load HTML or ZIP and materialize a workspace',
      meta: 'path -> Result + files',
      kind: 'adapter',
      order: 2,
      file: 'internal/bundle/bundle.go',
      role: 'Loads a local HTML file or ZIP site, localizes referenced files, invokes extraction, and writes an original page, split tree, and EJS project.',
      contract: 'Input path and output options in; Result paths plus <site>/index.html, unzip/, and ejs/ on disk. Bundle mode does not generate TSX.',
      invariants: [
        'The top-level index.html is the untouched selected source; rewritten content lives under unzip/ and ejs/.',
        'Only referenced existing local files are copied, with recursive CSS dependency discovery confined to the source root.',
        'Missing local references are silently retained and are not represented as capture-manifest failures.'
      ]
    },
    archive: {
      group: 'Intake paths',
      label: 'ARCHIVE GUARD',
      title: 'Safe ZIP intake',
      detail: 'Contain paths, cap expansion, select index',
      meta: '512 MiB expanded cap',
      kind: 'guard',
      order: 2,
      file: 'internal/bundle/bundle.go - extractZip and selectIndexHTML',
      role: 'Extracts ZIP entries into a temporary root, rejects traversal, bounds expanded bytes, and deterministically chooses a usable index.html.',
      contract: 'ZIP path in; temporary source root plus selected HTML document out. Any corrupt, unsafe, oversized, or unusable archive aborts the bundle.',
      invariants: [
        'Every resolved entry must stay under the temporary extraction root and total expanded content may not exceed 512 MiB.',
        'Only files named index.html with HTML-like content are candidates.',
        'Selection prefers a parent matching the ZIP stem, then larger content, then lexical path order.'
      ]
    },
    parse: {
      group: 'Capture modules',
      label: 'DOCUMENT MODEL',
      title: 'Parse and render DOM',
      detail: 'x/net/html plus deterministic formatter',
      meta: 'adapter-local DOM',
      kind: 'module',
      order: 3,
      file: 'internal/extractor, internal/scraper, internal/formatter',
      role: 'Provides tolerant HTML parsing, DOM mutation, browser-oriented rendering, and readable formatting to each source adapter.',
      contract: 'HTML text becomes an adapter-local DOM, then returns to text after mutations. The formatter reparses rendered HTML before serializing it.',
      invariants: [
        'There is no retained global DOM shared across routes; bundle mode parses, renders, and reparses through Extract.',
        'Child and attribute order is preserved as stored; script, style, pre, and textarea text remains raw.',
        'Scrape formatting failure falls back to rendered HTML, while direct extractor formatting failure aborts.'
      ]
    },
    inline: {
      group: 'Capture modules',
      label: 'SHARED HELPER',
      title: 'InlineCollector',
      detail: 'Extract executable style and script blocks in place',
      meta: 'internal/htmlutil',
      kind: 'module',
      order: 3,
      file: 'internal/htmlutil/inline.go',
      role: 'Moves nonempty inline style blocks and executable inline scripts into portable files while replacing each node at its original DOM position.',
      contract: 'Mutable DOM in; inline/style-N.css and inline/script-N.js records plus aggregate CSS and JS builders out.',
      invariants: [
        'CSS and JS counters are independent; each same-type slice follows depth-first document order.',
        'Nonce, defer, type, data attributes, and other non-resource attributes survive replacement.',
        'Empty blocks, external scripts, and non-executable data scripts such as JSON-LD stay in place.'
      ]
    },
    discovery: {
      group: 'Capture modules',
      label: 'RESOURCE GRAPH',
      title: 'Discover dependencies',
      detail: 'CSS, JS, srcset, media, icons, fonts, data-src',
      meta: 'first-seen order',
      kind: 'module',
      order: 3,
      file: 'internal/scraper/scraper.go - findAllAssetURLs and fetchCSSResources',
      role: 'Walks URL captures for render/runtime references and expands CSS imports and url() dependencies before downloads and path rewriting.',
      contract: 'Parsed page plus base URL in; typed CSS, JavaScript, binary, and retained-external URL lists out.',
      invariants: [
        'Within each asset kind, first-seen DOM order is preserved and duplicates are removed.',
        'CSS roots keep HTML order; imports are traversed depth-first in source order, once, with cycle protection.',
        'The srcset and CSS scanners are intentionally lightweight rather than standards-complete parsers.'
      ]
    },
    localize: {
      group: 'Capture modules',
      label: 'RESOURCE I/O',
      title: 'Fetch and localize',
      detail: 'Stable filenames, MIME data, and local paths',
      meta: 'fetcher + LocalAsset',
      kind: 'module',
      order: 3,
      file: 'internal/fetcher/fetcher.go and internal/scraper/scraper.go',
      role: 'Downloads text and binary dependencies, assigns collision-safe descriptive filenames, and builds the absolute-URL to local-path map.',
      contract: 'Ordered URL lists in; ordered FetchedResource slices, LocalAsset binaries, and rewrite mappings out.',
      invariants: [
        'FetchExternalResources is sequential and returns one slot per input, including failures, so within-kind order survives.',
        'Failed assets are omitted locally but keep their original remote references and become failed manifest entries.',
        'Imported CSS is emitted but marked separately so TSX does not import it twice.'
      ]
    },
    rewrite: {
      group: 'Capture modules',
      label: 'NORMALIZATION',
      title: 'Rewrite and rebase',
      detail: 'HTML URLs, srcset descriptors, and CSS references',
      meta: 'successful assets only',
      kind: 'module',
      order: 4,
      file: 'internal/scraper/scraper.go and internal/extractor/extractor.go',
      role: 'Replaces successfully localized source references with portable relative paths and leaves unresolved or intentionally external references intact.',
      contract: 'DOM, localized URL map, and CSS text in; renderable rewritten HTML and rebased stylesheets out.',
      invariants: [
        'srcset descriptors survive; URL-bearing data-src values are rewritten while ordinary metadata is ignored.',
        'CSS references are rebased from each localized stylesheet path; failed targets remain unchanged.',
        'Document order remains the only cross-resource ordering record; resources are not collapsed into one global sequence.'
      ]
    },
    contract: {
      group: 'Capture modules',
      label: 'EXCHANGE CONTRACT',
      title: 'ExtractedContent',
      detail: 'Rewritten HTML plus typed resource collections',
      meta: 'shared data, not a shared DOM',
      kind: 'contract',
      order: 5,
      file: 'internal/extractor/extractor.go - ExtractedContent',
      role: 'Is the stable seam between source-specific capture adapters and target-specific packaging or project generation.',
      contract: 'HTML, aggregate CSS/JS, InlineCSS, InlineJS, ExternalCSS, ExternalJS, LocalAssets, and Capture are carried together to output adapters.',
      invariants: [
        'The contract contains separate typed slices; there is no single globally ordered resource list.',
        'Within-kind ordering is meaningful, but archive entry order is not a runtime ordering contract.',
        'Target rewrites reparse the HTML; failure there returns the existing extracted HTML unchanged.'
      ]
    },
    manifest: {
      group: 'Capture modules',
      label: 'EVIDENCE',
      title: 'CaptureManifest v1',
      detail: 'localized, retained-external, or failed',
      meta: 'uncluster-capture.json',
      kind: 'evidence',
      order: 5,
      file: 'internal/extractor/capture.go',
      role: 'Records what happened to each attempted remote render/runtime dependency and whether every attempted localization succeeded.',
      contract: 'source_url, complete, and ordered asset records containing URL, optional path, type, status, and optional error.',
      invariants: [
        'Only failed status flips complete to false; retained external assets do not.',
        'complete=true does not mean fully offline or free of external runtime dependencies.',
        'API exports include this file; CLI split and bundle flows use split-manifest.json instead.'
      ]
    },
    standalone: {
      group: 'Target generators',
      label: 'HTML TARGET',
      title: 'Standalone archive',
      detail: 'index.html plus captured resources and evidence',
      meta: 'zipper.CreateZipWithMetadata',
      kind: 'output',
      order: 6,
      file: 'internal/zipper/zipper.go',
      role: 'Packages the rewritten document without translating its markup or moving its executable scripts into a framework runtime.',
      contract: 'ExtractedContent in; index.html, inline/, external/, assets/, and optional uncluster-capture.json in a ZIP or split directory.',
      invariants: [
        'Script elements remain at document positions with original attributes; duplicate script tags remain duplicate executions.',
        'Only successful and nonempty fetched resources are written; retained remote references continue to resolve at runtime.',
        'Any ZIP entry creation or write failure aborts standalone archive creation.'
      ]
    },
    target: {
      group: 'Target generators',
      label: 'PATH ADAPTER',
      title: 'Target rewrites',
      detail: 'RewriteForEJS and RewriteForNodeJS',
      meta: 'public-root contracts',
      kind: 'adapter',
      order: 6,
      file: 'internal/extractor/extractor.go',
      role: 'Translates portable capture paths into the public-root layout expected by Express/EJS or Vite/React projects.',
      contract: 'Extracted HTML in; equivalent HTML whose inline, external, style, and script references match the selected generated project tree.',
      invariants: [
        'EJS roots localized assets so nested extensionless routes do not request nested asset paths.',
        'TSX maps styles under /styles and scripts under /scripts before shell generation.',
        'If reparsing or rendering fails, the original extracted HTML is returned instead of aborting.'
      ]
    },
    decompose: {
      group: 'Target generators',
      label: 'SHARED SEAM',
      title: 'Component boundaries',
      detail: 'Root expansion, candidates, naming, and uniqueness',
      meta: 'deterministic heuristics',
      kind: 'module',
      order: 6,
      file: 'internal/nodejs/ejs_builder.go and tsx_builder.go',
      role: 'Chooses existing DOM subtrees as editable files without inventing content, data models, or layout wrappers.',
      contract: 'Body DOM in; named candidate subtrees out. Names prefer id, descriptive class, semantic tag, then block-N with collision suffixes.',
      invariants: [
        'Layout containers may expand up to two levels; surrounding and nested markup must remain in the main document/component.',
        'EJS scans semantic boundaries deeper and extracts only partials at least 500 bytes and 15 lines.',
        'TSX converts every selected subtree; no LLM or prop inference participates.'
      ]
    },
    ejs: {
      group: 'Target generators',
      label: 'EJS TARGET',
      title: 'GenerateEJSProject',
      detail: 'Full document, partials, public assets, Express',
      meta: 'views/ + public/',
      kind: 'output',
      order: 7,
      file: 'internal/nodejs/ejs_builder.go and ejs_templates.go',
      role: 'Preserves the complete HTML document, replaces qualifying subtrees with EJS includes, and emits a runnable Express project.',
      contract: 'Rewritten HTML and captured resources in; package.json, server.js, views/index.ejs, optional views/partials, and public resources out.',
      invariants: [
        'Executable scripts keep normal HTML execution timing and document/partial order; there is no hydration delay or deduplication.',
        'Express serves public/, renders index.ejs for extensionless routes, and returns plain-text 404 for missing extension-bearing assets.',
        'Runtime port defaults to 8080 and may be overridden by PORT.'
      ]
    },
    jsx: {
      group: 'Target generators',
      label: 'REACT TRANSLATION',
      title: 'TSX plus hydration HTML',
      detail: 'React attributes, controls, whitespace, escaping',
      meta: 'internal/converter',
      kind: 'module',
      order: 7,
      file: 'internal/converter/jsx.go',
      role: 'Converts selected HTML fragments into React-valid TSX while separately generating static body markup that matches React first render.',
      contract: 'HTML fragment in; deterministic TSX component or hydration-safe HTML out.',
      invariants: [
        'Node, attribute, text, order, significant whitespace, void elements, SVG, and form-control behavior are translated by fixed rules.',
        'Source event attributes become named handler stubs in TSX; no behavior is inferred.',
        'Non-executable script nodes are skipped, so data scripts preserved by HTML/EJS do not appear in TSX output.'
      ]
    },
    tsx: {
      group: 'Target generators',
      label: 'REACT TARGET',
      title: 'GenerateProject',
      detail: 'Vite shell, real components, hydration, script replay',
      meta: 'src/ + public/',
      kind: 'output',
      order: 8,
      file: 'internal/nodejs/builder.go, tsx_builder.go, and templates.go',
      role: 'Emits a runnable React/Vite/TypeScript project whose pre-rendered body and first React render match without a #root wrapper.',
      contract: 'Target-rewritten HTML and captured resources in; Vite config, Express server, src shell/components/styles, public scripts/assets, and project metadata out.',
      invariants: [
        'Source html/body attributes, title, metadata, canonical and icon links survive; stylesheets and scripts are emitted once through the generated runtime.',
        'After hydration commits, executable scripts are canonicalized, deduplicated, and appended sequentially in source DOM order with original attributes and parent placement.',
        'A replayed script load failure rejects the loader and prevents later scripts; inline CSS is imported before external CSS, so arbitrary stylesheet interleaving is not preserved.'
      ]
    }
  };

  const nodeOrder = [
    'api', 'cli',
    'request', 'direct', 'url', 'network', 'bundle', 'archive',
    'parse', 'inline', 'discovery', 'localize', 'rewrite', 'contract', 'manifest',
    'standalone', 'target', 'decompose', 'ejs', 'jsx', 'tsx'
  ];

  const edges = [
    { from: 'api', to: 'request', label: 'JSON / multipart' },
    { from: 'cli', to: 'direct', label: 'file modes' },
    { from: 'cli', to: 'bundle', label: 'bundle mode' },
    { from: 'request', to: 'direct', label: '/export*' },
    { from: 'request', to: 'url', label: '/scrape*' },
    { from: 'request', to: 'bundle', label: '/bundle-zip' },
    { from: 'url', to: 'network', label: 'guarded fetch' },
    { from: 'direct', to: 'parse', label: 'parse HTML' },
    { from: 'network', to: 'parse', label: 'page HTML' },
    { from: 'network', to: 'localize', label: 'asset fetches', style: 'conditional' },
    { from: 'bundle', to: 'archive', label: 'ZIP branch' },
    { from: 'archive', to: 'parse', label: 'rewrite + Extract', style: 'conditional' },
    { from: 'parse', to: 'inline', label: 'DOM walk' },
    { from: 'parse', to: 'discovery', label: 'references' },
    { from: 'discovery', to: 'localize', label: 'typed URLs' },
    { from: 'inline', to: 'rewrite', label: 'inline/*' },
    { from: 'localize', to: 'rewrite', label: 'URL map' },
    { from: 'rewrite', to: 'contract', label: 'formatted HTML' },
    { from: 'contract', to: 'manifest', label: 'asset outcomes' },
    { from: 'contract', to: 'standalone', label: 'package as-is' },
    { from: 'contract', to: 'target', label: 'project target' },
    { from: 'target', to: 'decompose', label: 'rewritten DOM' },
    { from: 'decompose', to: 'ejs', label: 'HTML partials' },
    { from: 'decompose', to: 'jsx', label: 'DOM fragments' },
    { from: 'jsx', to: 'tsx', label: 'shell + runtime' }
  ];

  function createLayout(key, width, height, zones, positions, views) {
    return {
      key,
      width,
      height,
      zones,
      views,
      edges,
      nodes: nodeOrder.map((id) => ({ id, ...nodeCatalog[id], ...positions[id] }))
    };
  }

  const desktopLayout = createLayout(
    'desktop',
    1880,
    1050,
    [
      { x: 20, y: 30, width: 220, height: 990, label: 'SURFACES', detail: 'HTTP and local entry points' },
      { x: 260, y: 30, width: 420, height: 990, label: 'INTAKE PATHS', detail: 'Source-specific orchestration and guards' },
      { x: 700, y: 30, width: 700, height: 990, label: 'CAPTURE MODULES', detail: 'Shared helpers and exchange contract' },
      { x: 1420, y: 30, width: 440, height: 990, label: 'TARGET GENERATORS', detail: 'Output-specific rendering and runtime' }
    ],
    {
      api: { x: 42, y: 130, width: 176, height: 142 },
      cli: { x: 42, y: 450, width: 176, height: 142 },
      request: { x: 280, y: 85, width: 176, height: 146 },
      direct: { x: 480, y: 85, width: 180, height: 152 },
      url: { x: 280, y: 370, width: 176, height: 152 },
      network: { x: 480, y: 370, width: 180, height: 152 },
      bundle: { x: 280, y: 700, width: 176, height: 158 },
      archive: { x: 480, y: 700, width: 180, height: 158 },
      parse: { x: 720, y: 90, width: 192, height: 152 },
      inline: { x: 950, y: 90, width: 192, height: 152 },
      discovery: { x: 720, y: 370, width: 192, height: 158 },
      localize: { x: 950, y: 370, width: 192, height: 158 },
      rewrite: { x: 1180, y: 230, width: 192, height: 158 },
      contract: { x: 950, y: 680, width: 212, height: 178 },
      manifest: { x: 1190, y: 700, width: 182, height: 152 },
      standalone: { x: 1450, y: 105, width: 182, height: 160 },
      target: { x: 1450, y: 390, width: 182, height: 152 },
      decompose: { x: 1655, y: 390, width: 184, height: 152 },
      ejs: { x: 1655, y: 650, width: 184, height: 164 },
      jsx: { x: 1450, y: 820, width: 182, height: 164 },
      tsx: { x: 1655, y: 820, width: 184, height: 164 }
    },
    {
      all: { x: 0, y: 0, width: 1880, height: 1050 },
      intake: { x: 0, y: 10, width: 700, height: 1020 },
      capture: { x: 680, y: 10, width: 740, height: 1020 },
      outputs: { x: 1400, y: 10, width: 480, height: 1020 }
    }
  );

  const mobileLayout = createLayout(
    'mobile',
    700,
    3100,
    [
      { x: 20, y: 30, width: 660, height: 310, label: 'SURFACES', detail: 'HTTP and local entry points' },
      { x: 20, y: 365, width: 660, height: 900, label: 'INTAKE PATHS', detail: 'Source-specific orchestration and guards' },
      { x: 20, y: 1290, width: 660, height: 1010, label: 'CAPTURE MODULES', detail: 'Shared helpers and exchange contract' },
      { x: 20, y: 2325, width: 660, height: 745, label: 'TARGET GENERATORS', detail: 'Output-specific rendering and runtime' }
    ],
    {
      api: { x: 40, y: 105, width: 280, height: 158 },
      cli: { x: 380, y: 105, width: 280, height: 158 },
      request: { x: 40, y: 455, width: 280, height: 162 },
      direct: { x: 380, y: 455, width: 280, height: 162 },
      url: { x: 40, y: 700, width: 280, height: 162 },
      network: { x: 380, y: 700, width: 280, height: 162 },
      bundle: { x: 40, y: 945, width: 280, height: 166 },
      archive: { x: 380, y: 945, width: 280, height: 166 },
      parse: { x: 40, y: 1380, width: 280, height: 162 },
      inline: { x: 380, y: 1380, width: 280, height: 162 },
      discovery: { x: 40, y: 1620, width: 280, height: 166 },
      localize: { x: 380, y: 1620, width: 280, height: 166 },
      rewrite: { x: 40, y: 1860, width: 280, height: 166 },
      contract: { x: 380, y: 1860, width: 280, height: 180 },
      manifest: { x: 210, y: 2105, width: 280, height: 166 },
      standalone: { x: 40, y: 2410, width: 280, height: 166 },
      target: { x: 380, y: 2410, width: 280, height: 166 },
      decompose: { x: 40, y: 2650, width: 280, height: 166 },
      ejs: { x: 380, y: 2650, width: 280, height: 166 },
      jsx: { x: 40, y: 2890, width: 280, height: 166 },
      tsx: { x: 380, y: 2890, width: 280, height: 166 }
    },
    {
      all: { x: 0, y: 0, width: 700, height: 3100 },
      intake: { x: 0, y: 20, width: 700, height: 1260 },
      capture: { x: 0, y: 1280, width: 700, height: 1030 },
      outputs: { x: 0, y: 2315, width: 700, height: 770 }
    }
  );

  const state = {
    layout: desktopLayout,
    view: 'all',
    width: 0,
    height: 0,
    pixelRatio: 1,
    scale: 1,
    fitScale: 1,
    offsetX: 0,
    offsetY: 0,
    dragging: false,
    pointerId: null,
    pointerX: 0,
    pointerY: 0,
    dragDistance: 0,
    hoveredId: '',
    selectedId: 'contract',
    progress: reducedMotion.matches ? 1 : 0,
    played: reducedMotion.matches
  };

  function readPalette() {
    const styles = getComputedStyle(root);
    return {
      canvas: styles.getPropertyValue('--canvas').trim(),
      surface: styles.getPropertyValue('--surface').trim(),
      surfaceStrong: styles.getPropertyValue('--surface-strong').trim(),
      ink: styles.getPropertyValue('--ink').trim(),
      muted: styles.getPropertyValue('--muted').trim(),
      line: styles.getPropertyValue('--line').trim(),
      lineStrong: styles.getPropertyValue('--line-strong').trim(),
      inverse: styles.getPropertyValue('--inverse').trim(),
      inverseInk: styles.getPropertyValue('--inverse-ink').trim(),
      inverseMuted: styles.getPropertyValue('--inverse-muted').trim(),
      sans: getComputedStyle(document.body).fontFamily,
      mono: styles.getPropertyValue('--font-mono').trim(),
      shadow: root.dataset.theme === 'dark' ? 'rgba(0, 0, 0, 0.48)' : 'rgba(17, 17, 17, 0.12)'
    };
  }

  function clamp(value, minimum, maximum) {
    return Math.min(maximum, Math.max(minimum, value));
  }

  function eased(value) {
    const bounded = clamp(value, 0, 1);
    return 1 - Math.pow(1 - bounded, 3);
  }

  function roundedRect(x, y, width, height, radius) {
    const nextRadius = Math.min(radius, width / 2, height / 2);
    context.beginPath();
    context.moveTo(x + nextRadius, y);
    context.arcTo(x + width, y, x + width, y + height, nextRadius);
    context.arcTo(x + width, y + height, x, y + height, nextRadius);
    context.arcTo(x, y + height, x, y, nextRadius);
    context.arcTo(x, y, x + width, y, nextRadius);
    context.closePath();
  }

  function wrapText(text, x, y, maxWidth, lineHeight, maximumLines = 2) {
    const words = text.split(' ');
    const lines = [];
    let current = '';

    words.forEach((word) => {
      const candidate = current ? `${current} ${word}` : word;
      if (context.measureText(candidate).width <= maxWidth || !current) {
        current = candidate;
        return;
      }
      lines.push(current);
      current = word;
    });

    if (current) lines.push(current);
    const visible = lines.slice(0, maximumLines);
    if (lines.length > maximumLines && visible.length > 0) {
      let finalLine = visible[visible.length - 1];
      while (finalLine.length > 1 && context.measureText(`${finalLine}...`).width > maxWidth) {
        finalLine = finalLine.slice(0, -1);
      }
      visible[visible.length - 1] = `${finalLine.trim()}...`;
    }
    visible.forEach((line, index) => context.fillText(line, x, y + index * lineHeight));
  }

  function nodeCenter(node) {
    return { x: node.x + node.width / 2, y: node.y + node.height / 2 };
  }

  function anchorToward(node, targetNode) {
    const source = nodeCenter(node);
    const target = nodeCenter(targetNode);
    const dx = target.x - source.x;
    const dy = target.y - source.y;
    if (Math.abs(dx) >= Math.abs(dy)) {
      return dx >= 0
        ? { x: node.x + node.width, y: source.y, side: 'right' }
        : { x: node.x, y: source.y, side: 'left' };
    }
    return dy >= 0
      ? { x: source.x, y: node.y + node.height, side: 'bottom' }
      : { x: source.x, y: node.y, side: 'top' };
  }

  function drawArrow(point, angle, palette, alpha) {
    const size = 9;
    context.save();
    context.globalAlpha = alpha;
    context.translate(point.x, point.y);
    context.rotate(angle);
    context.beginPath();
    context.moveTo(0, 0);
    context.lineTo(-size, size * 0.52);
    context.lineTo(-size, -size * 0.52);
    context.closePath();
    context.fillStyle = palette.lineStrong;
    context.fill();
    context.restore();
  }

  function bezierPoint(start, controlOne, controlTwo, end, t) {
    const remaining = 1 - t;
    return {
      x: remaining ** 3 * start.x + 3 * remaining ** 2 * t * controlOne.x + 3 * remaining * t ** 2 * controlTwo.x + t ** 3 * end.x,
      y: remaining ** 3 * start.y + 3 * remaining ** 2 * t * controlOne.y + 3 * remaining * t ** 2 * controlTwo.y + t ** 3 * end.y
    };
  }

  function drawConnector(edge, nodeMap, palette, alpha) {
    const fromNode = nodeMap.get(edge.from);
    const toNode = nodeMap.get(edge.to);
    const start = anchorToward(fromNode, toNode);
    const end = anchorToward(toNode, fromNode);
    const horizontal = start.side === 'left' || start.side === 'right';
    let controlOne;
    let controlTwo;

    if (horizontal) {
      const bend = Math.max(52, Math.abs(end.x - start.x) * 0.46);
      const direction = start.side === 'right' ? 1 : -1;
      controlOne = { x: start.x + bend * direction, y: start.y };
      controlTwo = { x: end.x - bend * direction, y: end.y };
    } else {
      const bend = Math.max(52, Math.abs(end.y - start.y) * 0.46);
      const direction = start.side === 'bottom' ? 1 : -1;
      controlOne = { x: start.x, y: start.y + bend * direction };
      controlTwo = { x: end.x, y: end.y - bend * direction };
    }

    context.save();
    context.globalAlpha = alpha;
    context.beginPath();
    context.moveTo(start.x, start.y);
    context.bezierCurveTo(controlOne.x, controlOne.y, controlTwo.x, controlTwo.y, end.x, end.y);
    context.strokeStyle = palette.lineStrong;
    context.lineWidth = edge.style === 'conditional' ? 1.35 : 1.65;
    if (edge.style === 'conditional') context.setLineDash([7, 7]);
    context.stroke();
    context.setLineDash([]);
    drawArrow(end, Math.atan2(end.y - controlTwo.y, end.x - controlTwo.x), palette, alpha);

    if (edge.label && state.scale >= 0.62 && alpha > 0.65) {
      const midpoint = bezierPoint(start, controlOne, controlTwo, end, 0.5);
      context.font = `600 9px ${palette.mono}`;
      const labelWidth = context.measureText(edge.label).width + 14;
      roundedRect(midpoint.x - labelWidth / 2, midpoint.y - 10, labelWidth, 19, 9.5);
      context.fillStyle = palette.canvas;
      context.fill();
      context.strokeStyle = palette.line;
      context.lineWidth = 0.75;
      context.stroke();
      context.fillStyle = palette.muted;
      context.textAlign = 'center';
      context.textBaseline = 'middle';
      context.fillText(edge.label, midpoint.x, midpoint.y - 0.5);
      context.textBaseline = 'alphabetic';
    }
    context.restore();
  }

  function drawZone(zone, palette, alpha = 1) {
    context.save();
    context.globalAlpha = 0.68 * alpha;
    roundedRect(zone.x, zone.y, zone.width, zone.height, 30);
    context.fillStyle = palette.surface;
    context.fill();
    context.globalAlpha = alpha;
    context.fillStyle = palette.muted;
    context.font = `600 10px ${palette.mono}`;
    context.textAlign = 'left';
    context.fillText(zone.label, zone.x + 20, zone.y + 26);
    if (state.scale >= 0.5) {
      context.font = `400 11px ${palette.sans}`;
      context.fillText(zone.detail, zone.x + 20, zone.y + 45);
    }
    context.restore();
  }

  function nodeColors(node, palette) {
    if (node.kind === 'contract') {
      return {
        fill: palette.inverse,
        stroke: palette.inverse,
        title: palette.inverseInk,
        copy: palette.inverseMuted
      };
    }
    return {
      fill: node.kind === 'guard' || node.kind === 'evidence' ? palette.surfaceStrong : palette.canvas,
      stroke: palette.lineStrong,
      title: palette.ink,
      copy: palette.muted
    };
  }

  function drawNode(node, palette, alpha) {
    const colors = nodeColors(node, palette);
    const selected = node.id === state.selectedId;
    const hovered = node.id === state.hoveredId;

    context.save();
    context.globalAlpha = alpha;
    context.shadowColor = selected ? palette.shadow : hovered ? palette.shadow : 'transparent';
    context.shadowBlur = selected ? 26 : 16;
    context.shadowOffsetY = selected ? 9 : 6;
    roundedRect(node.x, node.y, node.width, node.height, 16);
    context.fillStyle = colors.fill;
    context.fill();
    context.shadowColor = 'transparent';
    context.strokeStyle = colors.stroke;
    context.lineWidth = selected ? 3 : hovered ? 2 : 1.1;
    context.stroke();

    if (node.kind === 'guard') {
      roundedRect(node.x + 5, node.y + 5, node.width - 10, node.height - 10, 12);
      context.strokeStyle = palette.line;
      context.lineWidth = 0.8;
      context.stroke();
    }

    if (node.kind === 'output') {
      context.beginPath();
      context.moveTo(node.x + node.width - 28, node.y);
      context.lineTo(node.x + node.width, node.y + 28);
      context.moveTo(node.x + node.width - 28, node.y);
      context.lineTo(node.x + node.width - 28, node.y + 28);
      context.lineTo(node.x + node.width, node.y + 28);
      context.strokeStyle = colors.stroke;
      context.lineWidth = 1;
      context.stroke();
    }

    const padding = 17;
    context.textAlign = 'left';
    context.fillStyle = colors.copy;
    context.font = `600 9px ${palette.mono}`;
    context.fillText(node.label, node.x + padding, node.y + 23);

    context.fillStyle = colors.title;
    context.font = `700 18px ${palette.sans}`;
    wrapText(node.title, node.x + padding, node.y + 52, node.width - padding * 2, 19, 2);

    if (state.scale >= 0.48) {
      context.fillStyle = colors.copy;
      context.font = `400 12px ${palette.sans}`;
      wrapText(node.detail, node.x + padding, node.y + 86, node.width - padding * 2, 15, 3);
    }

    if (state.scale >= 0.72) {
      context.fillStyle = colors.copy;
      context.font = `500 9px ${palette.mono}`;
      wrapText(node.meta, node.x + padding, node.y + node.height - 16, node.width - padding * 2, 12, 1);
    }
    context.restore();
  }

  function drawGrid(layout, palette) {
    context.save();
    context.fillStyle = palette.line;
    context.globalAlpha = 0.4;
    const radius = 1.05 / state.scale;
    for (let x = 0; x <= layout.width; x += 38) {
      for (let y = 0; y <= layout.height; y += 38) {
        context.beginPath();
        context.arc(x, y, radius, 0, Math.PI * 2);
        context.fill();
      }
    }
    context.restore();
  }

  function nodeMatchesView(node) {
    if (state.view === 'all') return true;
    if (state.view === 'intake') return node.group === 'Surfaces' || node.group === 'Intake paths';
    if (state.view === 'capture') return node.group === 'Capture modules';
    return node.group === 'Target generators' || node.id === 'contract';
  }

  function zoneMatchesView(zone) {
    if (state.view === 'all') return true;
    if (state.view === 'intake') return zone.label === 'SURFACES' || zone.label === 'INTAKE PATHS';
    if (state.view === 'capture') return zone.label === 'CAPTURE MODULES';
    return zone.label === 'TARGET GENERATORS';
  }

  function draw() {
    const palette = readPalette();
    const layout = state.layout;
    const nodeMap = new Map(layout.nodes.map((node) => [node.id, node]));

    context.setTransform(state.pixelRatio, 0, 0, state.pixelRatio, 0, 0);
    context.clearRect(0, 0, state.width, state.height);
    context.fillStyle = palette.canvas;
    context.fillRect(0, 0, state.width, state.height);
    context.save();
    context.translate(state.offsetX, state.offsetY);
    context.scale(state.scale, state.scale);
    drawGrid(layout, palette);
    layout.zones.forEach((zone) => drawZone(zone, palette, zoneMatchesView(zone) ? 1 : 0.12));

    const connectorAlpha = eased((state.progress - 0.1) / 0.38);
    layout.edges.forEach((edge) => {
      const fromMatches = nodeMatchesView(nodeMap.get(edge.from));
      const toMatches = nodeMatchesView(nodeMap.get(edge.to));
      const focusAlpha = state.view === 'all' ? 1 : fromMatches && toMatches ? 1 : fromMatches || toMatches ? 0.3 : 0.08;
      drawConnector(edge, nodeMap, palette, connectorAlpha * focusAlpha);
    });
    layout.nodes.forEach((node) => {
      const nodeAlpha = eased((state.progress - node.order * 0.045) / 0.48);
      const focusAlpha = nodeMatchesView(node) ? 1 : 0.12;
      drawNode(node, palette, nodeAlpha * focusAlpha);
    });
    context.restore();
  }

  function updateZoomOutput() {
    const relative = state.fitScale > 0 ? state.scale / state.fitScale : 1;
    zoomOutput.value = `${Math.round(relative * 100)}%`;
    zoomOutput.textContent = zoomOutput.value;
  }

  function constrainView() {
    const worldWidth = state.layout.width * state.scale;
    const worldHeight = state.layout.height * state.scale;
    const margin = 72;

    if (worldWidth <= state.width - margin * 2) {
      state.offsetX = (state.width - worldWidth) / 2;
    } else {
      state.offsetX = clamp(state.offsetX, state.width - worldWidth - margin, margin);
    }

    if (worldHeight <= state.height - margin * 2) {
      state.offsetY = (state.height - worldHeight) / 2;
    } else {
      state.offsetY = clamp(state.offsetY, state.height - worldHeight - margin, margin);
    }
  }

  function fitBounds(bounds) {
    const compact = state.width < 600;
    const horizontalPadding = compact ? 18 : 44;
    const topPadding = compact ? 118 : 72;
    const bottomPadding = compact ? 24 : 44;
    const availableWidth = Math.max(1, state.width - horizontalPadding * 2);
    const availableHeight = Math.max(1, state.height - topPadding - bottomPadding);
    const readableWidth = compact && state.view !== 'all'
      ? Math.max(1, bounds.width - 100)
      : bounds.width;
    state.fitScale = Math.min(availableWidth / readableWidth, availableHeight / bounds.height);
    state.scale = state.fitScale;
    state.offsetX = horizontalPadding + (availableWidth - bounds.width * state.scale) / 2 - bounds.x * state.scale;
    const verticalAlignment = compact && state.view !== 'all'
      ? 0
      : (availableHeight - bounds.height * state.scale) / 2;
    state.offsetY = topPadding + verticalAlignment - bounds.y * state.scale;
    updateZoomOutput();
    draw();
  }

  function fitDiagram() {
    fitBounds(state.layout.views[state.view] || state.layout.views.all);
  }

  function zoomAt(factor, centerX = state.width / 2, centerY = state.height / 2) {
    const worldX = (centerX - state.offsetX) / state.scale;
    const worldY = (centerY - state.offsetY) / state.scale;
    const minimum = Math.min(state.fitScale * 0.7, 0.3);
    const maximum = Math.max(2.5, state.fitScale * 4);
    state.scale = clamp(state.scale * factor, minimum, maximum);
    state.offsetX = centerX - worldX * state.scale;
    state.offsetY = centerY - worldY * state.scale;
    constrainView();
    updateZoomOutput();
    draw();
  }

  function worldPoint(clientX, clientY) {
    const bounds = canvas.getBoundingClientRect();
    return {
      x: (clientX - bounds.left - state.offsetX) / state.scale,
      y: (clientY - bounds.top - state.offsetY) / state.scale
    };
  }

  function hitTest(clientX, clientY) {
    const point = worldPoint(clientX, clientY);
    return [...state.layout.nodes].reverse().find((node) => (
      point.x >= node.x && point.x <= node.x + node.width &&
      point.y >= node.y && point.y <= node.y + node.height
    ));
  }

  function viewForNode(node) {
    if (node.group === 'Target generators') return 'outputs';
    if (node.group === 'Capture modules') return 'capture';
    return 'intake';
  }

  function updateInspector(node) {
    inspectorKicker.textContent = `${node.group} / ${node.label}`;
    inspectorTitle.textContent = node.title;
    inspectorFile.textContent = node.file;
    inspectorRole.textContent = node.role;
    inspectorContract.textContent = node.contract;
    inspectorInvariants.replaceChildren(...node.invariants.map((item) => {
      const listItem = document.createElement('li');
      listItem.textContent = item;
      return listItem;
    }));
    nodeSelect.value = node.id;
  }

  function selectNode(id, options = {}) {
    const node = nodeCatalog[id];
    if (!node) return;
    state.selectedId = id;
    updateInspector({ id, ...node });

    if (options.focus) {
      state.view = viewForNode(node);
      viewSelect.value = state.view;
      fitDiagram();
      const renderedNode = state.layout.nodes.find((item) => item.id === id);
      if (renderedNode && state.scale < 0.7) {
        const center = nodeCenter(renderedNode);
        const nextScale = Math.min(1, Math.max(0.7, state.scale * 1.35));
        state.scale = nextScale;
        state.offsetX = state.width / 2 - center.x * nextScale;
        state.offsetY = state.height / 2 - center.y * nextScale;
        constrainView();
        updateZoomOutput();
      }
    }
    draw();
  }

  function populateNodeSelect() {
    const groups = new Map();
    nodeOrder.forEach((id) => {
      const node = nodeCatalog[id];
      if (!groups.has(node.group)) groups.set(node.group, []);
      groups.get(node.group).push({ id, node });
    });

    const fragments = [];
    groups.forEach((items, label) => {
      const group = document.createElement('optgroup');
      group.label = label;
      items.forEach(({ id, node }) => {
        const option = document.createElement('option');
        option.value = id;
        option.textContent = node.title;
        group.append(option);
      });
      fragments.push(group);
    });
    nodeSelect.replaceChildren(...fragments);
  }

  function resizeCanvas() {
    const bounds = canvas.getBoundingClientRect();
    const pixelRatio = Math.min(window.devicePixelRatio || 1, 2);
    const nextLayout = bounds.width < 700 ? mobileLayout : desktopLayout;
    state.width = Math.max(1, bounds.width);
    state.height = Math.max(1, bounds.height);
    state.pixelRatio = pixelRatio;
    state.layout = nextLayout;
    canvas.width = Math.round(state.width * pixelRatio);
    canvas.height = Math.round(state.height * pixelRatio);
    fitDiagram();
  }

  function playEntrance() {
    if (state.played) return;
    state.played = true;
    const startedAt = performance.now();
    const duration = 1250;

    function frame(now) {
      state.progress = clamp((now - startedAt) / duration, 0, 1);
      draw();
      if (state.progress < 1) requestAnimationFrame(frame);
    }
    requestAnimationFrame(frame);
  }

  populateNodeSelect();
  updateInspector({ id: state.selectedId, ...nodeCatalog[state.selectedId] });

  viewSelect.addEventListener('change', () => {
    state.view = viewSelect.value;
    fitDiagram();
  });

  nodeSelect.addEventListener('change', () => {
    selectNode(nodeSelect.value, { focus: true });
  });

  shell.addEventListener('pointerdown', (event) => {
    if (event.target !== canvas) return;
    state.dragging = true;
    state.pointerId = event.pointerId;
    state.pointerX = event.clientX;
    state.pointerY = event.clientY;
    state.dragDistance = 0;
    canvas.dataset.dragging = 'true';
    canvas.setPointerCapture(event.pointerId);
  });

  shell.addEventListener('pointermove', (event) => {
    if (state.dragging && event.pointerId === state.pointerId) {
      const deltaX = event.clientX - state.pointerX;
      const deltaY = event.clientY - state.pointerY;
      state.offsetX += deltaX;
      state.offsetY += deltaY;
      state.dragDistance += Math.hypot(deltaX, deltaY);
      state.pointerX = event.clientX;
      state.pointerY = event.clientY;
      constrainView();
      draw();
      return;
    }

    if (event.target === canvas) {
      const hovered = hitTest(event.clientX, event.clientY);
      const hoveredId = hovered?.id || '';
      if (hoveredId !== state.hoveredId) {
        state.hoveredId = hoveredId;
        canvas.style.cursor = hoveredId ? 'pointer' : 'grab';
        draw();
      }
    }
  });

  function endDrag(event) {
    if (!state.dragging || event.pointerId !== state.pointerId) return;
    const wasClick = state.dragDistance < 6;
    state.dragging = false;
    state.pointerId = null;
    delete canvas.dataset.dragging;
    if (wasClick) {
      const node = hitTest(event.clientX, event.clientY);
      if (node) selectNode(node.id, { focus: !nodeMatchesView(node) });
    }
  }

  shell.addEventListener('pointerup', endDrag);
  shell.addEventListener('pointercancel', endDrag);
  canvas.addEventListener('pointerleave', () => {
    if (state.dragging || !state.hoveredId) return;
    state.hoveredId = '';
    canvas.style.cursor = 'grab';
    draw();
  });

  canvas.addEventListener('wheel', (event) => {
    event.preventDefault();
    const bounds = canvas.getBoundingClientRect();
    const factor = Math.exp(-event.deltaY * 0.0012);
    zoomAt(factor, event.clientX - bounds.left, event.clientY - bounds.top);
  }, { passive: false });

  canvas.addEventListener('keydown', (event) => {
    const step = event.shiftKey ? 80 : 36;
    if (event.key === 'ArrowLeft') state.offsetX += step;
    else if (event.key === 'ArrowRight') state.offsetX -= step;
    else if (event.key === 'ArrowUp') state.offsetY += step;
    else if (event.key === 'ArrowDown') state.offsetY -= step;
    else if (event.key === '+' || event.key === '=') zoomAt(1.18);
    else if (event.key === '-') zoomAt(1 / 1.18);
    else if (event.key === '0') fitDiagram();
    else return;

    event.preventDefault();
    constrainView();
    draw();
  });

  shell.addEventListener('click', async (event) => {
    const button = event.target.closest('[data-flow-action]');
    if (!button) return;
    const action = button.dataset.flowAction;
    if (action === 'zoom-in') zoomAt(1.18);
    if (action === 'zoom-out') zoomAt(1 / 1.18);
    if (action === 'fit') fitDiagram();
    if (action === 'fullscreen') {
      try {
        if (document.fullscreenElement === shell) await document.exitFullscreen();
        else if (shell.requestFullscreen) await shell.requestFullscreen();
      } catch (error) {
        fullscreenButton.textContent = 'Full screen unavailable';
      }
    }
  });

  document.addEventListener('fullscreenchange', () => {
    fullscreenButton.textContent = document.fullscreenElement === shell ? 'Exit full screen' : 'Full screen';
    requestAnimationFrame(resizeCanvas);
  });

  new MutationObserver(draw).observe(root, {
    attributes: true,
    attributeFilter: ['data-theme']
  });

  if ('ResizeObserver' in window) {
    new ResizeObserver(resizeCanvas).observe(shell);
  } else {
    window.addEventListener('resize', resizeCanvas);
  }

  resizeCanvas();

  if (reducedMotion.matches || !('IntersectionObserver' in window)) {
    state.progress = 1;
    state.played = true;
    draw();
  } else {
    const canvasObserver = new IntersectionObserver((entries) => {
      if (!entries.some((entry) => entry.isIntersecting)) return;
      playEntrance();
      canvasObserver.disconnect();
    }, { threshold: 0.18 });
    canvasObserver.observe(shell);
  }
})();
