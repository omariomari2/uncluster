# HTML Formatter & JSX Converter

A powerful tool built with Go that formats clustered HTML with proper indentation and converts it to JSX with component suggestions. Perfect for cleaning up messy HTML from web scraping or converting HTML templates to React components.

## Features

- **HTML Formatting**: Transform clustered HTML into properly formatted code with tab indentation
- **JSX Conversion**: Convert HTML to JSX with proper attribute transformations
- **Component Analysis**: Automatically suggest reusable React components from HTML patterns
- **Dual Interface**: Both command-line tool and web interface
- **Fast Performance**: Built with Go for excellent performance
- **Single Binary**: No runtime dependencies required

## Installation

### Prerequisites

- Go 1.21 or later ([Download Go](https://golang.org/dl/))

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

# Process multiple files
for file in *.html; do
    htmlfmt -format -i "$file" -o "formatted_$file"
done
```

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
