# HTML Formatter & JSX Converter

A powerful tool built with Go that formats clustered HTML with proper indentation and converts it to JSX with component suggestions. Perfect for cleaning up messy HTML from web scraping or converting HTML templates to React components.

## Features

- **HTML Formatting**: Transform clustered HTML into properly formatted code with tab indentation
- **JSX Conversion**: Convert HTML to JSX with proper attribute transformations
- **AI-Powered Component Analysis**: Intelligently identify meaningful React components using Cloudflare Workers AI (Llama 3) - prevents every div from becoming a component
- **Component Analysis**: Automatically suggest reusable React components from HTML patterns
- **Zip Export**: Extract inline styles/scripts and download external resources into organized zip files
- **React TypeScript Export**: Transform HTML into production-ready React TypeScript applications with Vite, Express, and modern tooling
- **External Resource Fetching**: Automatically download external CSS and JS files from URLs
- **Dual Interface**: Both command-line tool and web interface
- **Fast Performance**: Built with Go for excellent performance
- **Single Binary**: No runtime dependencies required

## Installation

### Prerequisites

- Go 1.21 or later ([Download Go](https://golang.org/dl/))
- (Optional) Cloudflare account for AI-powered component analysis

### Build from Source

```bash
# Clone or download the project
git clone <repository-url>
cd htmlfmt

# Build the CLI tool
go build -o htmlfmt cmd/htmlfmt/main.go

# Build the web server
go build -o htmlfmt-server api/server.go api/handlers.go
```

### Cross-platform Builds

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o htmlfmt.exe cmd/htmlfmt/main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o htmlfmt-macos cmd/htmlfmt/main.go

# Linux
GOOS=linux GOARCH=amd64 go build -o htmlfmt-linux cmd/htmlfmt/main.go
```

## Cloudflare Workers AI Setup (Optional)

The tool includes AI-powered component analysis using Cloudflare Workers AI with Llama 3. This feature intelligently identifies which HTML elements should become React components, preventing generic divs from being converted unnecessarily.

### Benefits of AI-Powered Analysis

- **Smart Component Detection**: AI analyzes HTML structure and semantic meaning
- **Filters Generic Elements**: Prevents every div from becoming a component
- **Identifies Patterns**: Recognizes cards, buttons, forms, navigation items, and other meaningful patterns
- **Provides Reasoning**: Explains why elements should or shouldn't be components
- **Graceful Fallback**: If AI is unavailable, falls back to pattern-based analysis

### Setup Instructions

1. **Create a Cloudflare Account**
   - Sign up at [cloudflare.com](https://dash.cloudflare.com/sign-up) if you don't have an account

2. **Get Your Account ID**
   - Go to your Cloudflare dashboard
   - Your Account ID is shown in the **Overview** section (right sidebar)

3. **Generate an API Token**
   - Navigate to **My Profile** > **API Tokens**
   - Click **Create Token**
   - Use the **Edit Cloudflare Workers** template or create a custom token with:
     - **Account** > **Workers AI** > **Read**
   - Copy the generated token (you won't be able to see it again)

4. **Agree to Meta's License (Required for Llama 3)**
   - Before using Llama 3, you must agree to Meta's License
   - Run this command (replace with your credentials):
   ```bash
   curl https://api.cloudflare.com/client/v4/accounts/$CLOUDFLARE_ACCOUNT_ID/ai/run/@cf/meta/llama-3-8b-instruct \
     -X POST \
     -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
     -d '{ "prompt": "agree" }'
   ```

5. **Set Environment Variables**
   ```bash
   export CLOUDFLARE_ACCOUNT_ID="your_account_id_here"
   export CLOUDFLARE_API_TOKEN="your_api_token_here"
   export CLOUDFLARE_AI_MODEL="@cf/meta/llama-3-8b-instruct"  # Optional, this is the default
   ```

   Or create a `.env` file (if using a tool like `godotenv`):
   ```
   CLOUDFLARE_ACCOUNT_ID=your_account_id_here
   CLOUDFLARE_API_TOKEN=your_api_token_here
   CLOUDFLARE_AI_MODEL=@cf/meta/llama-3-8b-instruct
   ```

### Worker AI SDK Option (Alternative)

If you prefer using the Cloudflare Workers AI SDK, deploy the worker in `workers/component-analyzer` and set:

```bash
export CLOUDFLARE_WORKER_URL="https://your-worker.example.workers.dev"
export CLOUDFLARE_WORKER_TOKEN="your_worker_token"  # Optional if you do not enforce auth
export CLOUDFLARE_WORKER_MODEL="@cf/meta/llama-3-8b-instruct"  # Optional
```

When `CLOUDFLARE_WORKER_URL` is set, the server uses the worker endpoint and skips direct API credentials.
See `workers/component-analyzer/README.md` for deployment steps.

6. **Start the Server**
   - The server will automatically detect Cloudflare credentials and enable AI analysis
   - You'll see a log message: `✅ Cloudflare AI initialized (Model: @cf/meta/llama-3-8b-instruct)`

### Without Cloudflare AI

If you don't configure Cloudflare credentials, the tool will still work perfectly using pattern-based component detection. You'll see:
```
ℹ️  Cloudflare AI not configured (CLOUDFLARE_ACCOUNT_ID and CLOUDFLARE_API_TOKEN required)
ℹ️  Component analysis will use pattern-based detection only
```

### Available Models

- `@cf/meta/llama-3-8b-instruct` (default) - Recommended for component analysis
- `@cf/meta/llama-3.1-8b-instruct` - Alternative Llama 3.1 model
- `@cf/meta/llama-3.2-11b-vision-instruct` - Vision-capable model (requires license agreement)

## Usage

### Command Line Interface

#### Basic Usage

```bash
# Format HTML
htmlfmt -format -i messy.html -o formatted.html

# Convert to JSX
htmlfmt -jsx -i page.html -o component.jsx

# Analyze for component suggestions
htmlfmt -analyze -i complex.html

# Process from stdin
cat input.html | htmlfmt -format

# Quick JSX conversion
echo '<div class="card"><h2>Title</h2></div>' | htmlfmt -jsx
```

#### Command Line Options

- `-format`: Format HTML with proper indentation
- `-jsx`: Convert HTML to JSX
- `-analyze`: Analyze HTML and suggest components (outputs JSON)
- `-export`: Export HTML as zip with separated CSS/JS and external resources
- `-nodejs`: Export as React TypeScript project with Vite, Express, and tooling
- `-i <file>`: Input file (default: stdin)
- `-o <file>`: Output file (default: stdout)
- `-h`: Show help
- `-v`: Show version

#### Examples

```bash
# Format a messy HTML file
htmlfmt -format -i scraped.html -o clean.html

# Convert HTML template to JSX
htmlfmt -jsx -i template.html -o Template.jsx

# Get component suggestions as JSON
htmlfmt -analyze -i complex-page.html > suggestions.json

# Export HTML as zip with external resources
htmlfmt -export -i page.html -o extracted.zip

# Export as React TypeScript project
htmlfmt -nodejs -i page.html -o my-project.zip

# Process multiple files
for file in *.html; do
    htmlfmt -format -i "$file" -o "formatted_$file"
done
```

## External Resource Fetching

The zip export feature automatically downloads external CSS and JavaScript files referenced in your HTML, creating a complete offline package.

### How It Works

1. **Detection**: Scans HTML for `<link rel="stylesheet" href="...">` and `<script src="...">` tags with external URLs
2. **Download**: Fetches external resources using HTTP client with 10-second timeout
3. **Organization**: Places external files in `external/css/` and `external/js/` folders
4. **Rewriting**: Updates HTML links to point to local files instead of external URLs

### Zip Structure

```
extracted.zip/
├── index.html                    # Cleaned HTML with rewritten links
├── style.css                     # Inline styles from <style> tags
├── script.js                     # Inline scripts from <script> tags
└── external/
    ├── css/
    │   ├── bootstrap.min.css     # Downloaded external CSS
    │   ├── custom.css
    │   └── ...
    └── js/
        ├── jquery.min.js         # Downloaded external JS
        ├── app.js
        └── ...
```

### Features

- **Automatic Detection**: Finds external resources without configuration
- **Error Handling**: Skips failed downloads and continues processing
- **Safe Filenames**: Sanitizes URLs to create filesystem-safe filenames
- **Duplicate Handling**: Prevents filename conflicts with counters
- **Network Requirements**: Requires internet access for external resource downloads

### Example

```bash
# Export HTML with external resources
htmlfmt -export -i webpage.html -o complete.zip

# The zip will contain:
# - All inline styles and scripts extracted
# - All external CSS/JS files downloaded
# - HTML with links rewritten to local files
```

## React TypeScript Project Export

Transform your HTML into a production-ready React TypeScript application with modern tooling and development workflows.

### Features

- ⚛️ **React 18** - Modern React with hooks and concurrent features
- 📘 **TypeScript** - Type safety and enhanced developer experience
- ⚡ **Vite Build System** - Fast development server and optimized production builds
- 🚀 **Express Server** - Production-ready web server for deployment
- 🔥 **Hot Module Reloading** - Instant updates during development
- 📝 **ESLint** - Code quality and consistency checking with React rules
- 💅 **Prettier** - Automatic code formatting
- 🧩 **Component-based** - Modular JSX/TSX components
- 📁 **Organized Structure** - Clean separation of concerns with src/, components/, styles/

### Usage

**Web Interface:**
1. Select "Export as Node.js Project" mode
2. Paste or upload your HTML
3. Click "Process" to download the project

**CLI:**
```bash
htmlfmt -nodejs -i page.html -o my-project.zip
```

### Getting Started

After downloading and extracting:

```bash
cd my-project
npm install
npm run dev          # Start development server
npm run build        # Build for production
npm run serve        # Serve production build
npm run lint         # Check code quality
npm run format       # Format code
```

### Project Structure

```
my-project/
├── package.json          # Dependencies and scripts
├── vite.config.js        # Vite build configuration
├── server.js             # Express production server
├── .eslintrc.json        # ESLint configuration
├── .prettierrc           # Prettier configuration
├── tsconfig.json         # TypeScript configuration
├── .gitignore            # Git ignore rules
├── README.md             # Project documentation
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
```

### Development Workflow

1. **Development**: `npm run dev` - Start Vite dev server with hot reload
2. **Building**: `npm run build` - Create optimized production build
3. **Preview**: `npm run preview` - Preview production build locally
4. **Production**: `npm run serve` - Start Express server for production

### Customization

- **Components**: Edit files in `src/components/`
- **Styling**: Edit files in `src/styles/`
- **Main App**: Edit `src/App.tsx`
- **Entry Point**: Edit `src/main.tsx`
- **Build config**: Modify `vite.config.js`
- **Server config**: Modify `server.js`

### Web Interface

#### Starting the Server

```bash
# Run the web server
go run api/server.go api/handlers.go

# Or use the built binary
./htmlfmt-server
```

The web interface will be available at `http://localhost:3000`

#### Using the Web Interface

1. **Upload or Paste HTML**: Either upload an HTML file or paste your clustered HTML into the text area
2. **Choose Mode**: Select either "Format HTML" or "Convert to JSX"
3. **Process**: Click the "Process" button to transform your HTML
4. **View Results**: See the formatted output in the right panel
5. **Component Suggestions**: In JSX mode, view suggested React components
6. **Download**: Save the result as a file or copy to clipboard

#### Web Interface Features

- **File Upload**: Drag and drop or click to upload HTML files
- **Real-time Preview**: See formatted output immediately
- **Component Analysis**: Automatic suggestions for reusable components
- **Copy to Clipboard**: One-click copying of results
- **Download**: Save results as `.html` or `.jsx` files
- **Responsive Design**: Works on desktop and mobile devices

## API Endpoints

The web server provides REST API endpoints:

### POST /api/format
Format HTML with proper indentation.

**Request:**
```json
{
  "html": "<div><p>Hello</p></div>"
}
```

**Response:**
```json
{
  "success": true,
  "data": "<div>\n\t<p>Hello</p>\n</div>"
}
```

### POST /api/convert
Convert HTML to JSX.

**Request:**
```json
{
  "html": "<div class=\"card\"><h2>Title</h2></div>"
}
```

**Response:**
```json
{
  "success": true,
  "data": "<div className=\"card\">\n\t<h2>Title</h2>\n</div>"
}
```

### POST /api/analyze
Analyze HTML for component suggestions.

**Request:**
```json
{
  "html": "<div class=\"card\">...</div><div class=\"card\">...</div>"
}
```

**Response:**
```json
{
  "success": true,
  "suggestions": [
    {
      "name": "DivCardComponent",
      "description": "A reusable div component (appears 2 times)",
      "tagName": "div",
      "attributes": {
        "className": "{string}"
      },
      "count": 2,
      "jsxCode": "const DivCardComponent = ({ className=\"{string}\" }) => {\n\treturn (\n\t\t<div className={className}>\n\t\t\t{/* Add your content here */}\n\t\t</div>\n\t);\n};\n\nexport default DivCardComponent;"
    }
  ]
}
```

## HTML Formatting Features

- **Tab Indentation**: Uses tabs for consistent formatting
- **Proper Nesting**: Correctly indents nested elements
- **Self-closing Tags**: Handles void elements properly
- **Attribute Formatting**: Preserves all attributes with proper spacing
- **Comment Support**: Formats HTML comments correctly
- **Doctype Support**: Handles DOCTYPE declarations

## JSX Conversion Features

### Attribute Transformations

- `class` → `className`
- `for` → `htmlFor`
- `tabindex` → `tabIndex`
- `readonly` → `readOnly`
- `maxlength` → `maxLength`
- And many more camelCase conversions

### Style Conversion

Inline CSS styles are converted to JSX style objects:

```html
<!-- Input -->
<div style="color: red; font-size: 16px; margin-top: 10px">

<!-- Output -->
<div style={{color: "red", fontSize: 16, marginTop: "10px"}}>
```

### Self-closing Tags

All void elements are properly self-closed:

```html
<!-- Input -->
<img src="image.jpg" alt="Image">
<br>

<!-- Output -->
<img src="image.jpg" alt="Image" />
<br />
```

## Component Analysis

The tool automatically analyzes HTML structure to suggest reusable React components:

- **AI-Powered Analysis**: (When Cloudflare AI is configured) Intelligently filters components using Llama 3, preventing generic divs from becoming components
- **Pattern Recognition**: Identifies repeated element structures
- **Attribute Analysis**: Suggests props based on common attributes
- **Component Generation**: Creates example component code
- **Usage Statistics**: Shows how many times patterns appear

### Example Component Suggestion

For HTML like:
```html
<div class="card">
    <h3>Title</h3>
    <p>Description</p>
</div>
<div class="card">
    <h3>Another Title</h3>
    <p>Another Description</p>
</div>
```

The tool suggests:
```jsx
const DivCardComponent = ({ className="{string}" }) => {
    return (
        <div className={className}>
            {/* Add your content here */}
        </div>
    );
};

export default DivCardComponent;
```

## Project Structure

```
htmlfmt/
├── cmd/
│   └── htmlfmt/
│       └── main.go           # CLI entry point
├── internal/
│   ├── formatter/
│   │   └── formatter.go      # HTML formatting logic
│   ├── converter/
│   │   └── jsx.go            # HTML to JSX conversion
│   └── analyzer/
│       └── components.go     # Component analysis
├── api/
│   ├── server.go             # HTTP server
│   └── handlers.go           # API handlers
├── web/
│   └── static/
│       ├── index.html        # Web interface
│       ├── styles.css        # Styling
│       └── app.js            # Frontend logic
├── go.mod                    # Go module
└── README.md                 # This file
```

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
# CLI tool
go build -o htmlfmt cmd/htmlfmt/main.go

# Web server
go build -o htmlfmt-server api/server.go api/handlers.go
```

### Dependencies

- `golang.org/x/net/html` - HTML parsing
- `github.com/gofiber/fiber/v2` - Web framework

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Troubleshooting

### Common Issues

1. **Go not found**: Make sure Go is installed and in your PATH
2. **Permission denied**: Use `chmod +x htmlfmt` on Unix systems
3. **Port already in use**: Change the port by setting the `PORT` environment variable

### Getting Help

- Check the help: `htmlfmt -h`
- View version: `htmlfmt -v`
- Open an issue on GitHub for bugs or feature requests

## Examples

### Formatting Clustered HTML

**Input:**
```html
<div class="container"><h1>Title</h1><p>Paragraph with <strong>bold</strong> text</p><ul><li>Item 1</li><li>Item 2</li></ul></div>
```

**Output:**
```html
<div class="container">
	<h1>Title</h1>
	<p>Paragraph with <strong>bold</strong> text</p>
	<ul>
		<li>Item 1</li>
		<li>Item 2</li>
	</ul>
</div>
```

### Converting to JSX

**Input:**
```html
<div class="card" style="margin: 10px; padding: 20px;">
    <h2 class="title">Card Title</h2>
    <p class="content">Card content goes here</p>
    <button class="btn-primary" onclick="handleClick()">Click me</button>
</div>
```

**Output:**
```jsx
<div className="card" style={{margin: "10px", padding: "20px"}}>
	<h2 className="title">Card Title</h2>
	<p className="content">Card content goes here</p>
	<button className="btn-primary" onClick={handleClick}>Click me</button>
</div>
```
