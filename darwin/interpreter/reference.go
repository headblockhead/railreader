package interpreter

import (
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func (u UnitOfWork) InterpretReference(reference unmarshaller.Reference) error {
	u.log.Debug("interpreting a Reference")
	return u.referenceRepository.Insert(reference)
}
