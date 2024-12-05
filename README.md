## Generate HTML blog posts from Markdown

`microblog-gen` is a very simple static site generator that takes a directory with blog posts in markdown format, renders them as HTML and inserts them into a HTML template for your single-page website. Similar things can be achieved with a variety of other tools but I had fun building this for my own website&mdash;others might find it useful too.

In detail, this repository exposes 
1. a command-line tool `microblog-gen` for generating a full HTML page from an input template and blog entries as Markdown files (source code in `/cmd`)
2. a Go library with utility functions for transforming Markdown files into HTML microblog posts (source code in `/pkg`)

### Features

This tool is a thin layer around a markdown2html parser. It automatically adds a publication date to the generated HTML. A local Sqlite database is used to track whether a post has been published before (i.e. whether it has been included in a build before).

### Installation

#### CLI
You have several options to install the CLI binary.

1. Using go tooling: `go install github.com/felix-schott/microblog-gen/cmd@latest`
2. By going to the [release page](https://github.com/felix-schott/microblog-gen/releases) and downloading the latest release for your system

#### Library
To use as a Go library, simply run `go get github.com/felix-schott/microblog-gen`.

### Usage

#### CLI
```
$ microblog-gen -h

Usage of ./microblog-gen:
  -b string
        Directory that contains blog posts as Markdown files. (default "./blog")
  -f    Overwrite output directory contents.
  -i string
        Source directory with HTML template and other assets. (default "./src")
  -o string
        Output directory for generated files. (default "./build")
  -t string
        Path to a HTML template for generated blog posts
```

##### Workflow
1. Create a source folder (by default the tool looks for `src`) with all your CSS, JS and an `index.html.tmpl` file.
2. The `index.html.tmpl` file must contain valid HTML and a placeholder `{{.}}` for where you want to insert the generated blog posts.
3. Style the generate blog posts using CSS selectors in `index.css`. Check [Overwriting the default template](#template) to see which selectors you can use. You can also change the template and use custom class names.
4. Create a directory for markdown blog posts (by default the tool looks for `blog`) and add a blog post. The order of the files in the directory matters, prefix a number to the file name to order files, e.g. `001_first_post.md`. The name of the files itself don't get used and are meant to be purely descriptive.
5. Run the `microblog-gen` command in the root directory of the project, specifying flags as needed for non-default directory names.
6. Serve the build directory using your favourite web server.

#### Library

**Render markdown blog entries as HTML**

```go
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"

	microblog "github.com/felix-schott/microblog-gen/pkg"
)

func main() {
	blogDirectory := "~/blog"
	if err := os.WriteFile(filepath.Join(blogDirectory, "/HelloWorld.md"), []byte(`
	## Hello World!

	This is my first blog post on [example.com](http://example.com).
	`), 0644); err != nil {
		log.Fatal(err)
	}

	blog, err := microblog.NewBlog(blogDirectory)
	if err != nil {
		log.Fatal(err)
	}

	// get individual blog posts
	for _, post := range blog.GetBlogPosts() {

		// render html
		var buf bytes.Buffer
		if err := post.WriteHtml(&buf); err != nil {
			log.Fatal(err)
		}

		fmt.Println("html:", buf.String())
	}

	// render all posts
	allPostsHTML, err := blog.RenderPosts()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("all posts:", string(allPostsHTML))
}
```

#### <a name="template"></a> Overwriting the default template
The default template is defined at `microblog.DefaultOptions.Template`. When using the library mode, you can overwrite this struct field to apply changes globally. When using the CLI, you can pass the path to a template file using the `-t` flag. When modifying the template, make sure you keep the same variables.

```
<div class="blog-post">
	<h2>{{.Heading}}</h2>
	<span class="dt-posted">{{.DtPosted}}</span>
	{{.Content}}
</div>
```