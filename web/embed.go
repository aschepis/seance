package web

import "embed"

// FrontendFS embeds the built frontend assets.
//
//go:embed dist/*
var FrontendFS embed.FS
