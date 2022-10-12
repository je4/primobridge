package bridge

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

func (s *Server) ThreeDHandler(w http.ResponseWriter, r *http.Request) {

	var vars = mux.Vars(r)
	var box = vars["box"]

	s.InitTemplates() // todo: remove this
	tpl, ok := s.templates["3d.gohtml"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("cannot find 3d.gohtml"))
	}
	if err := tpl.Execute(w, struct {
		Signature, DocID, Barcode, Box string
	}{
		Box: strings.ReplaceAll(box, "_", ""),
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("cannot execute 3d.gohtml: %v", err)))
	}
}
