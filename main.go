package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

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
    <link rel="stylesheet" href="styles.css?v=%d">
</head>
<body>
    <div class="container">
        %s
        <div class="footer">Last updated: %s</div>
    </div>
</body>
</html>`

func main() {
	// Get current timestamp for cache busting
	timestamp := time.Now().Unix()

	// Read all markdown files from the content directory
	files, err := filepath.Glob("content/*.md")
	if err != nil {
		fmt.Printf("Error finding markdown files: %v\n", err)
		return
	}

	for _, file := range files {
		// Read the markdown file
		mdContent, err := os.ReadFile(file)
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

		// Get current date in the desired format
		currentDate := time.Now().Format("January 2006")

		// Insert the HTML content, timestamp, and date into the template
		completeHTML := fmt.Sprintf(
			htmlTemplate,
			timestamp,
			htmlContent,
			currentDate,
		)

		// Write the HTML file
		err = os.WriteFile(
			outFilename,
			[]byte(completeHTML),
			0644,
		)
		if err != nil {
			fmt.Printf("Error writing file %s: %v\n", outFilename, err)
			continue
		}

		fmt.Printf("Successfully converted %s to %s\n", file, outFilename)
	}

	fmt.Println("Website generation complete!")
}

func markdownToHTML(md []byte) string {
	// First, handle our custom tooltip syntax
	content := string(md)
	re := regexp.MustCompile(`\[([^\]]+)\]{tooltip="([^"]+)"}`)
	content = re.ReplaceAllString(content, `<span tooltip="$2">$1</span>`)

	// Create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(content))

	// Create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return string(markdown.Render(doc, renderer))
}
