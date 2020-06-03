package scoped

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
)

type JsonTagger struct {
	scope  string
	buffer *bytes.Buffer
}

func (j *JsonTagger) AsJSON() []byte {
	return j.buffer.Bytes()
}

func New(scope string, data interface{}) *JsonTagger {
	out := &JsonTagger{
		scope:  scope,
		buffer: bytes.NewBuffer(nil),
	}
	return out.scopeAll(data)
}

func (j *JsonTagger) Read(p []byte) (n int, err error) {
	return j.buffer.Read(p)
}

func (j *JsonTagger) Write(p []byte) (n int, err error) {
	return j.buffer.Write(p)
}

func Handler(scope string, handler func(w http.ResponseWriter, r *http.Request) interface{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		out := New(scope, handler(w, r))
		w.Header().Set("Content-Type", "application/json")
		w.Write(out.buffer.Bytes())
	})
}

func (j *JsonTagger) scopeAll(data interface{}) *JsonTagger {
	val := reflect.ValueOf(data)

	if val.Kind() == reflect.Slice {
		slicer := make([]map[string]interface{}, 0)
		for i := 0; i < val.Len(); i++ {
			v := val.Index(i).Interface()
			slicer = append(slicer, j.extractData(v))
		}
		j.marshalWrite(slicer)
		return j
	}
	j.marshalWrite(j.extractData(data))
	return j
}

func (j *JsonTagger) marshalWrite(data interface{}) {
	jD, _ := json.Marshal(data)
	j.Write(jD)
}

func (j *JsonTagger) extractData(data interface{}) map[string]interface{} {
	val := reflect.ValueOf(data)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	thisData := make(map[string]interface{})

	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tagVal := typeField.Tag

		tag := tagVal.Get("scope")
		tags := strings.Split(tag, ",")

		jTags := tagVal.Get("json")
		jsonTag := strings.Split(jTags, ",")

		if len(jsonTag) == 0 {
			continue
		}

		if jsonTag[0] == "" || jsonTag[0] == "-" {
			continue
		}

		if len(jsonTag) == 2 {
			if jsonTag[1] == "omitempty" && valueField.Interface() == "" {
				continue
			}
		}

		if tag == "" {
			thisData[jsonTag[0]] = valueField.Interface()
			continue
		}

		if forTag(tags, j.scope) {
			thisData[jsonTag[0]] = valueField.Interface()
		}
	}
	return thisData
}

func forTag(tags []string, scope string) bool {
	for _, v := range tags {
		if v == scope {
			return true
		}
	}
	return false
}
