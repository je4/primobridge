package web

import "embed"

//go:embed template/viewer.gohtml
var TemplateFS embed.FS

//go:embed static/3dthumb/jpg/*
//go:embed static/js/*
//go:embed static/img/*
//go:embed static/kistendata.json
var StaticFS embed.FS
