Tree Data Structure
HTML is parsed into a DOM Tree using golang.org/x/net/html parser. Each node represents an HTML element with parent-child relationships.

Node Structure: Each node contains Type, Data, Attributes, FirstChild, NextSibling pointers
Tree Traversal: Depth-first recursive traversal processes all HTML elements
Hash Map Pattern Matching
Component analysis uses hash maps to track element patterns efficiently.

Pattern Key Generation: Creates unique keys using tag.class#id for fast lookup
Attribute Frequency Map: Tracks attribute occurrences using map[string]int
Child Element Counting: Maps child tag names to frequencies
Recursive Algorithms
Multiple recursive algorithms process the DOM tree structure.

collectPatterns(): Recursively traverses tree to collect element patterns
formatNode(): Recursive formatting with depth tracking for indentation
extractStylesAndScripts(): Recursive extraction of inline CSS/JS
String Processing
Efficient string manipulation using builders and regex.

strings.Builder: Efficient string concatenation for building output
bytes.Buffer: Used for rendering HTML output efficiently
Regex Matching: Pattern-based attribute conversion (class→className, etc.)
String Replacement: Multiple passes for JSX conversion
Graph-like Structures
The DOM tree is essentially a directed acyclic graph (DAG).

Adjacency Representation: FirstChild/NextSibling pointers create linked structure
External Resource Graph: Tracks dependencies between HTML, CSS, and JS files
Dependency Resolution: Ensures proper ordering of external resources
AI-Powered Component Analysis
Uncluster uses Cloudflare Workers Llama 3 SDK AI to intelligently analyze HTML elements and determine which should become React components.

Pattern Recognition: Identifies meaningful, reusable patterns vs generic wrapper divs
Semantic Analysis: Distinguishes between components (cards, buttons, forms) and layout containers
Reusability Assessment: Determines if elements would benefit from props and componentization
Confidence Scoring: Provides high, medium, or low confidence ratings for each suggestion
