package microblog

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

// func TestBlogpostGetPublicationDateRerun(t *testing.T) {
// 	d := t.TempDir()

// 	// write dummy blog post
// 	postFp := filepath.Join(d, "test.md")
// 	os.WriteFile(postFp, []byte("hey"), 0644)

// 	// what happens if this blog post has already been published?
// 	// let's simulate a .published entry for above file
// 	os.WriteFile(filepath.Join(d, ".published"), []byte(fmt.Sprintf("test.md: %v\n", time.Now().AddDate(0, 0, -1).Format(time.DateOnly))), 0644)

// 	post, err := NewBlogPost(postFp)
// 	if err != nil {
// 		t.Errorf("could not instantiate post object: %v", err)
// 	}
// 	dt, err := post.GetPublicationDate()
// 	if err != nil {
// 		t.Errorf("could not get publication date: %v", err)
// 	}
// 	if dt == nil {
// 		t.Errorf("dt is nil")
// 		t.FailNow()
// 	}
// 	y1, m1, d1 := dt.Date()
// 	y2, m2, d2 := time.Now().AddDate(0, 0, -1).Date()
// 	if y1 != y2 || m1 != m2 || d1 != d2 {
// 		t.Errorf("expected yesterday's date to be returned, instead got %v", dt)
// 	}
// 	invFileB, err := os.ReadFile(filepath.Join(d, ".published"))
// 	if err != nil {
// 		t.Errorf("error reading file: %v", err)
// 		t.FailNow()
// 	}
// 	lines := strings.Split(string(invFileB), "\n")
// 	nonEmptyLines := make([]string, 0, len(lines))
// 	for _, l := range lines {
// 		if l != "" {
// 			nonEmptyLines = append(nonEmptyLines, l)
// 		}
// 	}
// 	if len(nonEmptyLines) != 1 {
// 		t.Errorf("expected exactly 1 line in .published, got %v", len(nonEmptyLines))
// 	}
// }

func TestBlogpostMarkdown(t *testing.T) {
	d := t.TempDir()

	// write dummy blog post
	postFp := filepath.Join(d, "test.md")
	content := "## Title\nhey [google](https://google.com)."
	os.WriteFile(postFp, []byte(content), 0644)

	post, err := NewBlogPost(postFp)
	if err != nil {
		t.Errorf("could not instantiate post object: %v", err)
	}

	md, err := post.Markdown()
	if err != nil {
		t.Error("could not retrieve markdown:", err)
	}
	if md != content {
		t.Errorf("md content (%v) does not match '%v'", md, content)
	}
}

func TestBlogpostHtml(t *testing.T) {
	d := t.TempDir()

	// write dummy blog post
	postFp := filepath.Join(d, "test.md")
	content := "## Title\nhey [google](https://google.com)."
	os.WriteFile(postFp, []byte(content), 0644)

	post, err := NewBlogPost(postFp)
	if err != nil {
		t.Errorf("could not instantiate post object: %v", err)
	}

	var html bytes.Buffer
	if err := post.WriteHtml(&html); err != nil {
		t.Error("could not retrieve markdown:", err)
	}
	expHtml := regexp.MustCompile(`<div class="blog-post">\s*<h2>\s*Title\s*</h2>\s*<span class="dt-posted">` + time.Now().Format(time.DateOnly) + `</span>\s*<p>hey <a href="https://google.com" target="_blank">google</a>\.</p>\s*</div>`)
	if !expHtml.Match(html.Bytes()) {
		t.Errorf("generated html (%s) does not match '%v'", html.Bytes(), expHtml)
	}
}

func TestBlogpostWithTemplateString(t *testing.T) {
	d := t.TempDir()

	// write dummy blog post
	postFp := filepath.Join(d, "test.md")
	content := "## Title\nhey [google](https://google.com)."
	os.WriteFile(postFp, []byte(content), 0644)

	// h3 instead of h2
	post, err := NewBlogPost(postFp, WithTemplateString(`
	<div class="blog-post">
		<h3>{{.Heading}}</h3>
		<span class="dt-posted">{{.DtPosted}}</span>
		{{.Content}}
	</div>
	`))
	if err != nil {
		t.Errorf("could not instantiate post object: %v", err)
	}

	var html bytes.Buffer
	if err := post.WriteHtml(&html); err != nil {
		t.Error("could not retrieve markdown:", err)
	}
	expHtml := regexp.MustCompile(`<div class="blog-post">\s*<h3>\s*Title\s*</h3>\s*<span class="dt-posted">` + time.Now().Format(time.DateOnly) + `</span>\s*<p>hey <a href="https://google.com" target="_blank">google</a>\.</p>\s*</div>`)
	if !expHtml.Match(html.Bytes()) {
		t.Errorf("generated html (%s) does not match '%v'", html.Bytes(), expHtml)
	}
}

