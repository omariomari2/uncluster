package nodejs

// packageJSONTemplate is the template for package.json
const packageJSONTemplate = `{
  "name": "{{.ProjectName}}",
  "version": "1.0.0",
  "type": "module",
  "description": "Generated Node.js project from HTML",
  "main": "server.js",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview",
    "serve": "node server.js",
    "lint": "eslint . --ext .js,.html",
    "format": "prettier --write .",
    "start": "npm run serve"
  },
  "dependencies": {
    "express": "^4.18.2"
  },
  "devDependencies": {
    "vite": "^5.0.0",
    "eslint": "^8.55.0",
    "prettier": "^3.1.0",
    "typescript": "^5.3.0",
    "@types/node": "^20.10.0"
  },
  "keywords": ["html", "vite", "express", "nodejs"],
  "author": "",
  "license": "MIT"
}`

// viteConfigTemplate is the template for vite.config.js
const viteConfigTemplate = `import { defineConfig } from 'vite'

export default defineConfig({
  root: 'src',
  publicDir: '../public',
  build: {
    outDir: '../dist',
    emptyOutDir: true,
    rollupOptions: {
      input: {
        main: 'src/index.html'
      }
    }
  },
  server: {
    port: 3000,
    open: true,
    host: true
  },
  preview: {
    port: 3000,
    open: true,
    host: true
  }
})`

// serverJSTemplate is the template for server.js
const serverJSTemplate = `import express from 'express'
import path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const app = express()
const PORT = process.env.PORT || 3000

// Serve static files from the dist directory
app.use(express.static(path.join(__dirname, 'dist')))

// Handle client-side routing - return index.html for all routes
app.get('*', (req, res) => {
  res.sendFile(path.join(__dirname, 'dist', 'index.html'))
})

app.listen(PORT, () => {
  console.log('Server running at http://localhost:' + PORT)
  console.log('Serving files from: ' + path.join(__dirname, 'dist'))
})`

// eslintConfigTemplate is the template for .eslintrc.json
const eslintConfigTemplate = `{
  "env": {
    "browser": true,
    "es2021": true,
    "node": true
  },
  "extends": [
    "eslint:recommended"
  ],
  "parserOptions": {
    "ecmaVersion": "latest",
    "sourceType": "module"
  },
  "rules": {
    "indent": ["error", 2],
    "linebreak-style": ["error", "unix"],
    "quotes": ["error", "single"],
    "semi": ["error", "always"],
    "no-unused-vars": "warn",
    "no-console": "off"
  },
  "globals": {
    "process": "readonly"
  }
}`

// prettierConfigTemplate is the template for .prettierrc
const prettierConfigTemplate = `{
  "semi": true,
  "trailingComma": "es5",
  "singleQuote": true,
  "printWidth": 80,
  "tabWidth": 2,
  "useTabs": false,
  "bracketSpacing": true,
  "arrowParens": "avoid"
}`

// tsconfigTemplate is the template for tsconfig.json
const tsconfigTemplate = `{
  "compilerOptions": {
    "target": "ES2020",
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "preserve",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist"]
}`

// gitignoreTemplate is the template for .gitignore
const gitignoreTemplate = `# Dependencies
node_modules/
npm-debug.log*
yarn-debug.log*
yarn-error.log*

# Build outputs
dist/
build/

# Environment variables
.env
.env.local
.env.development.local
.env.test.local
.env.production.local

# IDE files
.vscode/
.idea/
*.swp
*.swo

# OS generated files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# Logs
logs
*.log

# Runtime data
pids
*.pid
*.seed
*.pid.lock

# Coverage directory used by tools like istanbul
coverage/

# nyc test coverage
.nyc_output

# Dependency directories
jspm_packages/

# Optional npm cache directory
.npm

# Optional eslint cache
.eslintcache

# Microbundle cache
.rpt2_cache/
.rts2_cache_cjs/
.rts2_cache_es/
.rts2_cache_umd/

# Optional REPL history
.node_repl_history

# Output of 'npm pack'
*.tgz

# Yarn Integrity file
.yarn-integrity

# parcel-bundler cache (https://parceljs.org/)
.cache
.parcel-cache

# next.js build output
.next

# nuxt.js build output
.nuxt

# vuepress build output
.vuepress/dist

# Serverless directories
.serverless

# FuseBox cache
.fusebox/

# DynamoDB Local files
.dynamodb/

# TernJS port file
.tern-port
`

