package output

import (
	"dataset"
	log "dataset/logger"
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
)

func (o *Output) JSONStatus(status dataset.Status, debug bool) (string, dataset.Status) {
	var filename string
	var errStatus dataset.Status
	file, err := os.Create(filepath.Join(os.Getenv(`FCBH_DATASET_TMP`), o.requestName+".json"))
	//file, err := os.CreateTemp(os.Getenv(`FCBH_DATASET_TMP`), o.requestName+"_*.json")
	if err != nil {
		errStatus = log.Error(o.ctx, 500, err, status.Err)
		return filename, errStatus
	}
	filename = file.Name()
	if !debug {
		status.Trace = ``
	}
	bytes, err := json.MarshalIndent(status, "", "   ")
	if err != nil {
		errStatus = log.Error(o.ctx, 500, err, status.Err)
		return filename, errStatus
	}
	_, _ = file.Write(bytes)
	_ = file.Close()
	return filename, errStatus
}

func (o *Output) CSVStatus(status dataset.Status, debug bool) (string, dataset.Status) {
	var filename string
	var errStatus dataset.Status
	file, err := os.Create(filepath.Join(os.Getenv(`FCBH_DATASET_TMP`), o.requestName+".csv"))
	//file, err := os.CreateTemp(os.Getenv(`FCBH_DATASET_TMP`), o.requestName+"_*.csv")
	if err != nil {
		errStatus = log.Error(o.ctx, 500, err, status.Err)
		return filename, errStatus
	}
	filename = file.Name()
	writer := csv.NewWriter(file)
	_ = writer.Write([]string{`Name`, `Value`})
	_ = writer.Write([]string{`is_error`, strconv.FormatBool(status.IsErr)})
	_ = writer.Write([]string{`status`, strconv.Itoa(status.Status)})
	_ = writer.Write([]string{`message`, status.Message})
	_ = writer.Write([]string{`error`, status.Err})
	_ = writer.Write([]string{`request`, status.Request})
	if debug {
		_ = writer.Write([]string{`trace`, status.Trace})
	}
	writer.Flush()
	err = writer.Error()
	if err != nil {
		errStatus = log.Error(o.ctx, 500, err, status.Err)
	}
	_ = file.Close()
	return filename, errStatus
}
