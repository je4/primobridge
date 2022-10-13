package bridge

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func (s *Server) SystematikHandler(w http.ResponseWriter, r *http.Request) {

	var vars = mux.Vars(r)
	var systematik string

	systematik = vars["systematik"]

	hierarchy, err := s.mapper.GetSystematikHierarchy(systematik)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot get systematik hierarchy: %v", err), http.StatusInternalServerError)
		return
	}

	s.InitTemplates() // todo: remove this
	tpl, ok := s.templates["viewer.gohtml"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("cannot find viewer.gohtml"))
	}
	if err := tpl.Execute(w, struct {
		Signature, DocID, Barcode, Box, Systematik string
		Hierarchy                                  map[string]map[string]Class
	}{
		Systematik: systematik,
		Hierarchy:  hierarchy,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("cannot execute viewer.gohtml: %v", err)))
	}
}
