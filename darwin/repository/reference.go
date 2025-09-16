package repository

import (
	"context"
	"log/slog"

	"github.com/headblockhead/railreader/darwin/unmarshaller"
	"github.com/jackc/pgx/v5"
)

type Reference interface {
	Insert(reference unmarshaller.Reference) error
}

type PGXReference struct {
	ctx context.Context
	log *slog.Logger
	tx  pgx.Tx
}

func NewPGXReference(ctx context.Context, log *slog.Logger, tx pgx.Tx) PGXReference {
	return PGXReference{
		ctx: ctx,
		log: log,
		tx:  tx,
	}
}

func (rr PGXReference) Insert(reference unmarshaller.Reference) error {
	rr.log.Debug("inserting Reference")
	for _, loc := range reference.Locations {
		if _, err := rr.tx.Exec(rr.ctx, `
			INSERT INTO locations
				VALUES (
					@location_id
					,@computerised_reservation_system_id
					,@train_operating_company_id
					,@name
				) 
				ON CONFLICT (location_id) DO
				NOTHING;
		`, pgx.StrictNamedArgs{
			"location_id":                        string(loc.Location),
			"computerised_reservation_system_id": loc.CRS,
			"train_operating_company_id":         loc.TOC,
			"name":                               loc.Name,
		}); err != nil {
			return err
		}
	}
	for _, toc := range reference.TrainOperatingCompanies {
		if _, err := rr.tx.Exec(rr.ctx, `
			INSERT INTO train_operating_companies
				VALUES (
					@train_operating_company_id
					,@name
					,@url
				) 
				ON CONFLICT (train_operating_company_id) DO
				NOTHING;
		`, pgx.StrictNamedArgs{
			"train_operating_company_id": toc.ID,
			"name":                       toc.Name,
			"url":                        toc.URL,
		}); err != nil {
			return err
		}
	}
	for _, reason := range reference.LateReasons {
		if _, err := rr.tx.Exec(rr.ctx, `
			INSERT INTO late_reasons
				VALUES (
					@late_reason_id
					,@description
				) 
				ON CONFLICT (late_reason_id) DO
				NOTHING;
		`, pgx.StrictNamedArgs{
			"late_reason_id": reason.ReasonID,
			"description":    reason.Description,
		}); err != nil {
			return err
		}
	}
	for _, reason := range reference.CancellationReasons {
		if _, err := rr.tx.Exec(rr.ctx, `
			INSERT INTO cancellation_reasons
				VALUES (
					@cancellation_reason_id
					,@description
				) 
				ON CONFLICT (cancellation_reason_id) DO
				NOTHING;
		`, pgx.StrictNamedArgs{
			"cancellation_reason_id": reason.ReasonID,
			"description":            reason.Description,
		}); err != nil {
			return err
		}
	}
	for _, cis := range reference.CustomerInformationSystemSources {
		if _, err := rr.tx.Exec(rr.ctx, `
			INSERT INTO customer_information_system_sources
				VALUES (
					@customer_information_system_source_id
					,@name
				) 
				ON CONFLICT (customer_information_system_source_id) DO
				NOTHING;
		`, pgx.StrictNamedArgs{
			"customer_information_system_source_id": cis.CIS,
			"name":                                  cis.Name,
		}); err != nil {
			return err
		}
	}
	return nil
}
