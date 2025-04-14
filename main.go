package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// Template for the HTML page
const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Chaitanya Nettem - Personal Website</title>
    <link rel="stylesheet" href="styles.css">
</head>
<body>
    <div class="container">
        %s
    </div>
</body>
</html>`

func main() {
	// Create output directory if it doesn't exist
	err := os.MkdirAll("public", 0755)
	if err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	// Read all markdown files from the content directory
	files, err := filepath.Glob("content/*.md")
	if err != nil {
		fmt.Printf("Error finding markdown files: %v\n", err)
		return
	}

	for _, file := range files {
		// Read the markdown file
		mdContent, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", file, err)
			continue
		}

		// Convert markdown to HTML
		htmlContent := markdownToHTML(mdContent)

		// Custom post-processing for the navigation links
		re := regexp.MustCompile(`<p>(<a href=".*?">BLOG</a>) \| (<a href=".*?">RESUME</a>) \| (<a href=".*?">GITHUB</a>) \| (<a href=".*?">LINKEDIN</a>)</p>`)
		htmlContent = re.ReplaceAllString(
			htmlContent,
			`<div class="nav-links">$1 &nbsp;|&nbsp; $2 &nbsp;|&nbsp; $3 &nbsp;|&nbsp; $4</div>`,
		)

		// Get the output filename
		baseFilename := filepath.Base(file)
		outFilename := strings.TrimSuffix(
			baseFilename,
			filepath.Ext(baseFilename),
		) + ".html"

		// Insert the HTML content into the template
		completeHTML := fmt.Sprintf(htmlTemplate, htmlContent)

		// Write the HTML file
		err = ioutil.WriteFile(
			filepath.Join("public", outFilename),
			[]byte(completeHTML),
			0644,
		)
		if err != nil {
			fmt.Printf("Error writing file %s: %v\n", outFilename, err)
			continue
		}

		fmt.Printf("Successfully converted %s to %s\n", file, outFilename)
	}

	// Copy the CSS file to the public directory
	err = ioutil.WriteFile("public/styles.css", []byte(readCSSContent()), 0644)
	if err != nil {
		fmt.Printf("Error writing CSS file: %v\n", err)
		return
	}

	fmt.Println("Website generation complete! Check the 'public' directory for the result.")
}

func markdownToHTML(md []byte) string {
	// Create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// Create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return string(markdown.Render(doc, renderer))
}

func readCSSContent() []byte {
	return []byte(`@import url(https://fonts.googleapis.com/css?family=Roboto:300);
@import url(https://fonts.googleapis.com/css?family=Open+Sans);

html {
  font: 16px/1.5 "Roboto", sans-serif;
}

@media (min-width: 30rem) {
  html {
    font-size: 20px;
  }
}

body {
  margin: 0;
  color: #333;
  background-color: #fff;
}

.container {
  max-width: 42rem;
  margin: 0 auto;
  padding: 2rem 1rem 5rem;
}

a {
  color: #0074d9;
  text-decoration: none;
}

a:hover, a:focus {
  text-decoration: underline;
}

h1, h2, h3, h4, h5, h6 {
  font-family: "Open Sans", sans-serif;
  margin: 0 0 0.5rem -0.1rem;
  line-height: 1;
  color: #111;
  text-rendering: optimizeLegibility;
}

h1 {
  font-size: 2.5rem;
  margin-bottom: 2rem;
  max-width: 30rem;
}

@media (min-width: 30rem) {
  h1 {
    font-size: 3rem;
    margin-bottom: 2rem;
  }
}

h1 a {
  color: inherit;
}

.nav-links {
  margin: 2rem 0 2rem;
  font-weight: bold;
  text-transform: uppercase;
  font-size: 1rem;
}

h2 {
  margin-top: 2rem;
  font-size: 1.25rem;
  margin-bottom: 0.75rem;
}

@media (min-width: 30rem) {
  h2 {
    margin-top: 2.5rem;
    font-size: 1.5rem;
    margin-bottom: 1rem;
  }
}

h3, h4, h5, h6 {
  margin-top: 2rem;
  font-size: 1rem;
  text-transform: uppercase;
  margin-bottom: 0.75rem;
  clear: both; /* Prevents overlapping with content */
}

p, ul, ol, dl, table, pre, blockquote {
  margin-top: 0;
  margin-bottom: 1rem;
}

ul, ol {
  padding-left: 1.5rem;
}

li {
  margin-bottom: 0.75rem;
}

li p {
  margin-bottom: 0.5rem;
}

li:last-child p:last-child {
  margin-bottom: 0;
}

dd {
  margin-left: 1.5rem;
}

blockquote {
  margin-left: 0;
  margin-right: 0;
  padding: .5rem 1rem;
  border-left: .25rem solid #ccc;
  color: #666;
}

blockquote p:last-child {
  margin-bottom: 0;
}

hr {
  border: none;
  margin: 1.5rem 0;
  border-bottom: 1px solid #ccc;
  position: relative;
  top: -1px;
}

img {
  max-width: 100%;
  margin: 0 auto;
  display: block;
}

pre, code {
  font-family: monospace, serif;
  background-color: #f5f5f5;
}

pre {
  padding: .5rem 1rem;
  font-size: 0.8rem;
  white-space: pre-wrap;
}

code {
  padding: .1rem .25rem;
  font-size: 0.85em;
}

/* Fix spacing on section headers and content */
h3 + ul {
  margin-top: 0.75rem;
}

/* Fix for bold text */
strong {
  font-weight: bold;
}

/* Fix for italic text */
em {
  font-style: italic;
}

/* Fix positioning of inline code */
p code, li code {
  display: inline-block;
  vertical-align: baseline;
}

/* Site header specific styling */
h1 {
  margin-top: 1.5rem;
}`)
}
