package request

import (
	"bytes"
	"context"
	"dataset"
	log "dataset/logger"
	"gopkg.in/yaml.v3"
	"strings"
)

type RequestDecoder struct {
	ctx    context.Context
	errors []string
}

func NewRequestDecoder(ctx context.Context) RequestDecoder {
	var r RequestDecoder
	r.ctx = ctx
	return r
}

func (r *RequestDecoder) Process(yamlRequest []byte) (Request, dataset.Status) {
	var request Request
	var status dataset.Status
	request, status = r.Decode(yamlRequest)
	if status.IsErr {
		return request, status
	}
	r.Validate(&request)
	r.Prereq(&request)
	r.Depend(request)
	if len(r.errors) > 0 {
		status.IsErr = true
		status.Status = 400
		status.Message = strings.Join(r.errors, "\n")
	}
	return request, status
}

func (r *RequestDecoder) Decode(requestYaml []byte) (Request, dataset.Status) {
	var resp Request
	var status dataset.Status
	reader := bytes.NewReader(requestYaml)
	decoder := yaml.NewDecoder(reader)
	decoder.KnownFields(true)
	err := decoder.Decode(&resp)
	if err != nil {
		status = log.Error(r.ctx, 400, err, `Error decoding YAML to request`)
	}
	resp.Testament.BuildBookMaps() // Builds Map for t.HasOT(bookId), t.HasNT(bookId)
	return resp, status
}

func (r *RequestDecoder) Encode(req Request) (string, dataset.Status) {
	var result string
	var status dataset.Status
	d, err := yaml.Marshal(&req)
	if err != nil {
		status = log.Error(r.ctx, 500, err, `Error encoding request to YAML`)
		return result, status
	}
	result = string(d)
	return result, status
}
