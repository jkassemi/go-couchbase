package couchbase

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestViewDefinitionReduceOmission(t *testing.T) {
	view, e := json.Marshal(ViewDefinition{
		Map: `function(doc, meta){ emit(meta.id, doc); }`,
	})

	if e != nil {
		panic(e.Error())
	}

	if strings.Index(string(view), "reduce") != -1 {
		t.Fatalf("Reduce should be omitted if empty")
	}
}
