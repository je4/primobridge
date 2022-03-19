package bridge

import (
	"github.com/gorilla/mux"
	"net/http"
)

func (s *Server) SearchThumbnailHandler(w http.ResponseWriter, r *http.Request) {
	var vars = mux.Vars(r)

}
