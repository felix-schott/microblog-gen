package microblog

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// Represents a storage backend with functionality to retrieve and set the publication date for a blog post.
type Registry interface {
	GetPublicationDate(*blogPost) (*time.Time, error)
	SetPublicationDate(*blogPost) error
}

// Type of storage backend.
type RegistryType string

const (
	SQLite RegistryType = "sqlite"
)

type sqliteRegistry struct {
	DB       *sql.DB
	location string // path to db file
}

func (r *sqliteRegistry) GetLocation() string {
	return r.location
}

type sqlitePoolType struct {
	sync.Map
}

var sqlitePool sqlitePoolType // singleton

// Create a new *sqliteRegistry object for the given directory (= Blog) alongside the *sql.DB object
// it is based on, or return an existing one.
func (pool *sqlitePoolType) Acquire(directory string) (*sqliteRegistry, error) {
	loadedRegistry, ok := pool.Load(directory)
	if ok {
		return loadedRegistry.(*sqliteRegistry), nil
	}

	registry, err := createSqliteRegistry(directory)
	if err != nil {
		return nil, err
	}
	pool.LoadOrStore(directory, registry)
	return registry, nil
}

func createSqliteRegistry(directory string) (*sqliteRegistry, error) {
	fp := filepath.Join(directory, "blog.sqlite")

	f, err := os.Create(fp)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	db, err := sql.Open("sqlite", fmt.Sprintf("file:%v?mode=rw", fp))
	if err != nil {
		return nil, err
	}

	// initialise table if necessary
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS posts (
			name TEXT NOT NULL,
			dt_posted DATE DEFAULT CURRENT_DATE
		);
		CREATE UNIQUE INDEX IF NOT EXISTS name_idx ON posts (name);
	`)
	if err != nil {
		return nil, err
	}

	return &sqliteRegistry{DB: db, location: fp}, nil
}

// Set publication date for a given blog post. If the second parameter t is nil, the database default (current date)
// will be used.
// Returns the inserted date as a time.Time pointer and an error if applicable.
func (r *sqliteRegistry) SetPublicationDate(p BlogPost, t *time.Time) (*time.Time, error) {
	var rows *sql.Rows
	var err error
	if t == nil {
		rows, err = r.DB.Query("INSERT INTO posts (name) VALUES (?) RETURNING dt_posted;", p.GetName())
	} else {
		rows, err = r.DB.Query("INSERT INTO posts VALUES (?, ?) RETURNING dt_posted;", p.GetName(), t.Format(time.DateOnly))
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// first row
	var dtPosted *time.Time
	rows.Next()
	rows.Scan(&dtPosted)

	if rows.Next() {
		return nil, errors.New("expected only one row to be returned, got multiple")
	}

	return dtPosted, nil
}

// Get the publication date for a given BlogPost.
// If there is no entry for that blog post in the database (and there is no error otherwise),
// the function will return (nil, nil).
func (r *sqliteRegistry) GetPublicationDate(p BlogPost) (*time.Time, error) {
	row := r.DB.QueryRow("SELECT dt_posted FROM posts WHERE name = ?;", p.GetName())
	var dtPosted *time.Time

	if err := row.Scan(&dtPosted); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return dtPosted, nil
}
