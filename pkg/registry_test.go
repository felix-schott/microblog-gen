package microblog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSqliteRegistry(t *testing.T) {

	d := t.TempDir()

	registry, registryErr := (&sqlitePool).Acquire(d)
	if registryErr != nil {
		contents, err := os.ReadDir(d)
		if err != nil {
			t.Error(err)
		}
		if len(contents) != 1 {
			t.Error("empty directory, contents: ", contents)
		}
		t.Error(registryErr)
		t.FailNow()
	}
	if registry == nil {
		t.Error("registry is nil!")
		t.FailNow()
	}
	if registry.DB == nil {
		t.Error("registry.DB is nil!")
		t.FailNow()
	}
	if err := registry.DB.Ping(); err != nil {
		t.Error(err)
		t.FailNow()
	}

	if registry.GetLocation() != filepath.Join(d, "blog.sqlite") {
		t.Errorf("expected location of db file to be %v, instead got %v", filepath.Join(d, "blog.sqlite"), registry.GetLocation())
	}

	t.Run("TestSqliteGetPublicationDateNoData", func(t *testing.T) {
		fp := filepath.Join(d, "test.md")
		if err := os.WriteFile(fp, []byte("## hey"), 0644); err != nil {
			t.Error(err)
		}
		post, err := NewBlogPost(fp)
		if err != nil {
			t.Error(err)
		}
		date, err := registry.GetPublicationDate(post)
		if err != nil {
			t.Error(err)
		}
		if date != nil {
			t.Error("expected the function to return nil because there is no data in the DB")
		}
	})

	t.Run("TestSqliteSetAndGetPublicationDate", func(t *testing.T) {
		fp := filepath.Join(d, "test2.md")
		if err := os.WriteFile(fp, []byte("## hey"), 0644); err != nil {
			t.Error(err)
		}
		post, err := NewBlogPost(fp)
		if err != nil {
			t.Error(err)
		}
		today := time.Now().UTC()
		today = today.Truncate(24 * time.Hour)

		insertedDate, err := registry.SetPublicationDate(post, nil)
		if err != nil {
			t.Error(err)
		}
		if insertedDate == nil {
			t.Error("insertedDate is nil")
			t.FailNow()
		}
		if *insertedDate != today {
			t.Errorf("expected SetPublicationDate to return today's date (%v) because nil was passed as a second argument to SetPublicationDate, instead got %v", today, *insertedDate)
		}

		returnedDate, err := registry.GetPublicationDate(post)
		if err != nil {
			t.Error(err)
		}
		if *returnedDate != today {
			t.Errorf("expected GetPublicationDate to return today's date (%v) as was inserted into the DB, instead got %v", today, *returnedDate)
		}
	})

	t.Run("TestSqliteSetAndGetPublicationDate2", func(t *testing.T) {
		fp := filepath.Join(d, "test3.md")
		if err := os.WriteFile(fp, []byte("## hey"), 0644); err != nil {
			t.Error(err)
		}
		post, err := NewBlogPost(fp)
		if err != nil {
			t.Error(err)
		}

		newDate := time.Date(2024, 1, 1, 1, 1, 1, 1, time.UTC) // time component will be ignored!
		var newDateAsInserted time.Time = newDate
		newDateAsInserted = newDateAsInserted.Truncate(24 * time.Hour)

		insertedDate, err := registry.SetPublicationDate(post, &newDate)
		if err != nil {
			t.Error(err)
		}
		if insertedDate == nil {
			t.Error("inserted date is nil!")
			t.FailNow()
		}
		if *insertedDate != newDateAsInserted {
			t.Errorf("expected SetPublicationDate to return the inserted date without time component (%v), instead got %v", newDateAsInserted, *insertedDate)
		}

		returnedDate, err := registry.GetPublicationDate(post)
		if err != nil {
			t.Error(err)
		}
		if *returnedDate != newDateAsInserted {
			t.Errorf("expected GetPublicationDate to return inserted date without time component (%v), instead got %v", newDateAsInserted, *returnedDate)
		}
	})

	t.Run("TestSqliteSetPublicationDateDuplicate", func(t *testing.T) {
		fp := filepath.Join(d, "test4.md")
		if err := os.WriteFile(fp, []byte("## hey"), 0644); err != nil {
			t.Error(err)
		}
		post, err := NewBlogPost(fp)
		if err != nil {
			t.Error(err)
		}

		// insert once
		newDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

		insertedDate, err := registry.SetPublicationDate(post, &newDate)
		if err != nil {
			t.Error(err)
		}
		if insertedDate == nil {
			t.Error("inserted date is nil!")
			t.FailNow()
		}
		if *insertedDate != newDate {
			t.Errorf("expected SetPublicationDate to return the inserted date without time component (%v), instead got %v", newDate, *insertedDate)
		}

		// insert another time

		newDate2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

		insertedDate2, err := registry.SetPublicationDate(post, &newDate2)
		if err == nil {
			t.Error("expected error to be returned")
		}

		if !strings.Contains(err.Error(), "UNIQUE") {
			t.Error("expected unique violation error, got", err)
		}

		if insertedDate2 != nil {
			t.Error("expected inserted date to be null, instead got", insertedDate2)
			t.FailNow()
		}
	})
}
