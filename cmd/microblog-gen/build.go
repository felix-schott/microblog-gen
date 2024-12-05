package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	microblog "github.com/felix-schott/microblog-gen/pkg"
	"github.com/yosssi/gohtml"
)

// Recursively creates hard links between the contents of src to the directory dst.
// The function creates dst if necessary.
func overwriteDirectoryContents(src string, dst string, force bool) error {
	stat, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to check if %v is a directory: %v", src, err)
	}
	if !stat.IsDir() {
		return fmt.Errorf("%v is not a directory", src)
	}
	dstContents, err := os.ReadDir(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}
	if len(dstContents) != 0 && !force {
		return fmt.Errorf("directory %v is not empty. use flag force to overwrite", dst)
	}
	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("could not remove directory %v: %v", dst, err)
	}
	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("could not create directory %v: %v", dst, err)
	}
	srcContents, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("could not read the directory %v: %v", src, err)
	}
	for _, f := range srcContents {
		if f.IsDir() {
			if err := overwriteDirectoryContents(filepath.Join(src, f.Name()), filepath.Join(dst, f.Name()), force); err != nil {
				return err
			}
		} else if !strings.HasSuffix(f.Name(), ".tmpl") { // ignore template files, they're handled separately
			from := filepath.Join(src, f.Name())
			to := filepath.Join(dst, f.Name())
			if err := os.Link(from, to); err != nil {
				return fmt.Errorf("could not create a hard link from %v to %v: %v", from, to, err)
			}
		}
	}
	return nil
}

type buildOptions struct {
	Force            bool
	PostTemplateFile string
}

func build(sourceDirectory string, blogDirectory string, outputDirectory string, options buildOptions) error {
	if err := overwriteDirectoryContents(sourceDirectory, outputDirectory, options.Force); err != nil {
		return err
	}

	var blog microblog.Blog
	var err error
	// read blog posts, markdown to html
	if options.PostTemplateFile != "" {
		blog, err = microblog.NewBlog(blogDirectory, microblog.WithTemplateFile(options.PostTemplateFile), microblog.WithPublicationTracking())
	} else {
		blog, err = microblog.NewBlog(blogDirectory, microblog.WithPublicationTracking())
	}
	if err != nil {
		return fmt.Errorf("error when initialising blog: %v", err)
	}

	blogPostsHtml, err := blog.RenderPosts()
	if err != nil {
		return fmt.Errorf("error when trying to render html: %v", err)
	}

	// build index.html
	matches, err := filepath.Glob(filepath.Join(sourceDirectory, "*.html.tmpl"))
	if err != nil {
		return fmt.Errorf("error when trying to match template files: %v", err)
	}
	if len(matches) != 1 {
		return fmt.Errorf("expected exactly 1 template file (*.html.tmpl) in %v, got %v", sourceDirectory, len(matches))
	}
	tmplBytes, err := os.ReadFile(matches[0])
	if err != nil {
		return fmt.Errorf("could not read file %v: %v", matches[0], err)
	}
	tmpl, err := template.New("index.html").Parse(string(tmplBytes))
	if err != nil {
		return fmt.Errorf("could not open template: %v", err)
	}
	if tmpl.Tree == nil {
		return errors.New("template tree is empty")
	}
	outputFp := filepath.Join(outputDirectory, "index.html")
	outputFile, err := os.Create(outputFp)
	if err != nil {
		return fmt.Errorf("could not open output file %v: %v", outputFp, err)
	}
	defer outputFile.Close()

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, string(blogPostsHtml)); err != nil {
		return fmt.Errorf("could not generate output: %v", err)
	}
	formattedHtml := gohtml.FormatBytes(buf.Bytes())
	if _, err := outputFile.Write(formattedHtml); err != nil {
		return fmt.Errorf("could not write formatted html to %v: %v", outputFp, err)
	}

	return nil
}
