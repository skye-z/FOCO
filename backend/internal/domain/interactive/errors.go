package interactive

import "errors"

var (
	ErrVersionNotFound       = errors.New("interactive version not found")
	ErrVersionReadOnly       = errors.New("interactive version is read-only")
	ErrVersionNotPublishable = errors.New("interactive version is not publishable")
	ErrUnitNotFound          = errors.New("interactive unit not found")
)
