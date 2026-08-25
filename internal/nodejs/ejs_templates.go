package nodejs

const ejsPackageJSONTemplate = `{
  "name": "{{.ProjectName}}",
  "version": "1.0.0",
  "type": "module",
  "description": "Generated Express + EJS project from HTML",
  "main": "server.js",
  "scripts": {
    "start": "node server.js",
    "dev": "nodemon server.js"
  },
  "dependencies": {
    "express": "^4.18.2",
    "ejs": "^3.1.9"
  },
  "devDependencies": {
    "nodemon": "^3.0.2"
  }
}`

const ejsServerJSTemplate = `import express from 'express'
import path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const app = express()
const PORT = process.env.PORT || 8080

app.set('view engine', 'ejs')
app.set('views', path.join(__dirname, 'views'))

app.use(express.static(path.join(__dirname, 'public')))

// Only extensionless paths fall through to the page. A request for a file that
// does not exist has to 404: rendering HTML for it would return 200 with the
// wrong content type, so a missing stylesheet or script would fail silently in
// the browser instead of showing up.
app.get('*', (req, res, next) => {
  if (path.extname(req.path)) {
    return next()
  }
  res.render('index')
})

app.use((req, res) => {
  res.status(404).type('text/plain').send('Not found: ' + req.path)
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

3. Open your browser to http://localhost:8080

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
