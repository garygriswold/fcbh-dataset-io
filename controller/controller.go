package controller

import (
	"context"
	"dataset"
	"dataset/db"
	"dataset/encode"
	"dataset/fetch"
	"dataset/input"
	log "dataset/logger"
	"dataset/match"
	"dataset/output"
	"dataset/read"
	"dataset/request"
	"dataset/speech_to_text"
	"time"
)

type Controller struct {
	ctx         context.Context
	yamlRequest []byte
	req         request.Request
	user        fetch.DBPUser
	ident       db.Ident
	database    db.DBAdapter
}

func NewController(yamlContent []byte) Controller {
	var c Controller
	c.ctx = context.Background()
	c.yamlRequest = yamlContent
	return c
}

func (c *Controller) Process() (string, dataset.Status) {
	var start = time.Now()
	log.SetLevel(log.LOGDEBUG)
	log.SetOutput(c.ctx, `stderr`)
	log.Debug(c.ctx)
	var filename, status = c.processSteps()
	if status.IsErr {
		filename = c.outputStatus(status)
	}
	log.Info(c.ctx, "Duration", time.Since(start))
	log.Debug(c.ctx)
	return filename, status
}

func (c *Controller) processSteps() (string, dataset.Status) {
	var filename string
	var status dataset.Status
	// Decode YAML Request File
	reqDecoder := request.NewRequestDecoder(c.ctx)
	c.req, status = reqDecoder.Process(c.yamlRequest)
	if status.IsErr {
		return filename, status
	}
	var yaml string
	// Stuff YAML request into context
	yaml, status = reqDecoder.Encode(c.req)
	if status.IsErr {
		return filename, status
	}
	c.ctx = context.WithValue(context.Background(), `request`, yaml)
	// Get User
	c.user, status = fetch.GetDBPUser()
	if status.IsErr {
		return filename, status
	}
	// Open Database
	c.database, status = db.NewerDBAdapter(c.ctx, c.req.IsNew, c.user.Username, c.req.DatasetName)
	if status.IsErr {
		return filename, status
	}
	defer c.database.Close()
	// Fetch Ident Data from DBP
	c.ident, status = c.fetchData()
	if status.IsErr {
		return filename, status
	}
	// Collect Text Input
	var textFiles []input.InputFile
	textFiles, status = c.collectTextInput()
	if status.IsErr {
		return filename, status
	}
	// Collect Audio Input
	var audioFiles []input.InputFile
	audioFiles, status = c.collectAudioInput()
	if status.IsErr {
		return filename, status
	}
	// Update Ident Table
	status = input.UpdateIdent(c.database, &c.ident, textFiles, audioFiles)
	if status.IsErr {
		return filename, status
	}
	// Read Text Data
	status = c.readText(textFiles)
	if status.IsErr {
		return filename, status
	}
	// Speech to Text
	status = c.speechToText(audioFiles)
	if status.IsErr {
		return filename, status
	}
	// Timestamps
	status = c.timestamps(audioFiles)
	if status.IsErr {
		return filename, status
	}
	// Encode Audio
	status = c.encodeAudio(audioFiles)
	if status.IsErr {
		return filename, status
	}
	// Encode Text
	status = c.encodeText()
	if status.IsErr {
		return filename, status
	}
	// Compare
	if c.req.Compare.BaseDataset != `` {
		filename, status = c.matchText()
		return filename, status // return whether success or not
	}
	// Prepare output
	if c.req.OutputFormat.Sqlite {
		filename = c.database.DatabasePath
	} else {
		filename, status = c.output()
	}
	return filename, status
}

