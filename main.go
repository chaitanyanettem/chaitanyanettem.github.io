package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type BlogMetadata struct {
	Date  string   `json:"date"`
	Tags  []string `json:"tags"`
	Short string   `json:"short"`
	Slug  string   `json:"slug"`  // Add slug to metadata struct
}

type BlogPost struct {
	Metadata BlogMetadata
	Title    string
	Content  string
	Slug     string
}

type PageData struct {
	Timestamp   int64
	CurrentPage string
	Content     string
	LastUpdated string
	BlogPosts   []BlogPost
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Chaitanya Nettem - Personal Website</title>
    <link rel="stylesheet" href="/styles.css?v={{.Timestamp}}">
</head>
<body>
    <div class="container">
        <header>
            <h1 class="site-title"><a href="/">Chaitanya Nettem</a></h1>
            <nav class="nav-links">
                <a href="/blog/" {{if hasPrefix .CurrentPage "blog/"}}class="active"{{end}}>BLOG</a> /
                <a href="/photography.html" {{if eq .CurrentPage "photography.html"}}class="active"{{end}}>PHOTOGRAPHY</a> /
                <a href="/Chaitanya_Nettem_CV.pdf">RESUME</a> /
                <a href="https://github.com/chaitanyanettem" target="_blank">GITHUB</a> /
                <a href="https://www.linkedin.com/in/cnettem" target="_blank">LINKEDIN</a>
            </nav>
        </header>
        {{.Content}}
        <div class="footer">
            <span class="last-updated">Last updated: {{.LastUpdated}}</span>
            <span class="copyright">© Chaitanya Nettem</span>
        </div>
    </div>
    <script src="/script.js?v={{.Timestamp}}"></script>
</body>
</html>`

func main() {
	// Add the hasPrefix function to the template
	funcMap := template.FuncMap{
		"hasPrefix": strings.HasPrefix,
	}

	// Parse the template with the function map
	tmpl, err := template.New("page").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		fmt.Printf("Error parsing template: %v\n", err)
		return
	}

	// Get current timestamp for cache busting
	timestamp := time.Now().Unix()
	currentDate := time.Now().Format("January 2006")

	// Clean up blog directory
	if err := os.RemoveAll("blog"); err != nil {
		fmt.Printf("Error cleaning blog directory: %v\n", err)
		return
	}

	// Create fresh blog directory
	if err := os.MkdirAll("blog", 0755); err != nil {
		fmt.Printf("Error creating blog directory: %v\n", err)
		return
	}

	// Process blogs
	blogFiles, err := filepath.Glob("content/blogs/*.md")
	if err != nil {
		fmt.Printf("Error finding blog files: %v\n", err)
		return
	}

	fmt.Printf("Found %d blog files\n", len(blogFiles))

	// Process all blog posts
	var blogPosts []BlogPost
	for _, file := range blogFiles {
		fmt.Printf("Processing blog file: %s\n", file)
		post, err := processBlogPost(file)
		if err != nil {
			fmt.Printf("Error processing blog %s: %v\n", file, err)
			continue
		}
		blogPosts = append(blogPosts, post)

		// Create individual blog post page
		data := PageData{
			Timestamp:   timestamp,
			CurrentPage: "blog/" + post.Slug + ".html",
			Content:     fmt.Sprintf(`
				<div class="back-link">
					<a href="/blog/">← Back to All Blogs</a>
				</div>
				<h1>%s</h1>
				<div class="post-meta">
					<span class="post-date">%s</span>
				</div>
				%s`, 
				post.Title,
				post.Metadata.Date,
				post.Content),
			LastUpdated: currentDate,
		}

		outPath := filepath.Join("blog", post.Slug+".html")
		fmt.Printf("Generating blog post: %s\n", outPath)
		
		outFile, err := os.Create(outPath)
		if err != nil {
			fmt.Printf("Error creating blog file: %v\n", err)
			continue
		}

		if err := tmpl.Execute(outFile, data); err != nil {
			fmt.Printf("Error executing template: %v\n", err)
		}
		outFile.Close()
		
		fmt.Printf("Successfully generated blog post: %s\n", outPath)
	}

	// Sort blog posts by date
	sort.Slice(blogPosts, func(i, j int) bool {
		return blogPosts[i].Metadata.Date > blogPosts[j].Metadata.Date
	})

	// Create blog index page
	blogIndexContent := strings.Builder{}
	for _, post := range blogPosts {
		// Put date in the title line
		blogIndexContent.WriteString(fmt.Sprintf("### [%s](./%s.html) <span class=\"post-date\">%s</span>\n",
			post.Title,
			post.Slug,
			post.Metadata.Date))
		if post.Metadata.Short != "" {
			blogIndexContent.WriteString(post.Metadata.Short + "\n\n")
		}
	}

	data := PageData{
		Timestamp:   timestamp,
		CurrentPage: "blog/index.html",
		Content:     markdownToHTML([]byte(blogIndexContent.String())),
		LastUpdated: currentDate,
	}

	outFile, err := os.Create(filepath.Join("blog", "index.html"))
	if err != nil {
		fmt.Printf("Error creating blog index: %v\n", err)
		return
	}
	if err := tmpl.Execute(outFile, data); err != nil {
		fmt.Printf("Error executing template: %v\n", err)
	}
	outFile.Close()

	// Process regular content files
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
			Timestamp:   timestamp,
			CurrentPage: outFilename,
			Content:     markdownToHTML(mdContent),
			LastUpdated: currentDate,
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

func processBlogPost(filename string) (BlogPost, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return BlogPost{}, err
	}

	// Split content into metadata and markdown
	parts := strings.SplitN(string(content), "\n\n", 2)
	if len(parts) != 2 {
		return BlogPost{}, fmt.Errorf("invalid blog format")
	}

	// Parse metadata
	var metadata BlogMetadata
	if err := json.Unmarshal([]byte(parts[0]), &metadata); err != nil {
		return BlogPost{}, err
	}

	// Extract title from first heading
	titleRe := regexp.MustCompile(`(?m)^#\s+(.+)$`)
	matches := titleRe.FindStringSubmatch(parts[1])
	if len(matches) < 2 {
		return BlogPost{}, fmt.Errorf("no title found in content")
	}
	title := matches[1]

	// Remove the title from the content and convert to HTML
	contentWithoutTitle := titleRe.ReplaceAllString(parts[1], "")
	
	// Use metadata slug if available, otherwise derive from filename
	var slug string
	if metadata.Slug != "" {
		slug = metadata.Slug
	} else {
		// Remove .md extension and use filename
		base := filepath.Base(filename)
		slug = strings.TrimSuffix(base, ".md")
	}
	
	return BlogPost{
		Metadata: metadata,
		Title:    title,
		Content:  markdownToHTML([]byte(contentWithoutTitle)),
		Slug:     slug,
	}, nil
}
