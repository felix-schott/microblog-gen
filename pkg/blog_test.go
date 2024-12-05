package microblog

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestBlogRenderPosts(t *testing.T) {
	d := t.TempDir()

	// write dummy blog posts
	postFp1 := filepath.Join(d, "test1.md")
	os.WriteFile(postFp1, []byte("## Title\nhey [google](https://google.com)."), 0644)

	postFp2 := filepath.Join(d, "test2.md")
	os.WriteFile(postFp2, []byte("## Title\nhello [google](https://google.com)."), 0644)

	blog, err := NewBlog(d)
	if err != nil {
		t.Errorf("could not instantiate blog object: %v", err)
		t.FailNow()
	}
	if blog == nil {
		t.Error("blog object is null")
		t.FailNow()
	}

	if blog.GetDirectory() != d {
		t.Errorf("expected blog directory to be %v, got %v", d, blog.GetDirectory())
	}

	posts, err := blog.RenderPosts()
	if err != nil {
		t.Error("could not render posts:", err)
	}

	expHtml := regexp.MustCompile(`(<div class="blog-post">\s*<h2>\s*Title\s*</h2>\s*<span class="dt-posted">` + time.Now().Format(time.DateOnly) + `</span>\s*<p>\w+ <a href="https://google.com" target="_blank">google</a>\.</p>\s*</div>\s*){2}`)

	if !expHtml.Match(posts) || !strings.Contains(string(posts), "hello") || !strings.Contains(string(posts), "hey") {
		t.Errorf("html (%s) does not match", posts)
	}
}

func TestBlogWithTemplateStringRenderPosts(t *testing.T) {
	d := t.TempDir()

	// write dummy blog posts
	postFp1 := filepath.Join(d, "test1.md")
	os.WriteFile(postFp1, []byte("## Title\nhey [google](https://google.com)."), 0644)

	postFp2 := filepath.Join(d, "test2.md")
	os.WriteFile(postFp2, []byte("## Title\nhello [google](https://google.com)."), 0644)

	// h3 instead of h2
	blog, err := NewBlog(d, WithTemplateString(`
	<div class="blog-post">
		<h3>{{.Heading}}</h3>
		<span class="dt-posted">{{.DtPosted}}</span>
		{{.Content}}
	</div>
	`))
	if err != nil {
		t.Errorf("could not instantiate blog object: %v", err)
		t.FailNow()
	}
	if blog == nil {
		t.Error("blog object is null")
		t.FailNow()
	}

	if blog.GetDirectory() != d {
		t.Errorf("expected blog directory to be %v, got %v", d, blog.GetDirectory())
	}

	posts, err := blog.RenderPosts()
	if err != nil {
		t.Error("could not render posts:", err)
	}

	expHtml := regexp.MustCompile(`(<div class="blog-post">\s*<h3>\s*Title\s*</h3>\s*<span class="dt-posted">` + time.Now().Format(time.DateOnly) + `</span>\s*<p>\w+ <a href="https://google.com" target="_blank">google</a>\.</p>\s*</div>\s*){2}`)

	if !expHtml.Match(posts) || !strings.Contains(string(posts), "hello") || !strings.Contains(string(posts), "hey") {
		t.Errorf("html (%s) does not match", posts)
	}
}

func BenchmarkRenderPost(b *testing.B) {
	d := b.TempDir()

	var regexString strings.Builder

	// write dummy blog posts
	for i := range 100 {
		postFp := filepath.Join(d, fmt.Sprintf("test%03d.md", i))
		os.WriteFile(postFp, []byte(fmt.Sprintf("## Title %v\nhey [google](https://google.com).", i)), 0644)
		regexString.WriteString(fmt.Sprintf(`(.|\s)*Title %v`, i))
	}

	blog, err := NewBlog(d)
	if err != nil {
		b.Error(err)
		b.FailNow()
	}

	r := regexp.MustCompile(regexString.String())

	b.ResetTimer()
	for range b.N {
		html, err := blog.RenderPosts()
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
		if !r.Match(html) {
			b.Error("regex doesn't match")
			os.WriteFile("test.txt", html, 0644)
			b.FailNow()
		}
	}
}

func BenchmarkRenderPostAsync(b *testing.B) {
	d := b.TempDir()

	// write dummy blog posts
	var regexString strings.Builder
	for i := range 100 {
		postFp := filepath.Join(d, fmt.Sprintf("test%03d.md", i))
		os.WriteFile(postFp, []byte(fmt.Sprintf("## Title %v\nhey [google](https://google.com).", i)), 0644)
		regexString.WriteString(fmt.Sprintf(`(.|\s)*Title %v`, i))
	}

	blog, err := NewBlog(d)
	if err != nil {
		b.Error(err)
		b.FailNow()
	}

	r := regexp.MustCompile(regexString.String())

	b.ResetTimer()
	for range b.N {
		html, err := blog.RenderPostsAsync()
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
		if !r.Match(html) {
			b.Error("regex doesn't match")
			b.FailNow()
		}
	}
}
