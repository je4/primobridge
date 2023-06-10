package mediathek

import (
	"database/sql"
	"emperror.dev/errors"
	"fmt"
	"github.com/bluele/gcache"
	"github.com/je4/primobridge/v2/pkg/bridge"
	"github.com/op/go-logging"
	"github.com/skip2/go-qrcode"
	"golang.org/x/exp/slices"
	"io"
	"io/fs"
	"net/url"
	"strings"
	"sync"
)

type MediathekMapper struct {
	sync.RWMutex
	cache          gcache.Cache
	db             *sql.DB
	logger         *logging.Logger
	boxImageFS     fs.FS
	siteViewerLink string
	boxClass       map[string]string
	classLabel     map[string]bridge.Class
	classes        []string
}

func NewMediathekMapper(db *sql.DB,
	boxImageFS fs.FS,
	siteViewerLink string,
	logger *logging.Logger) (*MediathekMapper, error) {
	mm := &MediathekMapper{
		db:             db,
		boxImageFS:     boxImageFS,
		siteViewerLink: siteViewerLink,
		logger:         logger,
		cache:          gcache.New(500).LRU().Build(),
		boxClass:       map[string]string{},
		classLabel:     map[string]bridge.Class{},
		classes:        []string{},
	}

	return mm, mm.Init()
}

func (mm *MediathekMapper) Init() error {
	mm.Lock()
	defer mm.Unlock()
	sqlstr := "SELECT box, class FROM box_class"
	rows, err := mm.db.Query(sqlstr)
	if err != nil {
		return errors.Wrapf(err, "cannot load box and class from database %s", sqlstr)
	}
	defer rows.Close()
	for rows.Next() {
		var box, class string
		if err := rows.Scan(&box, &class); err != nil {
			return errors.Wrapf(err, "cannot scan result values of %s", sqlstr)
		}
		mm.boxClass[box] = class
	}
	sqlstr = "SELECT class, label_de, label_en FROM class_label"
	sqlstrLink := "SELECT `type`, `href`, `label` FROM class_link WHERE `class`=?"

	rows2, err := mm.db.Query(sqlstr)
	if err != nil {
		return errors.Wrapf(err, "cannot load box and class from database %s", sqlstr)
	}
	defer rows2.Close()
	for rows2.Next() {
		var class, de, en string
		if err := rows2.Scan(&class, &de, &en); err != nil {
			return errors.Wrapf(err, "cannot scan result values of %s", sqlstr)
		}
		cStruct := bridge.Class{DE: de, EN: en, Links: map[string][]bridge.Link{}}
		if err := func() error {
			rows, err := mm.db.Query(sqlstrLink, class)
			if err != nil {
				return errors.Wrapf(err, "cannot load box and class from database %s", sqlstrLink)
			}
			defer rows.Close()
			for rows.Next() {
				var t, href, label string
				if err := rows.Scan(&t, &href, &label); err != nil {
					return errors.Wrapf(err, "cannot scan values of %s - [%s]", sqlstrLink, class)
				}
				if _, ok := cStruct.Links[t]; !ok {
					cStruct.Links[t] = []bridge.Link{}
				}
				cStruct.Links[t] = append(cStruct.Links[t], bridge.Link{
					Type:  t,
					HRef:  href,
					Label: label,
				})
			}
			return nil
		}(); err != nil {
			return err
		}
		mm.classLabel[class] = cStruct
		mm.classes = append(mm.classes, class)
		slices.Sort(mm.classes)
	}
	return nil
}

func (mm *MediathekMapper) GetSystematikHierarchy(sys string) (map[string]map[string]bridge.Class, error) {
	mm.RLock()
	defer mm.RUnlock()
	result := map[string]map[string]bridge.Class{}
	if len(sys) < 5 {
		sys = "00-00"
		//return nil, errors.Errorf("%s is not a valid systematik", sys)
	}
	main := sys[0:2]
	//sub := sys[3:5]
	for _, c := range mm.classes {
		cMain := c[0:2]
		cSub := c[3:5]
		if _, ok := result[cMain]; !ok {
			result[cMain] = map[string]bridge.Class{}
		}
		if cMain == main || cSub == "00" {
			/*
				if _, ok := result[cMain][cSub]; !ok {
					result[cMain][cSub] = bridge.Class{}
				}
			*/
			result[cMain][cSub] = mm.classLabel[c]
		}
	}
	return result, nil
}

func (mm *MediathekMapper) GetSystematik(box string) (sys string, err error) {
	mm.RLock()
	defer mm.RUnlock()

	var ok bool
	sys, ok = mm.boxClass[strings.ToLower(box)]
	if !ok {
		return "", errors.Errorf("cannot find systematik for box %s", box)
	}
	return
}

func (mm *MediathekMapper) GetData(signature string) (barcode string, docID string, box string, err error) {
	sqlstr := "SELECT mms, barcode, box FROM kistebook WHERE signatur=?"
	if err := mm.db.QueryRow(sqlstr, signature).Scan(&docID, &barcode, &box); err != nil {
		return "", "", "", errors.Wrapf(err, "cannot get barcode and docid - %s [%s]", sqlstr, signature)
	}
	return
}

func (mm *MediathekMapper) GetImage(signature string) (imgData []byte, mime string, err error) {
	/*
		 $sig = $_REQUEST['search_key'];

		 $sql = "SELECT box FROM kistebook WHERE signatur=".$db->qstr( $sig );
		 $box = $db->GetOne( $sql );

		readfile( $config['3dthumbdir'].'/'.str_replace( '_', '', $box ).'.png');
	*/
	_, _, box, err := mm.GetData(signature)
	if err != nil {
		//return nil, "", err
		box = "info"
	}
	box = strings.ReplaceAll(box, "_", "")
	boxPath := fmt.Sprintf("%s.jpg", box)
	boxfile, err := mm.boxImageFS.Open(boxPath)
	if err != nil {
		boxfile, err = mm.boxImageFS.Open("info.jpg")
		if err != nil {
			return nil, "", errors.Wrapf(err, "cannot open %s", boxPath)
		}
	}
	boxData, err := io.ReadAll(boxfile)
	if err != nil {
		return nil, "", errors.Wrapf(err, "cannot read %s", boxPath)
	}
	return boxData, "image/jpeg", nil
}

func (mm *MediathekMapper) GetBarcode(signature, docID, barcode string) (imgData []byte, mime string, err error) {
	if signature == "" {
		return nil, "", errors.New("signature needed in call to GetBarcode()")
	}
	if docID == "" {
		docID, barcode, _, err = mm.GetData(signature)
		if err != nil {
			return nil, "", errors.Wrap(err, "cannot get barcode and docid")
		}
	}

	var urlStr = strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(
				mm.siteViewerLink,
				"{DOCID}", url.QueryEscape(docID)),
			"{SIGNATURE}", url.QueryEscape(signature)),
		"{BARCODE}", url.QueryEscape(barcode))
	var png []byte
	// mm.logger.Infof("QRCode: %s", urlStr)
	png, err = qrcode.Encode(urlStr, qrcode.Medium, 130)
	if err != nil {
		return nil, "", errors.Wrapf(err, "cannot generate qrcode for %s", urlStr)
	}
	return png, "image/png", nil
}

func (mm *MediathekMapper) SetData(signature, docID, barcode, projectID string) error {
	// nothing to do
	return nil
}
