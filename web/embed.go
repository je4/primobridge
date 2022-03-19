package web

import "embed"

//go:embed template/index.gohtml
var TemplateFS embed.FS

//go:embed static/3dthumb/jpg/*
var StaticFS embed.FS