// readmeTemplate is the template for README.md
const readmeTemplate = `# {{.ProjectName}}

A Node.js project generated from HTML with Vite build system and Express server.

## Features

- **Vite** - Fast build tool and development server
- **Express** - Production-ready web server
- **Hot Module Reloading** - Instant updates during development
- **ESLint** - Code quality and consistency
- **Prettier** - Code formatting
- **TypeScript** - Type safety and editor support
- **Organized Structure** - Clean separation of concerns

## Quick Start

### Prerequisites

- Node.js 18+ 
- npm (comes with Node.js)

### Installation

1. Install dependencies:
   ` + "```" + `bash
   npm install
   ` + "```" + `

2. Start development server:
   ` + "```" + `bash
   npm run dev
   ` + "```" + `

3. Open your browser to http://localhost:3000

## Available Scripts

- ` + "`" + `npm run dev` + "`" + ` - Start development server with hot reload
- ` + "`" + `npm run build` + "`" + ` - Build for production
- ` + "`" + `npm run preview` + "`" + ` - Preview production build locally
- ` + "`" + `npm run serve` + "`" + ` - Start production server
- ` + "`" + `npm run lint` + "`" + ` - Check code quality with ESLint
- ` + "`" + `npm run format` + "`" + ` - Format code with Prettier

## Project Structure

` + "```" + `
{{.ProjectName}}/
├── package.json          # Dependencies and scripts
├── vite.config.js        # Vite configuration
├── server.js             # Express production server
├── .eslintrc.json        # ESLint configuration
├── .prettierrc           # Prettier configuration
├── tsconfig.json         # TypeScript configuration
├── .gitignore            # Git ignore rules
├── README.md             # This file
└── src/
    ├── index.html        # Main HTML file
    ├── styles/
    │   ├── main.css      # Your inline styles
    │   └── external/     # Downloaded external CSS
    └── scripts/
        ├── main.js       # Your inline scripts
        └── external/     # Downloaded external JS
` + "```" + `

## Development

The project uses Vite for development, which provides:

- **Instant server start** - No bundling required
- **Hot Module Replacement (HMR)** - Update modules without page reload
- **Optimized builds** - Rollup-based production builds
- **TypeScript support** - Built-in TypeScript support

## Production Deployment

1. Build the project:
   ` + "```" + `bash
   npm run build
   ` + "```" + `

2. Start the production server:
   ` + "```" + `bash
   npm run serve
   ` + "```" + `

3. The server will run on http://localhost:3000 (or PORT environment variable)

## Customization

- **Styling**: Edit files in ` + "`" + `src/styles/` + "`" + `
- **JavaScript**: Edit files in ` + "`" + `src/scripts/` + "`" + `
- **HTML**: Edit ` + "`" + `src/index.html` + "`" + `
- **Build config**: Modify ` + "`" + `vite.config.js` + "`" + `
- **Server config**: Modify ` + "`" + `server.js` + "`" + `

## External Resources

This project includes the following external resources that were automatically downloaded:

{{if .ExternalCSS}}
### CSS Files
{{range .ExternalCSS}}
- ` + "`" + `src/styles/external/{{.Filename}}` + "`" + ` ({{.URL}})
{{end}}
{{end}}

{{if .ExternalJS}}
### JavaScript Files
{{range .ExternalJS}}
- ` + "`" + `src/scripts/external/{{.Filename}}` + "`" + ` ({{.URL}})
{{end}}
{{end}}

## License

MIT
`