func (c *Controller) fetchData() (db.Ident, dataset.Status) {
	//var ident db.Ident
	var status dataset.Status
	var info fetch.BibleInfoType
	client := fetch.NewAPIDBPClient(c.ctx, c.req.BibleId)
	info, status = client.BibleInfo()
	if status.IsErr {
		return c.ident, status
	}
	client.FindFilesets(&info, c.req.AudioData.BibleBrain, c.req.TextData.BibleBrain, c.req.Testament)
	download := fetch.NewAPIDownloadClient(c.ctx, c.req.BibleId)
	status = download.Download(info)
	if status.IsErr {
		return c.ident, status
	}
	//} else {
	//	var msg = make([]string, 0, 10)
	//	msg = append(msg, "Requested Fileset is not available")
	//	for _, rec := range info.DbpProd.Filesets {
	//		msg = append(msg, fmt.Sprintf("%+v", rec))
	//	}
	//	status.IsErr = true
	//	status.Status = 400
	//	status.Message = strings.Join(msg, "\n")
	//	return info, status
	//}
	c.ident, status = c.database.SelectIdent()
	if status.IsErr {
		return c.ident, status
	}
	c.ident = client.UpdateIdent(c.ident, info)
	textType := c.req.TextData.BibleBrain.TextType()
	if textType != request.TextNone {
		c.ident.TextSource = textType
	}
	status = c.database.InsertReplaceIdent(c.ident)
	return c.ident, status
}

func (c *Controller) collectTextInput() ([]input.InputFile, dataset.Status) {
	var files []input.InputFile
	var status dataset.Status
	bb := c.req.TextData.BibleBrain
	if bb.TextPlain || bb.TextPlainEdit || bb.TextUSXEdit {
		files, status = input.DBPDirectory(c.ctx, c.req.BibleId, c.ident.TextSource, c.ident.TextOTId,
			c.ident.TextNTId, c.req.Testament)
	} else if c.req.TextData.File != `` {
		files, status = input.FileInput(c.ctx, c.req.TextData.File, c.req.Testament)
	} else if c.req.TextData.AWSS3 != `` {
		files, status = input.AWSS3Input(c.ctx, c.req.TextData.AWSS3, c.req.Testament)
	} else if c.req.TextData.POST != `` {
		files, status = input.PostInput(c.ctx, c.req.UploadedFile, c.req.TextData.POST, c.req.Testament)
	}
	return files, status
}

func (c *Controller) collectAudioInput() ([]input.InputFile, dataset.Status) {
	var files []input.InputFile
	var status dataset.Status
	bb := c.req.AudioData.BibleBrain
	if bb.MP3_64 || bb.MP3_16 || bb.OPUS {
		bibleId := c.req.BibleId
		files, status = input.DBPDirectory(c.ctx, bibleId, request.Audio, c.ident.AudioOTId,
			c.ident.AudioNTId, c.req.Testament)
	} else if c.req.AudioData.File != `` {
		files, status = input.FileInput(c.ctx, c.req.AudioData.File, c.req.Testament)
	} else if c.req.AudioData.AWSS3 != `` {
		files, status = input.AWSS3Input(c.ctx, c.req.AudioData.AWSS3, c.req.Testament)
	} else if c.req.AudioData.POST != `` {
		files, status = input.PostInput(c.ctx, c.req.UploadedFile, c.req.AudioData.POST, c.req.Testament)
	}
	return files, status
}

func (c *Controller) readText(textFiles []input.InputFile) dataset.Status {
	var status dataset.Status
	if len(textFiles) == 0 {
		return status
	}
	if textFiles[0].MediaType == request.TextUSXEdit {
		reader := read.NewUSXParser(c.database)
		status = reader.ProcessFiles(textFiles)
		if status.IsErr {
			return status
		}
	} else if textFiles[0].MediaType == request.TextPlainEdit {
		reader := read.NewDBPTextEditReader(c.database, c.req)
		status = reader.Process()
		if status.IsErr {
			return status
		}
	} else if textFiles[0].MediaType == request.TextPlain {
		reader := read.NewDBPTextReader(c.database, c.req.Testament)
		status = reader.ProcessFiles(textFiles)
		if status.IsErr {
			return status
		}
	} else if textFiles[0].MediaType == request.TextScript { //`text_script` {
		reader := read.NewScriptReader(c.database)
		status = reader.ProcessFiles(textFiles)
		if status.IsErr {
			return status
		}
	} else {
		return status // This is not an error, it is nothing to do
	}
	if c.req.Detail.Words {
		words := read.NewWordParser(c.database)
		status = words.Parse()
	}
	return status
}

