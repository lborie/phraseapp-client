package main

import "encoding/json"

var BUILD_INFO string

type version struct {
	BuiltAt      string `json:"built_at,omitempty"`
	Version      string `json:"version,omitempty"`
	Changes      bool   `json:"changes,omitempty"`
	Hostname     string `json:"hostname,omitempty"`
	User         string `json:"user,omitempty"`
	Revision     string `json:"revision,omitempty"`
	DocsRevision string `json:"docs_revision,omitempty"`
}

func buildInfo() (v *version, err error) {
	return v, json.Unmarshal([]byte(BUILD_INFO), &v)
}
