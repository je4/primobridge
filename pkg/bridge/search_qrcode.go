package bridge

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/skip2/go-qrcode"
	"net/http"
	"strings"
)

func (s *Server) SearchQRCodeHandler(w http.ResponseWriter, r *http.Request) {
	var vars = mux.Vars(r)
	var urlStr = strings.ReplaceAll(s.PrimoDeepLink, "{DOCID}", vars["docID"])
	var png []byte
	png, err := qrcode.Encode(urlStr, qrcode.Medium, 130)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("cannot generate qr code: %v", err)))
		return
	}
	w.Header().Add("Content-type", "image/png")
	w.Write(png)
}
