package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildEmptyDest(t *testing.T) {
	src := t.TempDir()
	out := t.TempDir()
	blog := t.TempDir()

	// create index.html.tmpl in src and dummy css file
	os.WriteFile(filepath.Join(src, "index.html.tmpl"), []byte(`
        <div class="blog">
        	{{.}}
        </div>
	`), 0644)

	os.WriteFile(filepath.Join(src, "index.css"), []byte(`
        .foo {
			display: flex;
		}
	`), 0644)

	// create dummy blog post
	postFp := filepath.Join(blog, "test1.md")
	os.WriteFile(postFp, []byte("## Title\nhey [google](https://google.com)."), 0644)

	if err := build(src, blog, out, buildOptions{Force: false}); err != nil {
		t.Error("failed to build html with empty output directory:", err)
	}

	// check output directory
	outContents, err := os.ReadDir(out)
	if err != nil {
		t.Error(err)
	}

	if len(outContents) != 2 {
		t.Error("expected two files in the output directory, got", len(outContents))
	}

	if outContents[0].Name() != "index.css" {
		t.Error("expected index.css in the output directory, instead got", outContents[0].Name())
	}

	if outContents[1].Name() != "index.html" {
		t.Error("expected index.html in the output directory, instead got", outContents[1].Name())
	}

	indexHtml, err := os.ReadFile(filepath.Join(out, outContents[1].Name()))
	if err != nil {
		t.Error("could not read (assumed) index.html file:", err)
	}

	if !strings.Contains(string(indexHtml), `<a href="https://google.com" target="_blank">`) {
		t.Error("expected file to contain link to google")
	}
}

func TestBuildNonEmptyDestError(t *testing.T) {
	src := t.TempDir()
	out := t.TempDir()
	blog := t.TempDir()

	// create index.html.tmpl in src and dummy css file
	os.WriteFile(filepath.Join(src, "index.html.tmpl"), []byte(`
        <div class="blog">
        	{{.}}
        </div>
	`), 0644)

	os.WriteFile(filepath.Join(src, "index.css"), []byte(`
        .foo {
			display: flex;
		}
	`), 0644)

	// create dummy blog post
	postFp := filepath.Join(blog, "test1.md")
	os.WriteFile(postFp, []byte("## Title\nhey [google](https://google.com)."), 0644)

	// create file in output directory that should trigger error
	os.WriteFile(filepath.Join(out, "index.css"), []byte(`
		.foo {
			display: flex;
		}
	`), 0644)

	err := build(src, blog, out, buildOptions{Force: false})
	if err == nil {
		t.Error("expected error")
		t.FailNow()
	}

	if !strings.Contains(err.Error(), "not empty") {
		t.Errorf("expected error message (%v) to mention 'not empty'", err.Error())
	}
}

func TestBuildNonEmptyDestForce(t *testing.T) {
	src := t.TempDir()
	out := t.TempDir()
	blog := t.TempDir()

	// create index.html.tmpl in src and dummy css file
	os.WriteFile(filepath.Join(src, "index.html.tmpl"), []byte(`
        <div class="blog">
        	{{.}}
        </div>
	`), 0644)

	os.WriteFile(filepath.Join(src, "index.css"), []byte(`
        .foo {
			display: inline-flex;
		}
	`), 0644)

	// create dummy blog post
	postFp := filepath.Join(blog, "test1.md")
	os.WriteFile(postFp, []byte("## Title\nhey [google](https://google.com)."), 0644)

	// create file in output directory
	os.WriteFile(filepath.Join(out, "index.css"), []byte(`
		.foo {
			display: flex;
		}
	`), 0644)

	if err := build(src, blog, out, buildOptions{Force: true}); err != nil {
		t.Error("failed to build html with existing output directory:", err)
	}

	// check output directory
	outContents, err := os.ReadDir(out)
	if err != nil {
		t.Error(err)
	}

	if len(outContents) != 2 {
		t.Error("expected two files in the output directory, got", len(outContents))
	}

	if outContents[0].Name() != "index.css" {
		t.Error("expected index.css in the output directory, instead got", outContents[0].Name())
	}

	if outContents[1].Name() != "index.html" {
		t.Error("expected index.html in the output directory, instead got", outContents[1].Name())
	}

	indexCss, err := os.ReadFile(filepath.Join(out, outContents[0].Name()))
	if err != nil {
		t.Error("could not read index.css file:", err)
	}

	if !strings.Contains(string(indexCss), "display: inline-flex;") {
		t.Error("expected file to be overwritten and contain 'display: inline-flex;', instead got:", string(indexCss))
	}

	indexHtml, err := os.ReadFile(filepath.Join(out, outContents[1].Name()))
	if err != nil {
		t.Error("could not read index.html file:", err)
	}

	if !strings.Contains(string(indexHtml), `<a href="https://google.com" target="_blank">`) {
		t.Error("expected file to contain link to google")
	}
}
