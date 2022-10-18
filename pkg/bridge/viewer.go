package bridge

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

func (s *Server) ViewerHandler(w http.ResponseWriter, r *http.Request) {

	var vars = mux.Vars(r)
	var searchKey, docID, barcode string

	searchKey = vars["searchKey"]
	docID = vars["docID"]
	barcode = vars["barcode"]

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Max-Age", "86400")

	_, _, box, err := s.mapper.GetData(searchKey)
	if err != nil {
		box = "info"
		//		w.WriteHeader(http.StatusInternalServerError)
		//		w.Write([]byte(fmt.Sprintf("cannot find box for %s", searchKey)))
		//		return
	}
	box = strings.ReplaceAll(box, "_", "")
	systematik, err := s.mapper.GetSystematik(box)
	if err != nil {
		systematik = ""
	}
	hierarchy, err := s.mapper.GetSystematikHierarchy(systematik)
	if err != nil {
		//http.Error(w, fmt.Sprintf("cannot get systematik hierarchy: %v", err), http.StatusInternalServerError)
		//return
	}

	if s.dev {
		s.InitTemplates()
	}
	tpl, ok := s.templates["viewer.gohtml"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("cannot find viewer.gohtml"))
	}
	if err := tpl.Execute(w, struct {
		Signature, DocID, Barcode, Box, Systematik string
		Hierarchy                                  map[string]map[string]Class
	}{
		Signature:  searchKey,
		DocID:      docID,
		Barcode:    barcode,
		Box:        box,
		Systematik: systematik,
		Hierarchy:  hierarchy,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("cannot execute viewer.gohtml: %v", err)))
	}
}
