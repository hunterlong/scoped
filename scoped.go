package scoped

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
)

type Scoped struct {
	scope  string
	buffer *bytes.Buffer
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request) interface{}

func (j *Scoped) AsJSON() []byte {
	return j.buffer.Bytes()
}

func (j *Scoped) HttpResponder(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(j.buffer.Bytes())
}

func NewContext(ctx context.Context, data interface{}) *Scoped {
	scope, ok := ctx.Value("scope").(string)
	if !ok {
		return nil
	}
	return New(scope, data)
}

func New(scope string, data interface{}) *Scoped {
	out := &Scoped{
		scope:  scope,
		buffer: bytes.NewBuffer(nil),
	}
	return out.scopeAll(data)
}

func (j *Scoped) Read(p []byte) (n int, err error) {
	return j.buffer.Read(p)
}

func (j *Scoped) Write(p []byte) (n int, err error) {
	return j.buffer.Write(p)
}

func (j *Scoped) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write(j.buffer.Bytes())
}

func Handler(handler HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope, ok := r.Context().Value("scope").(string)
		if !ok {
			handler(w, r)
			return
		}
		data := handler(w, r)
		out := New(scope, data)
		w.Header().Set("Content-Type", "application/json")
		w.Write(out.buffer.Bytes())
	})
}

func (j *Scoped) scopeAll(data interface{}) *Scoped {
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

func (j *Scoped) marshalWrite(data interface{}) {
	jD, _ := json.Marshal(data)
	j.Write(jD)
}

func (j *Scoped) extractData(data interface{}) map[string]interface{} {
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
