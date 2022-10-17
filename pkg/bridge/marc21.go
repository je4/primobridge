package bridge

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var marcRegexp = regexp.MustCompile("^(?P<tag>[0-9]{3})\\s(?P<ind1>.)(?P<ind2>.)(?P<sub>.+)$")

type Book struct {
	Title   string
	Authors string
	Verlag  string
	City    string
	Year    string
	ISBN    string
}

func parseMarc(marc string) *Book {
	book := &Book{}
	lines := strings.Split(strings.ReplaceAll(marc, "\r\n", "\n"), "\n")
	for _, line := range lines {
		match := marcRegexp.FindStringSubmatch(line)
		if len(match) == 0 {
			continue
		}
		var result = map[string]string{}
		for i, name := range marcRegexp.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = match[i]
			}
		}
		var sub = map[string][]string{}
		subs := strings.Split(result["sub"], "$")
		for _, s := range subs {
			s = strings.TrimSpace(s)
			if len(s) == 0 {
				continue
			}
			st := s[0:1]
			if _, ok := sub[st]; !ok {
				sub[st] = []string{}
			}
			sub[st] = append(sub[st], s[1:])
		}
		switch result["tag"] {
		case "245":
			if ts, ok := sub["a"]; ok {
				book.Title = strings.Join(ts, "; ")
			}
			if as, ok := sub["c"]; ok {
				book.Authors = strings.Join(as, "; ")
			}
		case "264":
			if xs, ok := sub["a"]; ok {
				book.City = strings.Join(xs, "; ")
			}
			if xs, ok := sub["b"]; ok {
				book.Verlag = strings.Join(xs, "; ")
			}
			if xs, ok := sub["c"]; ok {
				book.Year = strings.Join(xs, "; ")
			}
		case "020":
			if xs, ok := sub["a"]; ok {
				book.ISBN = strings.Join(xs, "; ")
			}
		}
	}
	return book
}

func (s *Server) Marc21Handler(w http.ResponseWriter, r *http.Request) {

	var vars = mux.Vars(r)

	docid := vars["docid"]

	w.Header().Set("Content-Type", "text/plain")

	if s.marcCache.Has(docid) {
		data, err := s.marcCache.Get(docid)
		if err != nil {
			http.Error(w, fmt.Sprintf("cannot get %s from cache", docid), http.StatusInternalServerError)
			return
		}
		dataBook, ok := data.(*Book)
		if !ok {
			http.Error(w, fmt.Sprintf("%s in cache not []byte", docid), http.StatusInternalServerError)
			return
		}
		dataBytes, err := json.Marshal(dataBook)
		if err != nil {
			http.Error(w, fmt.Sprintf("cannot marshal %v: %v", dataBytes, err), http.StatusInternalServerError)
			return
		}
		w.Write(dataBytes)
		return
	}
	urlString := fmt.Sprintf("https://fhnw.swisscovery.slsp.ch/primaws/rest/pub/sourceRecord?docId=alma%s&vid=41SLSP_FNW:VU1&recordOwner=41SLSP_NETWORK&lang=de", url.QueryEscape(docid))
	resp, err := http.Get(urlString)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot query swisscovery: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	dBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "cannot read result of swisscovery query", http.StatusInternalServerError)
		return
	}
	dataBook := parseMarc(string(dBytes))
	if err := s.marcCache.Set(docid, dataBook); err != nil {
		http.Error(w, "cannot write result to cache", http.StatusInternalServerError)
		return
	}
	dataBytes, err := json.Marshal(dataBook)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot marshal %v: %v", dataBytes, err), http.StatusInternalServerError)
		return
	}
	w.Write(dataBytes)
}
