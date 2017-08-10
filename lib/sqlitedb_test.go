package paperless

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pmezard/go-difflib/difflib"
)

func structEquals(a, b interface{}) bool {
	return spew.Sdump(a) == spew.Sdump(b)
}

func diffStr(a, b interface{}) (ret string) {
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(spew.Sdump(a)),
		B:        difflib.SplitLines(spew.Sdump(b)),
		FromFile: "Expected",
		ToFile:   "Received",
		Context:  3,
	}

	ret, _ = difflib.GetUnifiedDiffString(diff)
	return
}

func compare(t *testing.T, msg string, a, b interface{}) {
	if !structEquals(a, b) {
		t.Error(msg, "\n", diffStr(a, b))
	}
}

var dbfile = ":memory:"

func setupDb() (*db, error) {
	return openDbFile(dbfile)
}

func clearDbFile(dbfile string) error {
	return nil
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
			_, err := d.addTag(Tag{Name: name, Comment: comment})
			return err
		}
	}

	dt := func(name string) testFunc {
		return func(d *db) error {
			return d.deleteTag(Tag{Name: name})
		}
	}

	ut := func(name, comment string) testFunc {
		return func(d *db) error {
			return d.updateTag(Tag{Name: name, Comment: comment})
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
		{"Update a tag", []testOp{
			at("name", ""), ut("name", "comment"),
		}, false, nil, []Tag{Tag{Id: 1, Name: "name", Comment: "comment"}}},
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

			compare(t, "db.getTags() not expected", tt.wantTags, tags)
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
			_, err := d.addScript(Script{Name: name, Script: script})
			return err
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
		name        string
		ops         []testOp
		wantErr     bool
		paging      *Page
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

			compare(t, "db.getScripts() not expected", tt.wantScripts, scripts)
		})
		db.Close()
		err = teardownDb()
		if err != nil {
			t.Errorf("Could not remove database file: %v", err)
		}
	}
}

