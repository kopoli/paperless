package paperless

import (
	"os"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

var dbfile = "test.sqlite"

func setupDb() (*db, error) {
	return openDbFile(dbfile)
}

func clearDbFile(dbfile string) error {
	return os.Remove(dbfile)
}

func teardownDb() (err error) {
	return clearDbFile(dbfile)
}

func Test_db_openDbFile(t *testing.T) {
	tests := []struct {
		name    string
		dbfile  string
		wantErr bool
	}{
		{"Empty filename", "", true},
		{"Improper filename", "././", true},
		{"Proper filename", "test.sqlite", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRet, err := openDbFile(tt.dbfile)
			if (err != nil) != tt.wantErr {
				t.Errorf("openDbFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if gotRet == nil {
				t.Errorf("openDbFile() returns nil and no error")
				return
			}
			err = gotRet.Close()
			if err != nil {
				t.Errorf("db.Close() error = %v", err)
			}
			_, err = os.Stat(tt.dbfile)
			if err != nil {
				t.Errorf("Statting %s errors = %v", tt.dbfile, err)
			}
			err = clearDbFile(tt.dbfile)
			if err != nil {
				t.Errorf("clearDbFile() error = %v", err)
			}
		})
	}
}

type testOp interface {
	run(*db) error
}

type testFunc func(*db) error

func (t testFunc) run(d *db) error {
	return t(d)
}

func Test_db_Tag(t *testing.T) {
	at := func(name, comment string) testFunc {
		return func(d *db) error {
			return d.addTag(Tag{Name: name, Comment: comment})
		}
	}

	dt := func(name string) testFunc {
		return func(d *db) error {
			return d.deleteTag(Tag{Name: name})
		}
	}

	ut := func(name, comment string) testFunc {
		return func(d *db) error {
			return d.upsertTag(Tag{Name: name, Comment: comment})
		}
	}

	tests := []struct {
		name     string
		ops      []testOp
		wantErr  bool
		paging   *Page
		wantTags []Tag
	}{
		{"Add empty tag", []testOp{at("", "")}, false, nil, []Tag{Tag{Id: 1}}},
		{"Add tag with contents", []testOp{at("name", "")}, false, nil, []Tag{Tag{Id: 1, Name: "name"}}},
		{"Add tag and remove it", []testOp{
			at("name", ""), at("abc", ""), dt("name"),
		}, false, nil, []Tag{Tag{Id: 2, Name: "abc"}}},
		{"Add tag and update it", []testOp{
			at("name", ""), ut("name", "comment"),
		}, false, nil, []Tag{Tag{Id: 1, Name: "name", Comment: "comment"}}},
		{"Upsert a tag", []testOp{
			ut("name", "comment"), ut("other", ""),
		}, false, nil, []Tag{Tag{Id: 1, Name: "name", Comment: "comment"}, Tag{Id: 2, Name: "other"}}},
		{"Add duplicate", []testOp{
			at("name", ""), at("name", "other"),
		}, true, nil, []Tag{Tag{Id: 1, Name: "name"}}},
		{"Pagination", []testOp{
			at("f1", ""), at("f2", ""), at("f3", ""), at("f4", ""),
		}, false, &Page{SinceId: 2, Count: 5}, []Tag{Tag{Id: 3, Name: "f3"}, Tag{Id: 4, Name: "f4"}}},
	}
	for _, tt := range tests {
		db, err := setupDb()
		if err != nil {
			t.Errorf("Setting up db failed with error = %v", err)
			return
		}
		t.Run(tt.name, func(t *testing.T) {

			var failed bool = false
			fail := struct {
				failed bool
				err    error
				i      int
			}{}

			for i, op := range tt.ops {
				err := op.run(db)
				failed = failed || (err != nil)
				if failed && !fail.failed {
					fail.failed = true
					fail.err = err
					fail.i = i
				}
			}
			if failed != tt.wantErr {
				t.Errorf("op no.%d error = %v, wantErr %v", fail.i, fail.err, tt.wantErr)
				return
			}

			tags, err := db.getTags(tt.paging)
			if err != nil {
				t.Errorf("db.getTags() error = %v", err)
			}

			if !reflect.DeepEqual(tags, tt.wantTags) {
				t.Errorf("db.getTags() = %v, want %v", tags, tt.wantTags)
			}
		})
		db.Close()
		err = teardownDb()
		if err != nil {
			t.Errorf("Could not remove database file: %v", err)
		}
	}
}


func Test_db_Script(t *testing.T) {
	at := func(name, script string) testFunc {
		return func(d *db) error {
			return d.addScript(Script{Name: name, Script: script})
		}
	}

	dt := func(name string) testFunc {
		return func(d *db) error {
			return d.deleteScript(Script{Name: name})
		}
	}

	ut := func(name, script string) testFunc {
		return func(d *db) error {
			return d.updateScript(Script{Name: name, Script: script})
		}
	}

	tests := []struct {
		name     string
		ops      []testOp
		wantErr  bool
		paging   *Page
		wantScripts []Script
	}{
		{"Add empty script", []testOp{at("", "")}, false, nil, []Script{Script{Id: 1}}},
		{"Add script with contents", []testOp{at("name", "")}, false, nil, []Script{Script{Id: 1, Name: "name"}}},
		{"Add script and remove it", []testOp{
			at("name", ""), at("abc", ""), dt("name"),
		}, false, nil, []Script{Script{Id: 2, Name: "abc"}}},
		{"Add script and update it", []testOp{
			at("name", ""), ut("name", "script"),
		}, false, nil, []Script{Script{Id: 1, Name: "name", Script: "script"}}},
		{"Update a script", []testOp{
			at("name", "script"), ut("name", "toinen"),
		}, false, nil, []Script{Script{Id: 1, Name: "name", Script: "toinen"}}},
		{"Add duplicate", []testOp{
			at("name", ""), at("name", "other"),
		}, true, nil, []Script{Script{Id: 1, Name: "name"}}},
		{"Pagination", []testOp{
			at("f1", ""), at("f2", ""), at("f3", ""), at("f4", ""),
		}, false, &Page{SinceId: 2, Count: 5}, []Script{Script{Id: 3, Name: "f3"}, Script{Id: 4, Name: "f4"}}},
	}
	for _, tt := range tests {
		db, err := setupDb()
		if err != nil {
			t.Errorf("Setting up db failed with error = %v", err)
			return
		}
		t.Run(tt.name, func(t *testing.T) {

			var failed bool = false
			fail := struct {
				failed bool
				err    error
				i      int
			}{}

			for i, op := range tt.ops {
				err := op.run(db)
				failed = failed || (err != nil)
				if failed && !fail.failed {
					fail.failed = true
					fail.err = err
					fail.i = i
				}
			}
			if failed != tt.wantErr {
				t.Errorf("op no.%d error = %v, wantErr %v", fail.i, fail.err, tt.wantErr)
				return
			}

			scripts, err := db.getScripts(tt.paging)
			if err != nil {
				t.Errorf("db.getScripts() error = %v", err)
			}

			if !reflect.DeepEqual(scripts, tt.wantScripts) {
				t.Errorf("db.getScripts() = %v, want %v", scripts, tt.wantScripts)
			}
		})
		db.Close()
		err = teardownDb()
		if err != nil {
			t.Errorf("Could not remove database file: %v", err)
		}
	}
}