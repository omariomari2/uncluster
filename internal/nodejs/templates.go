package nodejs

const packageJSONTemplate = `{
  "name": "{{.ProjectName}}",
  "version": "1.0.0",
  "type": "module",
  "description": "Generated React TypeScript project from HTML",
  "main": "server.js",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview",
    "serve": "node server.js",
    "lint": "eslint . --ext .ts,.tsx,.js,.jsx",
    "format": "prettier --write .",
    "start": "npm run serve",
    "type-check": "tsc --noEmit"
  },
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "express": "^4.18.2"
  },
  "devDependencies": {
    "@types/react": "^18.2.43",
    "@types/react-dom": "^18.2.17",
    "@typescript-eslint/eslint-plugin": "^6.14.0",
    "@typescript-eslint/parser": "^6.14.0",
    "@vitejs/plugin-react": "^4.2.1",
    "eslint": "^8.55.0",
    "eslint-plugin-react-hooks": "^4.6.0",
    "eslint-plugin-react-refresh": "^0.4.5",
    "prettier": "^3.1.0",
    "typescript": "^5.3.0",
    "vite": "^5.0.0"
  },
  "keywords": ["react", "typescript", "vite", "express", "jsx"],
  "author": "",
  "license": "MIT"
}`

const viteConfigTemplate = `import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  root: 'src',
  publicDir: '../public',
  build: {
    outDir: '../dist',
    emptyOutDir: true
  },
  server: {
    port: 8080,
    open: true,
    host: true
  },
  preview: {
    port: 8080,
    open: true,
    host: true
  }
})`

const serverJSTemplate = `import express from 'express'
import path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const app = express()
const PORT = process.env.PORT || 8080

app.use(express.static(path.join(__dirname, 'dist')))

app.get('*', (req, res) => {
  res.sendFile(path.join(__dirname, 'dist', 'index.html'))
})

app.listen(PORT, () => {
  console.log('Server running at http://localhost:' + PORT)
  console.log('Serving files from: ' + path.join(__dirname, 'dist'))
})`

const eslintConfigTemplate = `{
  "env": {
    "browser": true,
    "es2021": true,
    "node": true
  },
  "extends": [
    "eslint:recommended",
    "@typescript-eslint/recommended",
    "plugin:react-hooks/recommended"
  ],
  "parser": "@typescript-eslint/parser",
  "parserOptions": {
    "ecmaVersion": "latest",
    "sourceType": "module",
    "ecmaFeatures": {
      "jsx": true
    }
  },
  "plugins": [
    "react-refresh",
    "@typescript-eslint"
  ],
  "rules": {
    "indent": ["error", 2],
    "linebreak-style": ["error", "unix"],
    "quotes": ["error", "single"],
    "semi": ["error", "always"],
    "no-unused-vars": "off",
    "@typescript-eslint/no-unused-vars": "warn",
    "no-console": "off",
    "react-refresh/only-export-components": [
      "warn",
      { "allowConstantExport": true }
    ]
  },
  "globals": {
    "process": "readonly"
  }
}`

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
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "allowSyntheticDefaultImports": true,
    "esModuleInterop": true,
    "forceConsistentCasingInFileNames": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist"]
}`

const gitignoreTemplate = `
node_modules/
npm-debug.log*
yarn-debug.log*
yarn-error.log*

dist/
build/

.env
.env.local
.env.development.local
.env.test.local
.env.production.local

.vscode/
.idea/
*.swp
*.swo

.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

logs
*.log

pids
*.pid
*.seed
*.pid.lock

coverage/

.nyc_output

jspm_packages/

.npm

.eslintcache

.rpt2_cache/
.rts2_cache_cjs/
.rts2_cache_es/
.rts2_cache_umd/

.node_repl_history

*.tgz

.yarn-integrity

.cache
.parcel-cache

.next

.nuxt

.vuepress/dist

.serverless

.fusebox/

.dynamodb/

.tern-port
`

const readmeTemplate = `# {{.ProjectName}}

A React TypeScript project generated from HTML with Vite build system and Express server.

## Features

- **React 18** - Modern React with hooks and concurrent features
- **TypeScript** - Type safety and enhanced developer experience
- **Vite** - Fast build tool and development server
- **Express** - Production-ready web server
- **Hot Module Reloading** - Instant updates during development
- **ESLint** - Code quality and consistency with React rules
- **Prettier** - Code formatting
- **Component-based** - Modular JSX/TSX components
- **Modern Tooling** - Full TypeScript and React development setup

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

3. Open your browser to http://localhost:8080

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
    ├── index.html        # Vite entry HTML
    ├── main.tsx          # React entry point
    ├── App.tsx           # Main App component
    ├── components/
    │   ├── MainComponent.tsx  # Converted HTML component
    │   └── Component*.tsx     # Additional components
    └── styles/
        ├── main.css      # Your inline styles
        └── external/     # Downloaded external CSS
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

3. The server will run on http://localhost:8080 (or PORT environment variable)

## Customization

- **Components**: Edit files in ` + "`" + `src/components/` + "`" + `
- **Styling**: Edit files in ` + "`" + `src/styles/` + "`" + `
- **Main App**: Edit ` + "`" + `src/App.tsx` + "`" + `
- **Entry Point**: Edit ` + "`" + `src/main.tsx` + "`" + `
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

const mainTsxFallback = `import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
`

const indexHtmlTemplate = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>{{.ProjectName}}</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/main.tsx"></script>
  </body>
</html>
`

const appTsxTemplate = `import React from 'react'
import MainComponent from './components/MainComponent'

function App() {
  return (
    <div className="App">
      <MainComponent />
    </div>
  )
}

export default App
`
