package controller

import (
	"dataset"
	"dataset/db"
	"dataset/fetch"
	"dataset/read"
	"fmt"
	"log"
)

type Controller struct {
	request dataset.RequestType
}

func NewController(request dataset.RequestType) *Controller {
	var c Controller
	c.request = request
	return &c
}

func (c *Controller) Process() {
	textSource := string(c.request.TextSource)
	var databaseName = c.request.BibleId + "_" + textSource + ".db"
	db.DestroyDatabase(databaseName)
	var database = db.NewDBAdapter(databaseName)
	var info = c.fetchMetaData()
	fmt.Println("INFO", info)
	database.InsertIdent(info.BibleId, ``, ``, info.LanguageISO, info.VersionCode, textSource,
		info.LanguageId, info.RolvId, info.Alphabet.Alphabet, info.LanguageName, info.VersionName)
	c.readText(database)
}

func (c *Controller) fetchMetaData() fetch.BibleInfoType {
	req := c.request
	client := fetch.NewDBPAPIClient(req.BibleId)
	var info = client.BibleInfo()
	return info
}

func (c *Controller) fetchAudio() {

}

func (c *Controller) fetchText() {
	// Wait on this until talking to Brad. OR should I put in API
}

func (c *Controller) readAudio() {

}

func (c *Controller) readText(database db.DBAdapter) {
	switch c.request.TextSource {
	case dataset.USXEDIT:
		read.ReadUSXEdit(database, c.request.BibleId, c.request.Testament)
	case dataset.TEXTEDIT:
		reader := read.NewDBPTextEditReader(c.request.BibleId, database)
		reader.Process(c.request.Testament)
	case dataset.DBPTEXT:
		reader := read.NewDBPTextReader(database)
		reader.ProcessDirectory(c.request.BibleId, c.request.Testament)
	case dataset.SCRIPT:
		reader := read.NewScriptReader(database)
		file := reader.FindFile(c.request.BibleId)
		reader.Read(file)
	default:
		log.Println("Error: Could not process ", c.request.TextSource)
	}
	if c.request.TextDetail == dataset.WORDS || c.request.TextDetail == dataset.BOTH {
		words := read.NewWordParser(database)
		words.Parse()
	}
}

func (c *Controller) encodeAudio() {

}

func (c *Controller) encodeText() {

}

func (c *Controller) output() {

}
