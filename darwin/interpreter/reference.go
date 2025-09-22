package interpreter

import (
	"log/slog"

	"github.com/headblockhead/railreader/darwin/repository"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
)

func (u UnitOfWork) InterpretReference(log *slog.Logger, referenceRepository repository.Reference, reference unmarshaller.Reference) error {
	log.Debug("interpreting a Reference")
	var rrs repository.ReferenceRow
	rrs.ID = reference.ID
	return referenceRepository.Insert(rrs)
}
