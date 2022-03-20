package bridge

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func (s *Server) SearchThumbnailHandler(w http.ResponseWriter, r *http.Request) {
	var vars = mux.Vars(r)
	var searchKey string

	searchKey = vars["searchKey"]

	png, mime, err := s.mapper.GetImage(searchKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("cannot generate thumbnail: %v", err)))
		return
	}
	w.Header().Set("Content-type", mime)
	w.Write(png)
}
