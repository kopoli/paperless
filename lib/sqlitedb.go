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

// PageResult adds navigation information to the paginated data
type PageResult struct {
	// Number of items returned by the whole search
	ResultCount int

	// IDs of the items that are the first of their page
	SinceIDs []int

	// Count is the number of items in the page
	Count int
}

type ImageResult struct {
	PageResult
	Images []Image
}

type Search struct {
	ID      int
	OrderBy string
	Match   string
	Tag     string
}

func openDbFile(dbfile string) (ret *db, err error) {
	create := false

	dbfile = filepath.Clean(dbfile)

	i, err := os.Stat(dbfile)
	if err == nil && i.IsDir() {
		err = util.E.New("Given path is a directory")
		return
	}

	err = MkdirParents(dbfile)
	if err != nil {
		err = util.E.Annotate(err,
			"Could not create dbfile's parent directories")
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
	_, err = d.Exec("PRAGMA busy_timeout=10000")
	if err != nil {
		goto initfail
	}

	// Work around the multiple access problems
	d.SetMaxOpenConns(1)

	ret = &db{dbfile, d}
	return

initfail:
	d.Close()
	err = util.E.Annotate(err, "Initializing the database failed")
	ret = nil
	return
}

func (db *db) getTag(id int) (ret Tag, err error) {
	err = db.Get(&ret, "SELECT * from tag WHERE id = $1", id)
	return
}

func (db *db) getScript(id int) (ret Script, err error) {
	err = db.Get(&ret, "SELECT * from script WHERE id = $1", id)
	return
}

func (db *db) getImage(id int) (ret Image, err error) {
	if id < 0 {
		err = util.E.New("Negative ID for image is invalid")
		return
	}

	imgs, err := db.getImages(nil, &Search{ID: id})
	if err != nil {
		return
	}

	if len(imgs.Images) == 0 {
		err = util.E.New("No image found with id %d", id)
		return
	}
	if len(imgs.Images) > 1 {
		err = util.E.New("Internal error: Multiple images with the same id")
		return
	}
	ret = imgs.Images[0]
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

func (db *db) addTag(t Tag) (ret Tag, err error) {
	_, err = db.Exec("INSERT INTO tag(name, comment) VALUES($1, $2)", t.Name, t.Comment)
	if err != nil {
		return
	}
	err = db.Get(&ret, "SELECT * FROM tag WHERE name = $1", t.Name)

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

func (db *db) addScript(s Script) (ret Script, err error) {
	_, err = db.Exec("INSERT INTO script(name, script) VALUES($1, $2)", s.Name, s.Script)
	if err != nil {
		return
	}

	err = db.Get(&ret, "SELECT * FROM script WHERE name = $1", s.Name)
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

func (db *db) getImages(p *Page, s *Search) (ret ImageResult, err error) {
	query := "SELECT image.id FROM image, imgtext"
	order := " ORDER BY image.id ASC"

	where := " WHERE imgtext.rowid = image.id"

	args := map[string]interface{}{}

	if s != nil {
		if s.ID != 0 {
			where = where + " AND image.id = :id"
			args["id"] = fmt.Sprintf("%d", s.ID)
		}
		if s.Match != "" {
			where = where + " AND imgtext.text MATCH :match"
			args["match"] = s.Match
		}
		if s.Tag != "" {
			query = query + ", tag, imgtag"
			where = where + " AND tag.name = :tag AND imgtag.tagid = tag.id AND imgtag.imgid = image.id"
			args["tag"] = s.Tag
		}
		if s.OrderBy != "" {
			order = " ORDER BY :order ASC"
			args["order"] = s.OrderBy
		}
	}
	query = query + where + order

	nstmt, err := db.PrepareNamed(query)
	if err != nil {
		return
	}

	var ids []int

	err = nstmt.Select(&ids, args)
	if err != nil {
		nstmt.Close()
		return
	}
	nstmt.Close()

	ret.ResultCount = len(ids)

	// No images found
	if ret.ResultCount == 0 {
		ret.SinceIDs = make([]int, 0)
		return
	}

	if p == nil {
		ret.Count = 0
		ret.SinceIDs = make([]int, 1)
		ret.SinceIDs[0] = ids[0]
	} else {
		ret.Count = p.Count
		pages := ret.ResultCount / ret.Count
		if (ret.ResultCount % ret.Count) > 0 {
			pages += 1
		}

		ret.SinceIDs = make([]int, pages)
		for i := range ret.SinceIDs {
			ret.SinceIDs[i] = ids[ret.Count*i]
		}

		var start int = -1
		var realcount int = 0
		for i := range ids {
			if ids[i] == p.SinceId {
				start = i + 1
				break
			}
		}
		if start == -1 || start >= len(ids) {
			start = 0
		}

		realcount = ret.ResultCount - start
		if realcount > ret.Count {
			realcount = ret.Count
		}

		ids = ids[start : start+realcount]
	}

	q, qargs, err := sqlx.In(`SELECT * from image, imgtext WHERE imgtext.rowid = image.id AND image.id IN (?)`, ids)
	if err != nil {
		return
	}
	err = db.Select(&ret.Images, q, qargs...)
	if err != nil {
		return
	}

	err = withTx(db, func(tx *sqlx.Tx) (err error) {
		for i := range ret.Images {
			err = tx.Select(&ret.Images[i].Tags, `SELECT tag.id, tag.name, tag.comment FROM tag, imgtag
                                                       WHERE imgtag.tagid = tag.id AND imgtag.imgid = $1 `, ret.Images[i].Id)
			if err != nil {
				return
			}
		}
		return
	})
	return
}

func syncTagsToImage(tx *sqlx.Tx, i Image) (err error) {
	_, err = tx.NamedExec(`DELETE FROM imgtag WHERE imgid = :id`, i)
	if err != nil {
		return
	}

	for _, t := range i.Tags {
		_, err = tx.Exec(`INSERT INTO imgtag(imgid, tagid) SELECT $1, tag.id FROM tag WHERE tag.name = $2`, i.Id, t.Name)
		if err != nil {
			return
		}
	}
	return
}

func (db *db) addImage(i Image) (ret Image, err error) {
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
		if err != nil {
			return
		}

		err = syncTagsToImage(tx, i)
		ret = i
		return
	})
	return
}

func (db *db) updateImage(i Image) (err error) {
	err = withTx(db, func(tx *sqlx.Tx) (err error) {
		_, err = tx.NamedExec(`UPDATE image SET
                      interpretdate = :interpretdate,
                      processlog = :processlog
                      WHERE image.id = :id`, i)
		if err != nil {
			return
		}

		_, err = tx.NamedExec(`UPDATE imgtext SET
                      text = :text,
                      comment = :comment
                      WHERE rowid = :id`, i)
		if err != nil {
			return
		}

		err = syncTagsToImage(tx, i)
		return
	})
	return
}

func (db *db) deleteImage(s Image) (err error) {
	err = withTx(db, func(tx *sqlx.Tx) (err error) {
		_, err = tx.Exec(`DELETE FROM imgtag WHERE imgid IN
                                  (SELECT id FROM image WHERE image.checksum = $1)`, s.Checksum)
		if err != nil {
			return
		}
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
