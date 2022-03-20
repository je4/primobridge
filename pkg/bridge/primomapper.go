package bridge

type PrimoMapper interface {
	GetImage(signature string) (imgData []byte, mime string, err error)
	GetBarcode(signature, docID, barcode string) (imgData []byte, mime string, err error)
	SetData(signature, docID, barcode, projectID string) error
	GetData(signature string) (barcode string, docID string, box string, err error)
}
