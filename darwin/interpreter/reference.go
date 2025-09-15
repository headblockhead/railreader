package interpreter

import (
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func (u UnitOfWork) interpretReference(ref unmarshaller.Reference) error {
	u.log.Debug("interpreting a Reference")
	return nil
}
