package bridge

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func (s *Server) SearchQRCodeHandler(w http.ResponseWriter, r *http.Request) {
	var vars = mux.Vars(r)

	var searchKey, docID, barcode string

	searchKey = vars["searchKey"]
	docID = vars["docID"]
	barcode = vars["barcode"]

	png, mime, err := s.mapper.GetBarcode(searchKey, docID, barcode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("cannot generate barcode: %v", err)))
		return
	}
	w.Header().Set("Content-type", mime)
	w.Write(png)
}
