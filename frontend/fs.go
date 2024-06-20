package frontend

import "embed"

//go:embed dist/*
var BuiltSite embed.FS
