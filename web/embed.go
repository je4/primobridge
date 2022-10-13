package web

import "embed"

//go:embed template/*.gohtml
var TemplateFS embed.FS

//go:embed static/3dthumb/jpg/*
//go:embed static/3djson/*
//go:embed static/js/*
//go:embed static/img/*
//go:embed static/css/*
//go:embed static/fonts/*
//go:embed static/bootstrap/*
//go:embed static/kistendata.json
var StaticFS embed.FS
