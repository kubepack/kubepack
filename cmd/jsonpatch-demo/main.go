package main

import (
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
)

func main() {
	p2, err := jsonpatch.DecodePatch([]byte(`[{ "op": "replace", "path": "/licenseKey", "value": "xyz" }]`))
	if err != nil {
		panic(err)
	}

	d2, err := p2.ApplyIndent([]byte(`{ "a": "b" }`), "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(d2))
}
