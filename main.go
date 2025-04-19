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
	Slug  string   `json:"slug"`
}

type BlogPost struct {
	Metadata BlogMetadata
	Title    string
	Content  string
	RawContent string // Add this field
	Slug     string
}

type PageData struct {
	Timestamp   int64
	CurrentPage string
	Content     string
	LastUpdated string
	BlogPosts   []BlogPost
	Is404       bool
	MetaDesc    string
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{if .Is404}}404 - Page Not Found | {{end}}Chaitanya Nettem</title>
    
    <!-- Preload critical assets -->
    <link rel="preload" href="/styles.css?v={{.Timestamp}}" as="style">
    <link rel="preload" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css" as="style">
    
    <!-- Preconnect to external domains -->
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link rel="preconnect" href="https://cdnjs.cloudflare.com">
    
    <!-- Add meta description for SEO -->
    <meta name="description" content="{{.MetaDesc}}">
    
    <!-- Prevent FOUC -->
    <script>
        let FF_FOUC_FIX;
    </script>
    
    <!-- Add Prism.js CSS before your main stylesheet -->
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/themes/prism.min.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/plugins/toolbar/prism-toolbar.min.css">
    
    <!-- Load styles -->
    <link rel="stylesheet" href="/styles.css?v={{.Timestamp}}">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css">
    
    <!-- Add Prism.js and its plugins after your main script -->
    <script defer src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-core.min.js"></script>
    <script defer src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/plugins/autoloader/prism-autoloader.min.js"></script>
    <script defer src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/plugins/toolbar/prism-toolbar.min.js"></script>
    <script defer src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/plugins/copy-to-clipboard/prism-copy-to-clipboard.min.js"></script>
    
    <!-- Defer non-critical JavaScript -->
    <script defer src="/script.js?v={{.Timestamp}}"></script>
</head>
<body>
    <div class="container">
        <header>
            <h1 class="site-title"><a href="/">Chaitanya Nettem</a></h1>
            <div class="nav-container">
                <nav class="nav-links">
                    <a href="/blog/" {{if hasPrefix .CurrentPage "blog/"}}class="active"{{end}}>BLOG</a> /
                    <a href="/photography.html" {{if eq .CurrentPage "photography.html"}}class="active"{{end}}>PHOTOGRAPHY</a> /
                    <a href="/Chaitanya_Nettem_CV.pdf">RESUME</a>
                </nav>
                <div class="social-links">
                    <a href="https://github.com/chaitanyanettem" target="_blank" rel="noopener" title="GitHub"><i class="fab fa-github"></i></a>
                    <a href="https://www.linkedin.com/in/cnettem" target="_blank" rel="noopener" title="LinkedIn"><i class="fab fa-linkedin"></i></a>
                </div>
            </div>
        </header>
        {{.Content}}
        <div class="footer">
            <span class="last-updated">Last updated: {{.LastUpdated}}</span>
            <span class="copyright">© Chaitanya Nettem</span>
        </div>
    </div>
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
			MetaDesc:    post.Metadata.Short, // Just use the short description directly
			Content: fmt.Sprintf(`
				<div class="back-link">
					<a href="/blog/">Back to All Blogs</a>
				</div>
				<h1>%s</h1>
				<div class="post-meta">
					<span class="post-date">%s</span>
					<span class="reading-time">%d minute read</span>
				</div>
				%s`,
				post.Title,
				formatBlogDate(post.Metadata.Date),
				calculateReadingTime(post.RawContent),
				post.Content),
			LastUpdated: getLastModifiedDate(file),
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

	var lastUpdated string
	if len(blogFiles) > 0 {
		// Use the most recent blog file modification as the index last updated date
		latest := time.Time{}
		for _, file := range blogFiles {
			info, err := os.Stat(file)
			if err == nil && info.ModTime().After(latest) {
				latest = info.ModTime()
			}
		}
		lastUpdated = latest.Format("January 2006")
	} else {
		lastUpdated = time.Now().Format("January 2006")
	}

	data := PageData{
		Timestamp:   timestamp,
		CurrentPage: "blog/index.html",
		MetaDesc:    "Technical articles on software engineering, distributed systems, and programming by Chaitanya Nettem, Software Engineer at Rubrik.",
		Content:     markdownToHTML([]byte(blogIndexContent.String())),
		LastUpdated: lastUpdated,
		BlogPosts:   blogPosts,
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
			MetaDesc:    "Personal website of Chaitanya Nettem. Writing about distributed systems, software architecture, and sharing photography.",
			Content:     markdownToHTML(mdContent),
			LastUpdated: getLastModifiedDate(file),
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

	// Generate 404 page
	errorContent, err := os.ReadFile("content/404.md")
	if err != nil {
		fmt.Printf("Error reading error.md: %v\n", err)
		return
	}

	// Convert markdown to HTML and wrap in error-container
	htmlContent := fmt.Sprintf(`<div class="error-container">%s</div>`, markdownToHTML(errorContent))

	data404 := PageData{
		Timestamp:   timestamp,
		CurrentPage: "404.html",
		Content:     htmlContent,
		LastUpdated: time.Now().Format("January 2006"),
		Is404:       true,
	}

	// Create 404 page
	outFile, err = os.Create("404.html")
	if err != nil {
		fmt.Printf("Error creating 404 page: %v\n", err)
		return
	}
	if err := tmpl.Execute(outFile, data404); err != nil {
		fmt.Printf("Error executing template for 404 page: %v\n", err)
	}
	outFile.Close()

	fmt.Println("Website generation complete!")
}

func markdownToHTML(md []byte) string {
	content := string(md)

	// Handle code blocks with language specification
	codeBlockRe := regexp.MustCompile("```([a-zA-Z0-9]+)\n([^`]+)```")
	content = codeBlockRe.ReplaceAllString(content, `<div class="code-block-wrapper">
		<pre><code class="language-$1">$2</code></pre>
		<button class="expand-code" aria-label="Expand code">
			<span class="expand-text">Show more</span>
			<span class="collapse-text">Show less</span>
		</button>
	</div>`)

	// Handle code blocks without language specification
	plainCodeBlockRe := regexp.MustCompile("```\n([^`]+)```")
	content = plainCodeBlockRe.ReplaceAllString(content, `<div class="code-block-wrapper">
		<pre><code class="language-plaintext">$1</code></pre>
		<button class="expand-code" aria-label="Expand code">
			<span class="expand-text">Show more</span>
			<span class="collapse-text">Show less</span>
		</button>
	</div>`)

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

	// Remove the title from the content
	contentWithoutTitle := titleRe.ReplaceAllString(parts[1], "")

	return BlogPost{
		Metadata: metadata,
		Title:    title,
		Content:  markdownToHTML([]byte(contentWithoutTitle)),
		RawContent: contentWithoutTitle, // Store the raw content
		Slug:     metadata.Slug,
	}, nil
}

func getLastModifiedDate(filepath string) string {
	info, err := os.Stat(filepath)
	if err != nil {
		return time.Now().Format("January 2006")
	}
	return info.ModTime().Format("January 2006")
}

// Add this function to calculate reading time
func calculateReadingTime(content string) int {
	// Average reading speed: 200 words per minute
	words := len(strings.Fields(content))
	minutes := (words + 199) / 200 // Round up to nearest minute
	if minutes < 1 {
		minutes = 1
	}
	return minutes
}

// Add this new function to format the date
func formatBlogDate(dateStr string) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("January 2, 2006")
}
