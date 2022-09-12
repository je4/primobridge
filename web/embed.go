package web

import "embed"

//go:embed template/kisten.gohtml
//go:embed template/viewer.gohtml
var TemplateFS embed.FS

//go:embed static/3dthumb/jpg/*
//go:embed static/3djson/*
//go:embed static/js/*
//go:embed static/img/*
//go:embed static/kistendata.json
var StaticFS embed.FS