func TestBlogpostWithTemplateFile(t *testing.T) {
	d := t.TempDir()

	// write dummy blog post
	postFp := filepath.Join(d, "test.md")
	content := "## Title\nhey [google](https://google.com)."
	os.WriteFile(postFp, []byte(content), 0644)

	// h3 instead of h2
	tmpl := filepath.Join(d, "template.html")
	if err := os.WriteFile(tmpl, []byte(`
	<div class="blog-post">
		<h3>{{.Heading}}</h3>
		<span class="dt-posted">{{.DtPosted}}</span>
		{{.Content}}
	</div>
	`), 0644); err != nil {
		t.Error("could not write file:", err)
	}
	post, err := NewBlogPost(postFp, WithTemplateFile(tmpl))
	if err != nil {
		t.Errorf("could not instantiate post object: %v", err)
	}

	var html bytes.Buffer
	if err := post.WriteHtml(&html); err != nil {
		t.Error("could not write html:", err)
	}
	expHtml := regexp.MustCompile(`<div class="blog-post">\s*<h3>\s*Title\s*</h3>\s*<span class="dt-posted">` + time.Now().Format(time.DateOnly) + `</span>\s*<p>hey <a href="https://google.com" target="_blank">google</a>\.</p>\s*</div>`)
	if !expHtml.Match(html.Bytes()) {
		t.Errorf("generated html (%s) does not match '%v'", html.Bytes(), expHtml)
	}
}

func TestBlogpostWithPublicationTracking(t *testing.T) {
	d := t.TempDir()

	// write dummy blog post
	postFp := filepath.Join(d, "test.md")
	os.WriteFile(postFp, []byte("## Title\nhey [google](https://google.com)."), 0644)

	post, err := NewBlogPost(postFp, WithPublicationTracking())
	if err != nil {
		t.Errorf("could no instantiate post object: %v", err)
	}

	// first time publishing - should write current date to db
	var buf bytes.Buffer
	if err := post.WriteHtml(&buf); err != nil {
		t.Error("could not write buffer:", err)
	}

	expHtml := regexp.MustCompile(`<div class="blog-post">\s*<h2>\s*Title\s*</h2>\s*<span class="dt-posted">` + time.Now().Format(time.DateOnly) + `</span>\s*<p>hey <a href="https://google.com" target="_blank">google</a>\.</p>\s*</div>`)
	if !expHtml.Match(buf.Bytes()) {
		t.Errorf("generated html (%s) does not match '%v'", buf.Bytes(), expHtml)
	}

	// check registry
	registry, err := (&sqlitePool).Acquire(d)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	insertedDate, err := registry.GetPublicationDate(post)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if insertedDate == nil {
		t.Error("insertedDate is nil")
		t.FailNow()
	}

	today := time.Now().UTC()
	today = today.Truncate(24 * time.Hour)

	if *insertedDate != today {
		t.Errorf("expected inserted date to match today's date (%v), got %v", today, *insertedDate)
	}

	// second time publishing - should retrieve date from db
	// since we only store the date in the db, there is no easy way to check whether the date in the html
	// was taken from the db or not
	// we thus modify the value in the db to be able to distinguish between default (today) and value in db
	res, err := registry.DB.Exec("UPDATE posts SET dt_posted = '2024-01-01' WHERE name = ?;", post.GetName())
	if err != nil {
		t.Error(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		t.Error(err)
	}
	if affected != 1 {
		t.Error("")
	}

	var buf2 bytes.Buffer
	if err := post.WriteHtml(&buf2); err != nil {
		t.Error("could not write buffer:", err)
	}

	expHtml2 := regexp.MustCompile(`<div class="blog-post">\s*<h2>\s*Title\s*</h2>\s*<span class="dt-posted">` + time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.DateOnly) + `</span>\s*<p>hey <a href="https://google.com" target="_blank">google</a>\.</p>\s*</div>`)
	if !expHtml2.Match(buf2.Bytes()) {
		t.Errorf("generated html (%s) does not match '%v'", buf.Bytes(), expHtml2)
	}
}
