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

	_, _, box, err := s.mapper.GetData(searchKey)
	if err != nil {
		box = "info"
		//		w.WriteHeader(http.StatusInternalServerError)
		//		w.Write([]byte(fmt.Sprintf("cannot find box for %s", searchKey)))
		//		return
	}

	s.InitTemplates() // todo: remove this
	tpl, ok := s.templates["viewer.gohtml"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("cannot find viewer.gohtml"))
	}
	if err := tpl.Execute(w, struct {
		Signature, DocID, Barcode, Box string
	}{
		Signature: searchKey,
		DocID:     docID,
		Barcode:   barcode,
		Box:       strings.ReplaceAll(box, "_", ""),
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("cannot execute viewer.gohtml: %v", err)))
	}
}
