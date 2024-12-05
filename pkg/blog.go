package microblog

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"
)

// Represents a collection of BlogPosts in the same directory. Provides methods to render these BlogPosts
// as HTML.
type Blog interface {
	RenderPosts() ([]byte, error)
	RenderPostsAsync() ([]byte, error)
	GetDirectory() string
	GetBlogPosts() []BlogPost
}

type blog struct {
	Posts     []BlogPost
	Directory string
}

// Render all posts as HTML and return them as a single byte slice.
func (b *blog) RenderPosts() ([]byte, error) {
	var htmlBuffer bytes.Buffer
	for _, post := range b.Posts {
		err := post.WriteHtml(&htmlBuffer)
		if err != nil {
			return []byte{}, fmt.Errorf("could not render html for post %v: %v", post.GetFilePath(), err)
		}
	}
	return htmlBuffer.Bytes(), nil
}

// Alternative implementation of RenderPosts using goroutines.
func (b *blog) RenderPostsAsync() ([]byte, error) {
	g := new(errgroup.Group)

	var bufSlice = make([][]byte, len(b.Posts))
	for i, p := range b.Posts {
		g.Go(func() error {
			var buf bytes.Buffer
			if err := p.WriteHtml(&buf); err != nil {
				return err
			}
			bufSlice[i] = buf.Bytes()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return []byte{}, nil
	}
	// collect buffers
	var buf bytes.Buffer
	for _, b := range bufSlice {
		buf.Write(b)
	}
	return buf.Bytes(), nil
}

// Returns the directory that contains all blog posts.
func (b *blog) GetDirectory() string {
	return b.Directory
}

// Returns a slice of BlogPost structs that are associated with this Blog.
func (b *blog) GetBlogPosts() []BlogPost {
	return b.Posts
}

// Creates a new Blog.
// The first parameter is the path to the directory that contains the different blog posts as .md files.
// You can also pass post options to apply to all the BlogPost structs created within this function,
// see microblog.BlogPost for more information.
//
//	blog, err := microblog.NewBlog("/path/to/mm/directory")
//	blog, err := microblog.NewBlog("/path/to/mm/directory", microblog.WithTemplateFile("/path/to/html/template"))
func NewBlog(directory string, postOptions ...func(*blogPost) error) (Blog, error) {
	i, err := os.Stat(directory)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("directory %v does not exist", directory)
		}
		return nil, fmt.Errorf("could not acquire file info for %v: %v", directory, err)
	}
	if !i.IsDir() {
		return nil, fmt.Errorf("%v must be a directory", directory)
	}
	markdownFiles, err := filepath.Glob(filepath.Join(directory, "*.md"))
	if err != nil {
		return nil, fmt.Errorf("failed to search for markdown files in %v: %v", directory, err)
	}
	if len(markdownFiles) == 0 {
		return nil, errors.New("there must be at least .md file in the directory")
	}
	var blogPosts = make([]BlogPost, 0, len(markdownFiles))
	for _, md := range markdownFiles {
		post, err := NewBlogPost(md, postOptions...)
		if err != nil {
			return nil, fmt.Errorf("could not create blog object for %v: %v", post, err)
		}
		blogPosts = append(blogPosts, post)
	}
	return &blog{
		Directory: directory,
		Posts:     blogPosts,
	}, nil
}
