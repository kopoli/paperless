package paperless

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/kopoli/go-util"
)

type db struct {
	*sqlx.DB
}

type dbTx struct {
	*sqlx.Tx
}

func openDbFile(dbfile string) (ret *db, err error) {
	create := false

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
  id INTEGER PRIMARY KEY,
  name TEXT UNIQUE ON CONFLICT IGNORE
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
`)
		if err != nil {
			goto initfail
		}

	}
	d.Exec("PRAGMA busy_timeout=2000")
	if err != nil {
		goto initfail
	}

	ret = &db{d}
	return

initfail:
	d.Close()
	err = util.E.Annotate(err, "Initializing the database failed")
	ret = nil
	return
}
