package parser

import (
	"github.com/evanphx/json-patch"
	"testing"
)

func TestJSONPatch(t *testing.T) {
	original := []byte(`{ "type": "test" }`)
	patchJSON := []byte(`[
		{ "op": "add", "path": "/type", "value": "a" }	
	]`)

	patch, err := jsonpatch.DecodePatch(patchJSON)
	if err != nil {
		panic(err)
	}

	modified, err := patch.Apply(original)
	if err != nil {
		panic(err)
	}

	t.Logf("Original document: %s\n", original)
	t.Logf("Modified document: %s\n", modified)
}
