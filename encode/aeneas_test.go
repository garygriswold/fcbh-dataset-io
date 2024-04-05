package encode

import (
	"context"
	"dataset"
	"dataset/db"
	"testing"
)

func TestAeneasLines(t *testing.T) {
	var ctx = context.Background()
	var bibleId = `ENGWEB`
	var filesetId = `ENGWEBN2DA`
	var language = `eng`
	var conn = db.NewDBAdapter(ctx, `ENGWEB_DBPTEXT.db`)
	//files, status := ReadDirectory(ctx, bibleId, filesetId)
	aeneas := NewAeneas(ctx, conn, bibleId, filesetId)
	aeneas.Process(language, dataset.LINES)
}

func TestAeneasWords(t *testing.T) {
	var ctx = context.Background()
	var bibleId = `ENGWEB`
	var filesetId = `ENGWEBN2DA`
	var language = `eng`
	var conn = db.NewDBAdapter(ctx, `ENGWEB_DBPTEXT.db`)
	//files, status := ReadDirectory(ctx, bibleId, filesetId)
	aeneas := NewAeneas(ctx, conn, bibleId, filesetId)
	aeneas.Process(language, dataset.WORDS)
}
