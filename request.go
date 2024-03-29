package dataset

type SourceType string

const (
	API SourceType = "API"
	CLI SourceType = "CLI"
)

type TestamentType string

const (
	NT                   TestamentType = "NT"
	OT                   TestamentType = "OT"
	ONT                  TestamentType = "ONT"
	DefaultTestamentType               = NT
)

type AudioSourceType string

const (
	MP3                    AudioSourceType = "MP3"
	DefaultAudioSourceType                 = MP3
)

type TextDetailType string

const (
	LINES                 TextDetailType = "LINES"
	WORDS                 TextDetailType = "WORDS"
	BOTH                  TextDetailType = "BOTH"
	DefaultTextDetailType                = BOTH
)

type TextSourceType string

const (
	SCRIPT                TextSourceType = "SCRIPT"
	DBPTEXT               TextSourceType = "DBPTEXT"
	TEXTEDIT              TextSourceType = "TEXTEDIT"
	USXEDIT               TextSourceType = "USXEDIT"
	DefaultTextSourceType                = USXEDIT
)

type AudioEncodingType string

const (
	MFCC                     AudioEncodingType = "MFCC"
	DefaultAudioEncodingType                   = MFCC
)

type TextEncodingType string

const (
	FASTTEXT                TextEncodingType = "FASTTEXT"
	DefaultTextEncodingType                  = FASTTEXT
)

type OutputFormatType string

const (
	JSON                OutputFormatType = "JSON"
	PANDAS              OutputFormatType = "PANDAS"
	CSV                 OutputFormatType = "CSV"
	SQLITE              OutputFormatType = "SQLITE"
	DefaultOutputFormat                  = JSON
)

type RequestType struct {
	Email         string
	BibleId       string
	Source        SourceType
	Testament     TestamentType
	AudioSource   AudioSourceType
	TextDetail    TextDetailType
	TextSource    TextSourceType
	AudioEncoding AudioEncodingType
	TextEncoding  TextEncodingType
	OutputForm    OutputFormatType
}
