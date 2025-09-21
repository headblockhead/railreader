package interpreter

import (
	"log/slog"

	"github.com/headblockhead/railreader/darwin/repository"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func (u UnitOfWork) InterpretReference(log *slog.Logger, referenceRepository repository.Reference, reference unmarshaller.Reference) error {
	log.Debug("interpreting a Reference")
	var rrs repository.ReferenceRowsSet
	// TODO: interpret the reference into rows
	return referenceRepository.Insert(rrs)
}
