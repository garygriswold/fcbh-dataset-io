package db

type Ident struct {
	DatasetId      int
	BibleId        string
	AudioFilesetId string
	TextFilesetId  string
	TextSource     string
	LanguageIso    string
	VersionCode    string
	LanguageId     int
	RolvId         int
	Alphabet       string
	LanguageName   string
	VersionName    string
}

type Script struct {
	ScriptId      int
	DatasetId     int
	BookId        string
	ChapterNum    int
	AudioFile     string
	ScriptNum     string
	UsfmStyle     string
	Person        string
	Actor         string
	VerseNum      int
	VerseStr      string
	ScriptText    string
	ScriptTexts   []string
	ScriptBeginTS float64
	ScriptEndTS   float64
	ScriptMFCC    []byte
	MFCCRows      int
	MFCCCols      int
}

type Word struct {
	WordId          int
	WordSeq         int
	VerseNum        int
	Ttype           string
	Word            string
	WordBeginTS     float64
	WordEndTS       float64
	WordMfcc        []byte
	MFCCRows        int
	MFCCCols        int
	MFCCNorm        []byte
	MFCCNormRows    int
	MFCCNormCols    int
	WordEnc         []byte
	SrcWordEnc      []byte
	WordMultiEnc    []byte
	SrcWordMultiEnc []byte
}