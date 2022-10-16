package bridge

import (
	"fmt"
	"net/http"
)

func (s *Server) InitMapperHandler(w http.ResponseWriter, r *http.Request) {
	if err := s.mapper.Init(); err != nil {
		http.Error(w, fmt.Sprintf("cannot load mapper data: %v", err), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("ok"))
}
