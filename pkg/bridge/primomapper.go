package bridge

type PrimoMapper interface {
	GetImage(signature string) (imgData []byte, mime string, err error)
	GetBarcode(signature string) (imgData []byte, mime string, err error)
	SetData(signature, docID, barcode string) error
}
