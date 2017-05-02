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

func Test_openDbFile(t *testing.T) {
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

// func Test_dbTags(t *testing.T) {
// 	type args struct {
// 		p *Page
// 	}
// 	tests := []struct {
// 		name       string
// 		addTags    []string
// 		updateTags []string
// 		deleteTags []string
// 		// fields  fields
// 		// args    args
// 		wantRet []Tag
// 		wantErr bool
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			db, err := setupDb()
// 			if err != nil {
// 				t.Errorf("Setting up db failed with error = %v", err)
// 				return
// 			}
// 			defer teardownDb()

// 			gotRet, err := db.getTags(tt.args.p)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("db.getTags() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(gotRet, tt.wantRet) {
// 				t.Errorf("db.getTags() = %v, want %v", gotRet, tt.wantRet)
// 			}
// 		})
// 	}
// }

func Test_db_addTag(t *testing.T) {
	tests := []struct {
		name    string
		args    Tag
		wantErr bool
	}{
		{"Add empty tag", Tag{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := setupDb()
			if err != nil {
				t.Errorf("Setting up db failed with error = %v", err)
				return
			}
			defer teardownDb()

			if err := db.addTag(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("db.addTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			tags, err := db.getTags(nil)
			if err != nil {
				t.Errorf("db.getTags() error = %v", err)
			}

			if !reflect.DeepEqual(tags[0], tt.args) {
				t.Errorf("db.getTags() = %v, want %v", tags[0], tt.args)
			}
		})
	}
}
