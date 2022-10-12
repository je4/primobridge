package bridge

type Link struct {
	Type, HRef, Label string
}
type Class struct {
	DE, EN string
	Links  map[string][]Link
}
type PrimoMapper interface {
	GetImage(signature string) (imgData []byte, mime string, err error)
	GetBarcode(signature, docID, barcode string) (imgData []byte, mime string, err error)
	SetData(signature, docID, barcode, projectID string) error
	GetData(signature string) (barcode string, docID string, box string, err error)
	GetSystematik(box string) (string, error)
	GetSystematikHierarchy(sys string) (map[string]map[string]Class, error)
}
