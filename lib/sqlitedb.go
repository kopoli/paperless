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

type Search struct {
	Where   string
	OrderBy string
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
CREATE TABLE IF NOT EXISTS image (
  id INTEGER PRIMARY KEY ASC AUTOINCREMENT,
  checksum TEXT UNIQUE NOT NULL ON CONFLICT ABORT,-- checksum of the file
  fileid TEXT DEFAULT "",                       -- used to construct the processed image,
						--   thumbnail and text files
  scandate DATETIME,                            -- timestamp when it was scanned
  adddate  DATETIME DEFAULT CURRENT_TIMESTAMP,  -- timestamp when it was created in db
  interpretdate DATETIME,                       -- timestamp when it was interpret

  processlog TEXT DEFAULT "",                   -- Log of processing
  filename TEXT DEFAULT ""                     -- The original filename
);

CREATE VIRTUAL TABLE IF NOT EXISTS imgtext USING fts4 (
  text DEFAULT "",				-- the OCR'd text
  comment DEFAULT ""				-- freeform comment
);

-- Tags for an image
CREATE TABLE IF NOT EXISTS imgtag (
  tagid INTEGER REFERENCES tag(id) NOT NULL,
  imgid INTEGER REFERENCES img(id) NOT NULL,
  UNIQUE (tagid, imgid)
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
		return db.Select(&ret, query+order)
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

func (db *db) updateTag(t Tag) (err error) {
	_, err = db.Exec("UPDATE tag SET comment = $1 WHERE name = $2", t.Comment, t.Name)
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
		return db.Select(&ret, query+order)
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

func withTx(db *db, f func(*sqlx.Tx) error) (err error) {
	tx, err := db.Beginx()
	if err != nil {
		return
	}

	err = f(tx)
	if err != nil {
		tx.Rollback()
		return
	}

	err = tx.Commit()
	return
}

func (db *db) getImages(p *Page, s *Search) (ret []Image, err error) {
	query := "SELECT * FROM image, imgtext"
	order := " ORDER BY image.id ASC"

	where := " WHERE imgtext.rowid = image.id"

	args := map[string]interface{}{}

	if s != nil {
		where = where + " AND imgtext.text MATCH :where"
		args["where"] = s.Where
		order = " ORDER by :order ASC"
		args["order"] = s.OrderBy
	}
	if p != nil {
		where = where + " AND (image.id > :id)"
		args["id"] = fmt.Sprintf("%d", p.SinceId)
		order = order + " LIMIT :limit"
		args["limit"] = fmt.Sprintf("%d", p.Count)
	}

	query = query + where + order

	fmt.Println("Query on", query)

	nstmt, err := db.PrepareNamed(query)
	if err != nil {
		return
	}
	defer nstmt.Close()

	err = nstmt.Select(&ret, args)
	return
}

func (db *db) addImage(i Image) (err error) {
	err = withTx(db, func(tx *sqlx.Tx) (err error) {
		_, err = tx.NamedExec(`INSERT INTO
                   image(  checksum,  fileid,  scandate,  adddate,  interpretdate,  processlog,  filename)
                   VALUES(:checksum, :fileid, :scandate, :adddate, :interpretdate, :processlog, :filename)`, i)
		if err != nil {
			return
		}

		var id int
		err = tx.Get(&id, "SELECT id FROM image WHERE checksum=$1", i.Checksum)
		if err != nil {
			return
		}
		i.Id = id

		_, err = tx.NamedExec(`INSERT INTO imgtext(rowid, text, comment) VALUES (:id, :text, :comment)`, i)
		return
	})
	return
}

func (db *db) updateImage(s Image) (err error) {
	err = withTx(db, func(tx *sqlx.Tx) (err error) {
		_, err = tx.NamedExec(`UPDATE image SET
                      interpretdate = :interpretdate,
                      processlog = :processlog
                      WHERE image.id = :id`, s)

		if err != nil {
			return
		}

		_, err = tx.NamedExec(`UPDATE imgtext SET
                      text = :text,
                      comment = :comment
                      WHERE rowid = :id`, s)
		return
	})
	return
}

func (db *db) deleteImage(s Image) (err error) {
	err = withTx(db, func(tx *sqlx.Tx) (err error) {
		_, err = tx.Exec(`DELETE FROM imgtext WHERE rowid IN
                                  (SELECT id FROM image WHERE image.checksum = $1)`, s.Checksum)
		if err != nil {
			return
		}
		_, err = tx.Exec(`DELETE FROM image WHERE image.checksum = $1`, s.Checksum)
		return
	})
	return
}
