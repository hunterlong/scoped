package scoped_test

import (
	"encoding/json"
	"github.com/hunterlong/scoped"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Example struct {
	ID           int64  `json:"id"`
	AdminOnlyStr string `json:"admin_only" scope:"admin"`
	UserOnlyStr  string `json:"user_only,omitempty" scope:"user"`
	BothStr      string `json:"both,omitempty" scope:"user,admin"`
	OmitStr      string `json:"omiter,omitempty" scope:"user,admin"`
	Hidden       string `json:"-"`
	All          string `json:"all,omitempty"`
}

var example = &Example{
	ID:           1,
	AdminOnlyStr: "im an admin",
	UserOnlyStr:  "im a user",
	BothStr:      "im on both",
	Hidden:       "cant see this",
	All:          "should always be known",
}

var example2 = &Example{
	ID:           2,
	AdminOnlyStr: "im an admin",
	UserOnlyStr:  "im a user",
	BothStr:      "im on both",
	Hidden:       "cant see this",
	All:          "should always be known",
}

var examples = []*Example{example, example2}

func TestNew(t *testing.T) {
	out := scoped.New("admin", example)
	g := decode(out.AsJSON())
	assert.Equal(t, int64(1), g.ID)
	assert.Equal(t, "im an admin", g.AdminOnlyStr)
	assert.Empty(t, g.UserOnlyStr)
	assert.Equal(t, "im on both", g.BothStr)
	assert.Empty(t, g.Hidden)

	out = scoped.New("user", example)
	g = decode(out.AsJSON())
	assert.Equal(t, int64(1), g.ID)
	assert.Equal(t, "im a user", g.UserOnlyStr)
	assert.Empty(t, g.AdminOnlyStr)
	assert.Equal(t, "im on both", g.BothStr)
	assert.Empty(t, g.Hidden)

	out = scoped.New("user", examples)
	assert.Len(t, decodeMulti(out.AsJSON()), 2)
}

func decode(val []byte) Example {
	var e Example
	json.Unmarshal(val, &e)
	return e
}

func decodeMulti(val []byte) []Example {
	var e []Example
	json.Unmarshal(val, &e)
	return e
}

func exampleHandler(w http.ResponseWriter, r *http.Request) interface{} {
	return example
}

func TestHttp(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := scoped.Handler("admin", exampleHandler)
	handler.ServeHTTP(rr, req)

	d, _ := ioutil.ReadAll(rr.Body)

	example := decode(d)

	assert.Equal(t, int64(1), example.ID)
	assert.Equal(t, "im an admin", example.AdminOnlyStr)
	assert.Empty(t, example.UserOnlyStr)
	assert.Empty(t, example.Hidden)
	assert.Equal(t, "im on both", example.BothStr)
	assert.Empty(t, example.OmitStr)

	assert.Equal(t, 200, rr.Code)
}
