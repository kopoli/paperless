package paperless

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/kopoli/go-util"
)

type db struct {
	file string
	*sqlx.DB
}

type dbTx struct {
	*sqlx.Tx
}

// Pagination support
type Page struct {
	// Id that was the last of the previous page
	SinceId int

	// Count is the number of items in the page
	Count int
}

func openDbFile(dbfile string) (ret *db, err error) {
	create := false

	dbfile = filepath.Clean(dbfile)

	i, err := os.Stat(dbfile)
	if err == nil && i.IsDir() {
		err = util.E.New("Given path is a directory")
		return
	}

	if _, err = os.Stat(dbfile); os.IsNotExist(err) {
		create = true
		err = nil
	}

	d, err := sqlx.Open("sqlite3", fmt.Sprintf("file:%s?cache=shared&mode=rwc", dbfile))
	if err != nil {
		err = util.E.Annotate(err, "Opening sqlite dbfile failed")
		return
	}

	if create {
		_, err = d.Exec(`
CREATE TABLE IF NOT EXISTS tag (
  id INTEGER PRIMARY KEY ASC AUTOINCREMENT,
  name TEXT DEFAULT "" NOT NULL UNIQUE ON CONFLICT ABORT,
  comment TEXT DEFAULT ""
);

-- The image data
CREATE TABLE IF NOT EXISTS img (
  id INTEGER PRIMARY KEY ASC AUTOINCREMENT,
  checksum TEXT UNIQUE ON CONFLICT ABORT,	-- checksum of the file
  fileid TEXT DEFAULT "",                       -- used to construct the processed image,
						--   thumbnail and text files
  isdiscarded INTEGER DEFAULT 0,                -- Should the image be ignored on ocr
  parentid INTEGER DEFAULT -1,                  -- ID of a parent image
  tmscanned INTEGER DEFAULT 0,                  -- timestamp when it was scanned
  tmprocessed INTEGER DEFAULT 0,                -- timestamp when it was processed
  tmreinterpret INTEGER DEFAULT 0,              -- timestamp when it was reinterpret

  processlog TEXT DEFAULT "",                   -- Log of processing
  filename TEXT DEFAULT "",                     -- The original filename
  isuploaded INTEGER DEFAULT 0,                 -- has a real image been uploaded to paperless yet.

  istobeocrd INTEGER DEFAULT 0			-- Needs to be re-ocr'd
);

CREATE VIRTUAL TABLE IF NOT EXISTS imgtext USING fts4 (
  text DEFAULT "",				-- the OCR'd text
  comment DEFAULT ""				-- freeform comment
);

-- Tags for an image
CREATE TABLE IF NOT EXISTS imgtag (
  tagid INTEGER REFERENCES tag(id) NOT NULL,
  imgid INTEGER REFERENCES img(id) NOT NULL
);

-- Script for processing the images
CREATE TABLE IF NOT EXISTS script (
  id INTEGER PRIMARY KEY ASC AUTOINCREMENT,
  name TEXT UNIQUE ON CONFLICT ABORT,
  script TEXT DEFAULT ""
);

`)
		if err != nil {
			goto initfail
		}

	}
	d.Exec("PRAGMA busy_timeout=2000")
	if err != nil {
		goto initfail
	}

	ret = &db{dbfile, d}
	return

initfail:
	d.Close()
	err = util.E.Annotate(err, "Initializing the database failed")
	ret = nil
	return
}

func (db *db) getTags(p *Page) (ret []Tag, err error) {
	query := "SELECT * from tag"
	order := " ORDER BY name ASC"
	sel := func() error {
		return db.Select(&ret, query + order)
	}

	if p != nil {
		query += " WHERE (id > ?) " + order + " LIMIT ?"
		sel = func() error {
			return db.Select(&ret, query, p.SinceId, p.Count)
		}
	}

	err = sel()
	return
}

func (db *db) addTag(t Tag) (err error) {
	_, err = db.Exec("INSERT INTO tag(name, comment) VALUES($1, $2)", t.Name, t.Comment)
	return
}

func (db *db) upsertTag(t Tag) (err error) {
	tx, err := db.Beginx()
	if err != nil {
		return
	}

	_, _ = tx.Exec("UPDATE tag SET comment = $1 WHERE name = $2", t.Comment, t.Name)
	_, err = tx.Exec("INSERT OR IGNORE INTO tag(name, comment) VALUES($1, $2)", t.Name, t.Comment)
	if err != nil {
		return
	}

	err = tx.Commit()
	return
}

func (db *db) deleteTag(t Tag) (err error) {
	_, err = db.Exec("DELETE FROM tag WHERE name = $1", t.Name)
	return
}

func (db *db) getScripts(p *Page) (ret []Script, err error) {
	query := "SELECT * from script"
	order := " ORDER BY name ASC"
	sel := func() error {
		return db.Select(&ret, query + order)
	}

	if p != nil {
		query += " WHERE (id > ?) " + order + " LIMIT ?"
		sel = func() error {
			return db.Select(&ret, query, p.SinceId, p.Count)
		}
	}

	err = sel()
	return
}

func (db *db) addScript(s Script) (err error) {
	_, err = db.Exec("INSERT INTO script(name, script) VALUES($1, $2)", s.Name, s.Script)
	return
}

func (db *db) updateScript(s Script) (err error) {
	_, err = db.Exec("UPDATE script SET script = $1 WHERE name = $2", s.Script, s.Name)
	return
}

func (db *db) deleteScript(s Script) (err error) {
	_, err = db.Exec("DELETE FROM script WHERE name = $1", s.Name)
	return
}
