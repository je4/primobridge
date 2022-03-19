package mediathek

import (
	"database/sql"
	"github.com/bluele/gcache"
	"github.com/op/go-logging"
)

type PrimoMapper struct {
	cache  gcache.Cache
	db     sql.DB
	logger *logging.Logger
}

func NewPrimoMapper(db sql.DB, logger *logging.Logger) (*PrimoMapper, error) {
	pm := &PrimoMapper{
		db:     db,
		logger: logger,
		cache:  gcache.New(500).LRU().Build(),
	}
	return pm, nil
}

func (pm *PrimoMapper) GetImage(signature string) (imgData []byte, mime string, err error) {
	/*
		 $sig = $_REQUEST['search_key'];

		 $sql = "SELECT box FROM kistebook WHERE signatur=".$db->qstr( $sig );
		 $box = $db->GetOne( $sql );

		readfile( $config['3dthumbdir'].'/'.str_replace( '_', '', $box ).'.png');
	*/

}

func (pm *PrimoMapper) GetBarcode(signature string) (imgData []byte, mime string, err error) {

}

func (pm *PrimoMapper) SetData(signature, docID, barcode string) error {
	// nothing to do
	return nil
}
