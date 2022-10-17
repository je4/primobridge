package bridge

import (
	"context"
	"crypto/tls"
	sprig "github.com/Masterminds/sprig/v3"
	"github.com/bluele/gcache"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	dcert "github.com/je4/utils/v2/pkg/cert"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"html/template"
	"io"
	"io/fs"
	"net"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	service          string
	host, port       string
	srv              *http.Server
	linkTokenExp     time.Duration
	jwtKey           string
	jwtAlg           []string
	log              *logging.Logger
	AddrExt          string
	primoSourceData  string
	primoDeepLink    string
	siteViewerLink   string
	accessLog        io.Writer
	templates        map[string]*template.Template
	httpStaticServer http.Handler
	templateFS       fs.FS
	staticFS         fs.FS
	mapper           PrimoMapper
	dev              bool
	marcCache        gcache.Cache
}

func NewServer(service, addr, addrExt, primoSourceData, primoDeepLink string, staticFS, templateFS fs.FS, mapper PrimoMapper, log *logging.Logger, accessLog io.Writer, dev bool) (*Server, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot split address %s", addr)
	}
	/*
		extUrl, err := url.Parse(addrExt)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse external address %s", addrExt)
		}
	*/

	srv := &Server{
		service:          service,
		host:             host,
		port:             port,
		primoSourceData:  primoSourceData,
		primoDeepLink:    primoDeepLink,
		staticFS:         staticFS,
		httpStaticServer: http.FileServer(http.FS(staticFS)),
		AddrExt:          strings.TrimRight(addrExt, "/"),
		log:              log,
		accessLog:        accessLog,
		templateFS:       templateFS,
		templates:        map[string]*template.Template{},
		mapper:           mapper,
		dev:              dev,
		marcCache:        gcache.New(128).Expiration(time.Hour * 24).ARC().Build(),
	}

	return srv, srv.InitTemplates()
}

func (s *Server) InitTemplates() error {
	entries, err := fs.ReadDir(s.templateFS, ".")
	if err != nil {
		return errors.Wrapf(err, "cannot read template folder %s", "template")
	}
	for _, entry := range entries {
		name := entry.Name()
		tpl, err := template.New(name).Funcs(sprig.FuncMap()).ParseFS(s.templateFS, name)
		if err != nil {
			return errors.Wrapf(err, "cannot parse template: %s", name)
		}
		s.templates[name] = tpl
	}
	return nil
}

func (s *Server) ListenAndServe(cert, key string) (err error) {
	router := mux.NewRouter()

	router.HandleFunc("/initmapper/hydmUN.DftakhLJDNT_5", s.InitMapperHandler).
		Methods("GET")
	router.HandleFunc("/marc21/{docid}", s.Marc21Handler).
		Methods("GET").
		Name("marc21")

	router.HandleFunc("/static_images/projects/{project_id}/search_qrcode.png", s.SearchQRCodeHandler).
		Methods("GET").
		Queries(
			// "project_id", "{projectID}",
			"search_key", "{searchKey}",
			// "language", "{language}",
			"e", "{docID}",
			"bcd", "{barcode}").
		Name("qrcode")
	router.HandleFunc("/static_images/projects/{project_id}/search_thumbnail.jpg", s.SearchThumbnailHandler).
		Methods("GET").
		Queries(
			// "project_id", "{projectID}",
			"search_key", "{searchKey}",
			// "language", "{language}",
			// "e", "{docID}",
			// "bcd", "{barcode}"
		).
		Name("thumb")
	router.HandleFunc("/viewer", s.ViewerHandler).
		Methods("GET").
		Queries(
			// "project_id", "{projectID}",
			"search_key", "{searchKey}",
			// "language", "{language}",
			"e", "{docID}",
			"bcd", "{barcode}",
		).
		Name("viewer")
	router.HandleFunc("/viewer", s.SystematikHandler).
		Methods("GET").
		Queries(
			"sys", "{systematik}",
		).
		Name("viewer")
	router.HandleFunc("/3d", s.ThreeDHandler).
		Methods("GET").
		Queries(
			"box", "{box}",
		).
		Name("3d")
	router.HandleFunc("/kisten", s.KistenlisteHandler).
		Methods("GET").
		Name("kisten")
	router.PathPrefix("/static").Handler(http.StripPrefix("/static", s.httpStaticServer)).Methods("GET")

	router.Use(mux.CORSMethodMiddleware(router))
	loggedRouter := handlers.CombinedLoggingHandler(s.accessLog, handlers.ProxyHeaders(router))
	addr := net.JoinHostPort(s.host, s.port)
	s.srv = &http.Server{
		Handler: loggedRouter,
		Addr:    addr,
	}

	if cert == "auto" || key == "auto" {
		s.log.Info("generating new certificate")
		cert, err := dcert.DefaultCertificate()
		if err != nil {
			return errors.Wrap(err, "cannot generate default certificate")
		}
		s.srv.TLSConfig = &tls.Config{Certificates: []tls.Certificate{*cert}}
		s.log.Infof("starting salon digital at %v - https://%s:%v/", s.AddrExt, s.host, s.port)
		return s.srv.ListenAndServeTLS("", "")
	} else if cert != "" && key != "" {
		s.log.Infof("starting salon digital at %v - https://%s:%v/", s.AddrExt, s.host, s.port)
		return s.srv.ListenAndServeTLS(cert, key)
	} else {
		s.log.Infof("starting salon digital at %v - http://%s:%v/", s.AddrExt, s.host, s.port)
		return s.srv.ListenAndServe()
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
