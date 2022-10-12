package bridge

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type KistePoint struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

type KisteLevel struct {
	Indent float64 `json:"indent"`
	Boxes  int     `json:"boxes"`
}

type Kiste struct {
	Point []*KistePoint       `json:"point"`
	ASide string              `json:"aside"`
	Area  string              `json:"area"`
	Level map[int]*KisteLevel `json:"level"`
}

type KisteData map[string]Kiste

func (s *Server) KistenlisteHandler(w http.ResponseWriter, r *http.Request) {

	kdh, err := s.staticFS.Open("kistendata.json")
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read kistendata.json: %v", err), 500)
		return
	}
	defer kdh.Close()
	var kisteData = KisteData{}
	dec := json.NewDecoder(kdh)
	if err := dec.Decode(&kisteData); err != nil {
		http.Error(w, fmt.Sprintf("cannot decode kistendata.json: %v", err), 500)
		return
	}

	s.InitTemplates() // todo: remove this
	tpl, ok := s.templates["kisten.gohtml"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("cannot find kisten.gohtml"))
	}
	var kisten = map[string]struct {
		Img  string
		JSON string
	}{}
	for regal, kiste := range kisteData {
		for level, row := range kiste.Level {
			for i := 1; i <= row.Boxes; i++ {
				for _, k := range []string{
					fmt.Sprintf("%s%d%02da", regal, level, i),
					fmt.Sprintf("%s%d%02db", regal, level, i),
				} {
					img := ""
					json := ""
					if f, err := s.staticFS.Open(fmt.Sprintf("3dthumb/jpg/%s.jpg", k)); err == nil {
						f.Close()
						img = fmt.Sprintf("static/3dthumb/jpg/%s.jpg", k)
					}
					if f, err := s.staticFS.Open(fmt.Sprintf("3djson/%s.json", k)); err == nil {
						b, err := io.ReadAll(f)
						if err != nil {
							http.Error(w, fmt.Sprintf("cannot read %s: %v", fmt.Sprintf("3djson/%s.json", k), err), 500)
							return
						}
						json = strings.TrimSpace(string(b))
						f.Close()
					}
					kisten[k] = struct {
						Img  string
						JSON string
					}{Img: img, JSON: json}
				}
			}
		}
	}
	if err := tpl.Execute(w, struct {
		Kisten map[string]struct {
			Img  string
			JSON string
		}
	}{
		Kisten: kisten,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("cannot execute viewer.gohtml: %v", err)))
	}

}
