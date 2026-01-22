package nodejs

const ejsPackageJSONTemplate = `{
  "name": "{{.ProjectName}}",
  "version": "1.0.0",
  "type": "module",
  "description": "Generated Express + EJS project from HTML",
  "main": "server.js",
  "scripts": {
    "start": "node server.js",
    "dev": "node server.js"
  },
  "dependencies": {
    "express": "^4.18.2",
    "ejs": "^3.1.9"
  }
}`

const ejsServerJSTemplate = `import express from 'express'
import path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const app = express()
const PORT = process.env.PORT || 3000

app.set('view engine', 'ejs')
app.set('views', path.join(__dirname, 'views'))

// Serve static assets from /public
app.use(express.static(path.join(__dirname, 'public')))

app.get('*', (req, res) => {
  res.render('index')
})

app.listen(PORT, () => {
  console.log('Server running at http://localhost:' + PORT)
  console.log('Serving views from: ' + path.join(__dirname, 'views'))
})
`

const ejsReadmeTemplate = `# {{.ProjectName}}

An Express + EJS project generated from HTML.

## Quick Start

1. Install dependencies:
   ` + "```" + `bash
   npm install
   ` + "```" + `

2. Start the server:
   ` + "```" + `bash
   npm start
   ` + "```" + `

3. Open your browser to http://localhost:3000

## Project Structure

` + "```" + `
{{.ProjectName}}/
  package.json
  server.js
  .gitignore
  README.md
  views/
    index.ejs
    partials/
  public/
    inline/
    external/
` + "```" + `

## Notes

- The original HTML is preserved in ` + "`" + `views/index.ejs` + "`" + `.
- Reusable sections are extracted into ` + "`" + `views/partials/` + "`" + `.
- Static assets are served from ` + "`" + `public/` + "`" + `.
`