func (c *Controller) speechToText(audioFiles []input.InputFile) dataset.Status {
	var status dataset.Status
	bibleId := c.req.BibleId
	var whisperModel = c.req.TextData.SpeechToText.Whisper.Model.String()
	if whisperModel != `` {
		var whisper = speech_to_text.NewWhisper(bibleId, c.database, whisperModel)
		status = whisper.ProcessFiles(audioFiles)
		if status.IsErr {
			return status
		}
		c.ident.TextSource = request.TextSTT
		c.database.UpdateIdent(c.ident)
	}
	return status
}

func (c *Controller) timestamps(audioFiles []input.InputFile) dataset.Status {
	var status dataset.Status
	if c.req.Timestamps.BibleBrain {
		var filesetIds []string
		if c.ident.AudioOTId != `` {
			filesetIds = append(filesetIds, c.ident.AudioOTId)
		}
		if c.ident.AudioNTId != `` {
			filesetIds = append(filesetIds, c.ident.AudioNTId)
		}
		for _, filesetId := range filesetIds {
			api := fetch.NewAPIDBPTimestamps(c.database, filesetId)
			// Load returns bool, which could be used to invoke aeneas
			_, status = api.LoadTimestamps(c.req.Testament)
			if status.IsErr {
				return status
			}
		}
	} else if c.req.Timestamps.Aeneas {
		bibleId := c.req.BibleId
		aeneas := encode.NewAeneas(c.ctx, c.database, bibleId, c.ident.LanguageISO, c.req.Detail)
		status = aeneas.ProcessFiles(audioFiles)
		if status.IsErr {
			return status
		}
	}
	return status
}

func (c *Controller) encodeAudio(audioFiles []input.InputFile) dataset.Status {
	var status dataset.Status
	bibleId := c.req.BibleId
	if c.req.AudioEncoding.MFCC {
		mfcc := encode.NewMFCC(c.ctx, c.database, bibleId, c.req.Detail, 7)
		status = mfcc.ProcessFiles(audioFiles)
		if status.IsErr {
			return status
		}
	}
	return status
}

func (c *Controller) encodeText() dataset.Status {
	var status dataset.Status
	if c.req.TextEncoding.FastText {
		fast := encode.NewFastText(c.ctx, c.database)
		status = fast.Process()
	}
	return status
}

func (c *Controller) matchText() (string, dataset.Status) {
	var filename string
	var status dataset.Status
	compare := match.NewCompare(c.ctx, c.user, c.req.Compare.BaseDataset, c.database, c.req.Testament, c.req.Compare.CompareSettings)
	if c.ident.TextSource == request.TextSTT {
		filename, status = compare.CompareChapters()
	} else {
		filename, status = compare.Process()
	}
	return filename, status
}

func (c *Controller) output() (string, dataset.Status) {
	var filename string
	var status dataset.Status
	var out = output.NewOutput(c.ctx, c.database, c.req.DatasetName, false, false)
	var records []any
	var meta []output.Meta
	if c.req.Detail.Lines {
		records, meta = out.PrepareScripts()
	} else {
		records, meta = out.PrepareWords()
	}
	if c.req.OutputFormat.CSV {
		filename, status = out.WriteCSV(records, meta)
		if status.IsErr {
			return filename, status
		}
	} else if c.req.OutputFormat.JSON {
		filename, status = out.WriteJSON(records, meta)
		if status.IsErr {
			return filename, status
		}
	}
	return filename, status
}

func (c *Controller) outputStatus(status dataset.Status) string {
	var filename string
	var status2 dataset.Status
	var out = output.NewOutput(c.ctx, db.DBAdapter{}, c.req.DatasetName, false, false)
	if c.req.OutputFormat.CSV {
		filename, status2 = out.CSVStatus(status, true)
	} else if c.req.OutputFormat.JSON {
		filename, status2 = out.JSONStatus(status, true)
	} else {
		filename = status.String()
	}
	if status2.IsErr {
		filename = status2.String()
	}
	return filename
}
