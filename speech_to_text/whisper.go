package speech_to_text

import (
	"bytes"
	"context"
	"dataset"
	"dataset/db"
	"dataset/input"
	log "dataset/logger"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

/*
Docs:
https://github.com/openai/whisper
Install:
pip3 install git+https://github.com/openai/whisper.git
Whisper is an open source Speech to Text program developed by OpenAI.
Executable:
/Users/gary/Library/Python/3.9/bin/whisper
*/

type Whisper struct {
	ctx     context.Context
	conn    db.DBAdapter
	bibleId string
	model   string
	lang2   string // 2 char language code
	tempDir string
}

func NewWhisper(bibleId string, conn db.DBAdapter, model string, lang2 string) Whisper {
	var w Whisper
	w.ctx = conn.Ctx
	w.conn = conn
	w.bibleId = bibleId
	w.model = model
	w.lang2 = lang2
	return w
}

func (w *Whisper) ProcessFiles(files []input.InputFile) dataset.Status {
	var status dataset.Status
	var outputFile string
	var err error
	w.tempDir, err = os.MkdirTemp(os.Getenv(`FCBH_DATASET_TMP`), "Whisper_")
	if err != nil {
		return log.Error(w.ctx, 500, err, `Error creating temp dir`)
	}
	defer os.RemoveAll(w.tempDir)
	if w.lang2 == `` {
		var iso639 db.Sil639
		iso639, status = db.FindWhisperCompatibility(w.ctx, strings.ToLower(w.bibleId[:3]))
		w.lang2 = iso639.Lang2
		log.Info(w.ctx, `Using language`, w.lang2, iso639.Name)
	}
	for _, file := range files {
		fmt.Println(`INPUT FILE:`, file)
		var timestamps []db.Timestamp
		timestamps, status = w.conn.SelectScriptTimestamps(file.BookId, file.Chapter)
		if status.IsErr {
			return status
		}
		status = w.conn.DeleteScripts(file.BookId, file.Chapter)
		if status.IsErr {
			return status
		}
		var records []db.Script
		if len(timestamps) > 0 {
			timestamps, status = w.ChopByTimestamp(file, timestamps)
			if status.IsErr {
				return status
			}
			for pieceNum, piece := range timestamps {
				fmt.Println(`VERSE PIECE:`, piece)
				inputFile := filepath.Join(w.tempDir, piece.AudioFile)
				outputFile, status = w.RunWhisper(inputFile)
				var rec db.Script
				rec, status = w.loadWhisperVerses(outputFile, file, pieceNum, piece)
				records = append(records, rec)
			}
		} else {
			outputFile, status = w.RunWhisper(file.FilePath())
			records, status = w.loadWhisperOutput(outputFile, file)
		}
		w.conn.InsertScripts(records)
		records = nil
	}
	return status
}

func (w *Whisper) RunWhisper(audioFile string) (string, dataset.Status) {
	var status dataset.Status
	whisperPath := os.Getenv(`WHISPER_EXE`)
	cmd := exec.Command(whisperPath,
		audioFile,
		`--model`, w.model,
		`--output_format`, `json`,
		`--fp16`, `False`,
		`--language`, w.lang2,
		`--word_timestamps`, `True`, // Runs about 10% faster with this off.  Should it be conditional?
		`--output_dir`, w.tempDir)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err := cmd.Run()
	if err != nil {
		status = log.Error(w.ctx, 500, err, stderrBuf.String())
		// Do not return immediately, must get std error
	}
	stderrStr := stderrBuf.String()
	if stderrStr != `` {
		log.Warn(w.ctx, `Whisper Stderr:`, stderrStr)
	}
	fileType := filepath.Ext(audioFile)
	filename := filepath.Base(audioFile)
	outputFile := filepath.Join(w.tempDir, filename[:len(filename)-len(fileType)]) + `.json`
	return outputFile, status
}