func Test_db_Image(t *testing.T) {
	at := func(name, comment string) testFunc {
		return func(d *db) error {
			_, err := d.addTag(Tag{Name: name, Comment: comment})
			return err
		}
	}

	ai := func(i Image) testFunc {
		return func(d *db) error {
			_, err := d.addImage(i)
			return err
		}
	}

	di := func(checksum string) testFunc {
		return func(d *db) error {
			return d.deleteImage(Image{Checksum: checksum})
		}
	}

	ui := func(i Image) testFunc {
		return func(d *db) error {
			return d.updateImage(i)
		}
	}

	cmp := func(i1, i2 []Image) {
		for n := range i1 {
			i1[n].AddDate = time.Time{}
		}
		for n := range i2 {
			i2[n].AddDate = time.Time{}
		}

		compare(t, "db.getImages() not expected", i1, i2)
	}

	tests := []struct {
		name       string
		ops        []testOp
		wantErr    bool
		paging     *Page
		search     *Search
		wantImages []Image
	}{
		{"Add an image", []testOp{
			ai(Image{Checksum: "a", Fileid: "fid"}),
		}, false, nil, nil, []Image{Image{Id: 1, Checksum: "a", Fileid: "fid"}}},
		{"Add an images with text", []testOp{
			ai(Image{Checksum: "a", Fileid: "fid"}), ai(Image{Checksum: "b", Text: "b"}),
		}, false, nil, nil, []Image{Image{Id: 1, Checksum: "a", Fileid: "fid"}, Image{Id: 2, Checksum: "b", Text: "b"}}},
		{"Add image and remove it", []testOp{
			ai(Image{Checksum: "a", Text: "fid"}), ai(Image{Checksum: "b", ProcessLog: "pl"}), di("a"),
		}, false, nil, nil, []Image{Image{Id: 2, Checksum: "b", ProcessLog: "pl"}}},
		{"Add image and update it", []testOp{
			ai(Image{Checksum: "a", Text: "fid"}), ui(Image{Id: 1, Checksum: "a", Text: "other"}),
		}, false, nil, nil, []Image{Image{Id: 1, Checksum: "a", Text: "other"}}},
		{"Add a duplicate", []testOp{
			ai(Image{Checksum: "a", Text: "jeje"}), ai(Image{Checksum: "a", Text: "b"}),
		}, true, nil, nil, []Image{Image{Id: 1, Checksum: "a", Text: "jeje"}}},
		{"Pagination", []testOp{
			ai(Image{Checksum: "f1"}), ai(Image{Checksum: "f2"}),
			ai(Image{Checksum: "f3"}), ai(Image{Checksum: "f4"}),
		}, false, &Page{SinceId: 2, Count: 5}, nil, []Image{Image{Id: 3, Checksum: "f3"}, Image{Id: 4, Checksum: "f4"}}},
		{"Add image with tags", []testOp{
			at("eka", ""), at("toka", ""),
			ai(Image{Checksum: "a", Text: "jeje", Tags: []Tag{Tag{Name: "toka"}}}),
		}, false, nil, nil, []Image{Image{Id: 1, Checksum: "a", Text: "jeje", Tags: []Tag{Tag{Id: 2, Name: "toka"}}}}},
		{"Update image with tags", []testOp{
			at("eka", ""),
			ai(Image{Checksum: "a"}),
			ui(Image{Id: 1, Checksum: "a", Tags: []Tag{Tag{Id: 1, Name: "eka"}}}),
		}, false, nil, nil, []Image{Image{Id: 1, Checksum: "a", Tags: []Tag{Tag{Id: 1, Name: "eka"}}}}},
		{"Add images and search them", []testOp{
			ai(Image{Checksum: "a", Text: "first"}),
			ai(Image{Checksum: "b", Text: "second"}),
		}, false, nil, &Search{Match: "seco*"},
			[]Image{Image{Id: 2, Checksum: "b", Text: "second"}}},
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

			images, err := db.getImages(tt.paging, tt.search)
			if err != nil {
				t.Errorf("db.getImages() error = %v", err)
			}

			cmp(tt.wantImages, images)
		})
		db.Close()
		err = teardownDb()
		if err != nil {
			t.Errorf("Could not remove database file: %v", err)
		}
	}
}

func withDb(f func(*db) error) (err error) {
	_ = teardownDb()
	db, err := setupDb()
	if err != nil {
		return
	}

	e2 := f(db)

	db.Close()
	err = teardownDb()
	if err != nil {
		return
	}

	return e2
}

func addImages(db *db, count int) (err error) {
	for i := 0; i < count; i++ {
		idStr := "id:" + strconv.Itoa(i + 1)
		_, err = db.addImage(Image{
			Checksum:   idStr,
			Text: "jep " + idStr,
			ProcessLog: "jeje",
		})
		if err != nil {
			return
		}
	}
	return
}

func Benchmark_getImages(b *testing.B) {

	var countImages int = 100

	err := withDb(func(db *db) (err error) {
		err = addImages(db, countImages)
		if err != nil {
			b.Errorf("Adding images failed: %v", err)
		}

		b.ResetTimer()

		tests := []struct {
			name string
			page *Page
			search *Search
			count int
		}{
			{"Get all images", nil, nil, countImages},
			{"Get page of 10 images", &Page{SinceId: 3, Count: 10}, nil, 10},
			{"Get all images with string 7", nil, &Search{Match: "*7*"}, 11},
		}

		for _, tt := range tests {
			b.Run(tt.name, func(b *testing.B) {
				var is []Image
				for i := 0; i < b.N; i++ {
					is, err = db.getImages(tt.page, tt.search)
					if err != nil {
						b.Errorf("Getting images failed: %v", err)
						return
					}
					if len(is) != tt.count {
						b.Errorf("Expected %d images got %d images", tt.count, len(is))
					}
				}
			})
		}


		return
	})

	if err != nil {
		b.Errorf("Database handling failed with: %v", err)
	}
}
