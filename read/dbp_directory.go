package read

import (
	"context"
	"dataset"
	"dataset/db"
	log "dataset/logger"
	"dataset/request"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type InputFiles struct {
	BookId    string // not used for text_plain
	Chapter   int    // only used for audio
	Filename  string
	Directory string
}

// DBPDirectory 1. Assign pattern for OT, NT.  2. Glob files.  3. Assign book/chapter & Prune
func DBPDirectory(ctx context.Context, bibleId string, fsType string, otFileset string, ntFileset string,
	testament request.Testament) ([]InputFiles, dataset.Status) {
	var results []InputFiles
	var files []InputFiles
	var status dataset.Status
	type run struct {
		filesetId string
		tType     string
	}
	var runs []run
	if otFileset != `` {
		runs = append(runs, run{filesetId: otFileset, tType: `OT`})
	}
	if ntFileset != `` {
		runs = append(runs, run{filesetId: ntFileset, tType: `NT`})
	}
	for _, r := range runs {
		files, status = Directory(ctx, bibleId, fsType, r.filesetId, r.tType, testament)
		if status.IsErr {
			return results, status
		}
		results = append(results, files...)
	}
	return results, status
}

func Directory(ctx context.Context, bibleId string, fsType string, filesetId string, tType string,
	testament request.Testament) ([]InputFiles, dataset.Status) {
	var status dataset.Status
	var directory string
	var search string
	if fsType == `text_plain` {
		directory = filepath.Join(os.Getenv("FCBH_DATASET_FILES"), bibleId)
		search = filepath.Join(directory, filesetId+".json")
	} else if fsType == `text_usx` {
		directory = filepath.Join(os.Getenv("FCBH_DATASET_FILES"), bibleId, filesetId)
		search = filepath.Join(directory, "*.usx")
	} else if fsType == `audio` {
		directory = filepath.Join(os.Getenv("FCBH_DATASET_FILES"), bibleId, filesetId)
		if tType == `OT` {
			search = filepath.Join(directory, "A*.*")
		} else {
			search = filepath.Join(directory, "B*.*")
		}
	}
	//fmt.Println("search:", tType, search)
	var files []InputFiles
	files, status = Glob(ctx, directory, search)
	if status.IsErr {
		return files, status
	}
	var inputFiles []InputFiles
	if fsType == `text_plain` {
		inputFiles = files
	} else if fsType == `text_usx` {
		for _, file := range files {
			file.BookId = file.Filename[3:6]
			if testament.Has(tType, file.BookId) {
				inputFiles = append(inputFiles, file)
			}
		}
	} else if fsType == `audio` {
		for _, file := range files {
			file.BookId, file.Chapter, status = ParseAudioFilename(ctx, file.Filename)
			if status.IsErr {
				return inputFiles, status
			}
			if testament.Has(tType, file.BookId) {
				inputFiles = append(inputFiles, file)
			}
		}
	}
	return inputFiles, status
}

func Glob(ctx context.Context, directory string, search string) ([]InputFiles, dataset.Status) {
	var results []InputFiles
	var status dataset.Status
	if search != `` {
		files, err := filepath.Glob(search)
		if err != nil {
			status = log.Error(ctx, 500, err, `Error reading directory`)
			return results, status
		}
		for _, file := range files {
			var input InputFiles
			input.Directory = directory
			input.Filename = filepath.Base(file)
			results = append(results, input)
		}
	}
	return results, status
}

func ParseAudioFilename(ctx context.Context, filename string) (string, int, dataset.Status) {
	var bookId string
	var chapterNum int
	var status dataset.Status
	chapter, err := strconv.Atoi(filename[6:8])
	if err != nil {
		status = log.Error(ctx, 500, err, `Error convert chapter to int`, filename[6:8])
		return bookId, chapterNum, status
	}
	book := strings.Trim(filename[9:21], `_`)
	bookId = db.USFMBookId(ctx, book)
	return bookId, chapter, status
}