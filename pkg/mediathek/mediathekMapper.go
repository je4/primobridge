package mediathek

import (
	"database/sql"
	"fmt"
	"github.com/bluele/gcache"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"github.com/skip2/go-qrcode"
	"io"
	"io/fs"
	"net/url"
	"strings"
)

type MediathekMapper struct {
	cache          gcache.Cache
	db             *sql.DB
	logger         *logging.Logger
	boxImageFS     fs.FS
	siteViewerLink string
}

func NewMediathekMapper(db *sql.DB,
	boxImageFS fs.FS,
	siteViewerLink string,
	logger *logging.Logger) (*MediathekMapper, error) {
	pm := &MediathekMapper{
		db:             db,
		boxImageFS:     boxImageFS,
		siteViewerLink: siteViewerLink,
		logger:         logger,
		cache:          gcache.New(500).LRU().Build(),
	}
	return pm, nil
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
