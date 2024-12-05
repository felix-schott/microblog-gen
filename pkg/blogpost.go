package microblog

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

var htmlRenderer *html.Renderer

type defaultOptions struct {
	Template                  string
	Backend                   RegistryType
	EnablePublicationTracking bool
}

// Default options used in BlogPost.WriteHtml(). You can pass options `WithTemplateFile`/`WithTemplateString` and
// `WithPublicationTracking` to Blog and BlogPost to override the default behaviour. If you wish, you can also manipulate
// this struct directly to change behaviour globally.
var DefaultOptions defaultOptions = defaultOptions{
	Template: `
	<div class="blog-post">
		<h2>{{.Heading}}</h2>
		<span class="dt-posted">{{.DtPosted}}</span>
		{{.Content}}
	</div>
	`,
	Backend:                   SQLite,
	EnablePublicationTracking: false,
}

// Represents a blog post in the form of a Markdown file. Exposes methods to render its contents as HTML.
type BlogPost interface {
	GetFilePath() string
	GetName() string
	Markdown() (string, error)
	WriteHtml(io.Writer) error
}

type blogPost struct {
	FilePath            string
	publicationTracking bool
	PublicationDate     *time.Time
	template            string
	Heading             string
	Content             string
	DtPosted            string
}

// Create a new BlogPost
// The first argument is the filepath to the Markdown file that contains the blog post.
// You can use the following options to configure the behaviour further:
//
//	blogPost, err := microblog.NewBlogPost("/path/to/post.md")) // use default template
//	blogPost, err := microblog.NewBlogPost("/path/to/post.md", microblog.WithTemplateString(`<html>...</html>`))
//	blogPost, err := microblog.NewBlogPost("/path/to/post.md", microblog.WithTemplateFile("/path/to/html/template"))
//
// The template itself must be written using the Golang template language and contain the same variables as the
// default template. You can however change the tags and classes.
// You can access the default template at microblog.Template.
func NewBlogPost(fp string, options ...func(*blogPost) error) (BlogPost, error) {
	if _, err := os.Stat(fp); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("file %v does not exist", fp)
		}
		return nil, fmt.Errorf("could not acquire file info for %v: %v", fp, err)
	}
	b := &blogPost{FilePath: fp}
	b.template = DefaultOptions.Template
	for _, o := range options {
		if err := o(b); err != nil {
			return nil, err
		}
	}
	return b, nil
}

// options

// Override the default template (microblog.DefaultOptions.Template) with a string
func WithTemplateString(s string) func(*blogPost) error {
	return func(b *blogPost) error {
		b.template = s
		return nil
	}
}

// Override the default template (microblog.DefaultOptions.Template) with a template file
func WithTemplateFile(fp string) func(*blogPost) error {
	return func(b *blogPost) error {
		content, err := os.ReadFile(fp)
		if err != nil {
			return err
		}
		b.template = string(content)
		return nil
	}
}

// Enable publication tracking using a database backend.
// With this option enabled, the date of the first blogpost.WriteHtml call
// will be stored in a database backend and used in subsequent blogpost.WriteHtml calls
func WithPublicationTracking() func(*blogPost) error {
	return func(b *blogPost) error {
		b.publicationTracking = true
		return nil
	}
}

// methods

// Returns the file path of the markdown file that the BlogPost represents.
func (p *blogPost) GetFilePath() string {
	return p.FilePath
}

// Returns the file name of the markdown file that the BlogPost represents. This is used as
// a unique identifier in the storage backend.
func (p *blogPost) GetName() string {
	return filepath.Base(p.GetFilePath())
}

// Returns the plain file content of the .md file that the BlogPost represents.
func (p *blogPost) Markdown() (string, error) {
	file, err := os.ReadFile(p.FilePath)
	if err != nil {
		return "", fmt.Errorf("could not read file %v: %v", p.FilePath, err)
	}
	return string(file), nil
}

func (p *blogPost) String() string {
	c, err := p.Markdown()
	if err != nil {
		panic(err)
	}
	return c
}

// Renders BlogPost as HTML. Method accepts the io.Writer interface which means you can write to a bytes.Buffer
// and all other structs that implement the io.Writer interface,
// for example:
//
//	var buf bytes.Buffer
//	err := post.WriteHtml(&buf)
//
// To change the underlying HTML template, use the `WithTemplateFile`/`WithTemplateString` options to `NewBlog()`
// or `NewBlogPost()`. To control whether the publication date rendered is fetched from/written to a database backend,
// use the `WithPublicationTracking()` option.
func (p *blogPost) WriteHtml(w io.Writer) error {
	file, err := os.ReadFile(p.FilePath)
	if err != nil {
		return fmt.Errorf("could not read file %v: %v", p.FilePath, err)
	}
	markdownParser := parser.New()
	if htmlRenderer == nil {
		htmlFlags := html.CommonFlags | html.HrefTargetBlank
		opts := html.RendererOptions{Flags: htmlFlags}
		htmlRenderer = html.NewRenderer(opts)
	}

	if markdownParser == nil {
		log.Fatal("parser is nil")
	}
	doc := markdownParser.Parse(file)

	nodes := doc.GetChildren()
	var heading *ast.Heading
	var paragraphs []*ast.Paragraph = make([]*ast.Paragraph, 0, len(nodes))
	for _, c := range nodes {
		switch n := c.(type) {
		case *ast.Heading:
			if heading != nil {
				return errors.New("more than one heading in blog post")
			}
			heading = n
		case *ast.Paragraph:
			paragraphs = append(paragraphs, n)
		default:
			return fmt.Errorf("unexpected node of type %T", c)
		}
	}
	if heading == nil {
		return errors.New("no heading found")
	}
	r := strings.NewReplacer(
		"<h2>", "",
		"</h2>", "",
		"<h3>", "",
		"</h3>", "",
	)
	p.Heading = r.Replace(string(markdown.Render(heading, htmlRenderer)))
	if len(paragraphs) == 0 {
		return errors.New("no paragraphs in blog post")
	}
	var s strings.Builder
	for idx := range paragraphs {
		s.WriteString(string(markdown.Render(paragraphs[idx], htmlRenderer)))
	}
	p.Content = s.String()

	var dtPosted *time.Time

	// add publication date
	if DefaultOptions.EnablePublicationTracking || p.publicationTracking {
		switch DefaultOptions.Backend {
		case SQLite:
			backend, err := (&sqlitePool).Acquire(filepath.Dir(p.GetFilePath()))
			if err != nil {
				return err
			}
			dtPosted, err = backend.GetPublicationDate(p)
			if err != nil {
				return err
			}
			if dtPosted == nil { // hasn't been published before
				dtPosted, err = backend.SetPublicationDate(p, nil)
				if err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unrecognised backend option %v", DefaultOptions.Backend)
		}
	} else {
		today := time.Now()
		dtPosted = &today
	}
	p.DtPosted = dtPosted.Format(time.DateOnly)

	tmpl, err := template.New("post").Parse(p.template)
	if err != nil {
		return fmt.Errorf("could not open template: %v", err)
	}
	if tmpl.Tree == nil {
		return errors.New("template tree is empty")
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, p); err != nil {
		return fmt.Errorf("could not generate output: %v", err)
	}
	if _, err := w.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}
