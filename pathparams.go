package aio

import (
	"reflect"
	"slices"
	"strings"

	"github.com/Piggey/aio/internal/binder"
)

// checkPathParams returns the path wildcards that no field of inType binds with the path tag.
func checkPathParams(path string, inType reflect.Type) []string {
	params := pathParams(path)
	tags := pathTags(inType)

	missing := []string{}
	for _, param := range params {
		if !slices.Contains(tags, param) {
			missing = append(missing, param)
		}
	}

	return missing
}

// pathParams returns the wildcard names of a net/http pattern path, e.g. "id" and "file" for /users/{id}/{file...}.
func pathParams(path string) []string {
	params := []string{}
	for segment := range strings.SplitSeq(path, "/") {
		name, ok := strings.CutPrefix(segment, "{")
		if !ok {
			continue
		}

		name, ok = strings.CutSuffix(name, "}")
		if !ok {
			continue
		}

		name = strings.TrimSuffix(name, "...")
		if name == "$" {
			continue // {$} only anchors the end of the path
		}

		params = append(params, name)
	}

	return params
}

// pathTags returns the path tags of the exported fields, the same ones binder.BindIn fills.
func pathTags(t reflect.Type) []string {
	tags := []string{}
	for field := range t.Fields() {
		if !field.IsExported() {
			continue
		}

		tag := field.Tag.Get(binder.TagPath)
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	return tags
}
