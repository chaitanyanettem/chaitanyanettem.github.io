package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// Simplified PageData struct - removed ContainerClass as it's no longer used
type PageData struct {
	Timestamp   int64
	CurrentPage string
	Content     string
	LastUpdated string
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Chaitanya Nettem - Personal Website</title>
    <link rel="stylesheet" href="styles.css?v={{.Timestamp}}">
</head>
<body>
    <div class="container">
        <div class="nav-links">
            <a href="./"{{if eq .CurrentPage "index.html"}} class="current"{{end}}>HOME</a> &nbsp;|&nbsp;
            <a href="./blog/"{{if eq .CurrentPage "blog/index.html"}} class="current"{{end}}>BLOG</a> &nbsp;|&nbsp; 
            <a href="./photography.html"{{if eq .CurrentPage "photography.html"}} class="current"{{end}}>PHOTOGRAPHY</a> &nbsp;|&nbsp; 
            <a href="./Chaitanya_Nettem_CV.pdf">RESUME</a> &nbsp;|&nbsp; 
            <a href="https://github.com/chaitanyanettem" target="_blank">GITHUB</a> &nbsp;|&nbsp; 
            <a href="https://www.linkedin.com/in/cnettem" target="_blank">LINKEDIN</a>
        </div>
        {{.Content}}
        <div class="footer">Last updated: {{.LastUpdated}}</div>
    </div>
    <script src="script.js?v={{.Timestamp}}"></script>
</body>
</html>`

func main() {
	// Parse the template
	tmpl, err := template.New("page").Parse(htmlTemplate)
	if err != nil {
		fmt.Printf("Error parsing template: %v\n", err)
		return
	}

	// Get current timestamp for cache busting
	timestamp := time.Now().Unix()
	currentDate := time.Now().Format("January 2006")

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

		outFilename := strings.TrimSuffix(filepath.Base(file), ".md") + ".html"
		
		data := PageData{
			Timestamp:     timestamp,
			CurrentPage:   outFilename,
			Content:       markdownToHTML(mdContent),
			LastUpdated:   currentDate,
		}

		outFile, err := os.Create(outFilename)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", outFilename, err)
			continue
		}

		if err := tmpl.Execute(outFile, data); err != nil {
			fmt.Printf("Error executing template for %s: %v\n", outFilename, err)
		}
		outFile.Close()

		fmt.Printf("Successfully converted %s to %s\n", file, outFilename)
	}

	fmt.Println("Website generation complete!")
}

func markdownToHTML(md []byte) string {
	content := string(md)
	
	// Handle tooltips
	tooltipRe := regexp.MustCompile(`\[([^\]]+)\]{tooltip="([^"]+)"}`)
	content = tooltipRe.ReplaceAllString(content, `<span tooltip="$2">$1</span>`)

	// Handle photo gallery
	photoRe := regexp.MustCompile(`!\[@photo="photos/([^"]+)" caption="([^"]+)"\]`)
	photoMatches := photoRe.FindAllStringSubmatch(content, -1)
	
	if len(photoMatches) > 0 {
		photoHTML := `<div class="photo-grid">`
		for _, match := range photoMatches {
			photoHTML += fmt.Sprintf(`<div class="photo-item">
				<a href="./content/photos/%s" class="lightbox">
					<img src="./content/photos/%s" alt="%s">
					<p class="photo-caption">%s</p>
				</a>
			</div>`, match[1], match[1], match[2], match[2])
		}
		photoHTML += `</div>`
		
		// Replace all photo markdowns with empty string
		content = photoRe.ReplaceAllString(content, "")
		
		// Create markdown parser with extensions
		extensions := parser.CommonExtensions | parser.AutoHeadingIDs
		p := parser.NewWithExtensions(extensions)
		doc := p.Parse([]byte(content))

		// Create HTML renderer with extensions
		htmlFlags := html.CommonFlags | html.HrefTargetBlank
		opts := html.RendererOptions{Flags: htmlFlags}
		renderer := html.NewRenderer(opts)

		// Convert markdown to HTML first
		htmlContent := string(markdown.Render(doc, renderer))
		
		// Insert the photo grid after the first paragraph
		firstParaEnd := strings.Index(htmlContent, "</p>")
		if firstParaEnd != -1 {
			firstParaEnd += 4 // length of "</p>"
			htmlContent = htmlContent[:firstParaEnd] + "\n" + photoHTML + htmlContent[firstParaEnd:]
		} else {
			// If no paragraph found, append to the end
			htmlContent += "\n" + photoHTML
		}
		
		return htmlContent
	}

	// If no photos, just render the markdown normally
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(content))

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return string(markdown.Render(doc, renderer))
}
