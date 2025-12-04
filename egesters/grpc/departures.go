package grpc

import (
	"github.com/headblockhead/railreader/egesters/grpc/proto"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc"
)

func (e *Egester) GetDepartures(in *proto.DepartureRequest, sc grpc.ServerStreamingServer[proto.DepartureResponse]) error {
	rows, err := e.dbpool.Query(e.ctx, `
		SELECT schedule_id FROM darwin.schedules WHERE location_id = @location_id
	`, pgx.StrictNamedArgs{
		"location_id": in.StationCode,
	})
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var scheduleID string
		if err := rows.Scan(&scheduleID); err != nil {
			return err
		}
		sc.Send(&proto.DepartureResponse{
			ScheduleId: scheduleID,
		})
	}
	return nil
}
