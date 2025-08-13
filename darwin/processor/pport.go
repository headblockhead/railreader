package processor

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/headblockhead/railreader/darwin/decoder"
)

func (p *Processor) processPushPortMessage(log *slog.Logger, msg *decoder.PushPortMessage) error {
	log.Debug("processing PushPortMessage")
	if msg.NewTimetableFiles != nil {
		// TODO
		return errors.New("unimplemented: NewTimetableFiles")
	}
	if msg.StatusUpdate != nil {
		// TODO
		return errors.New("unimplemented: StatusUpdate")
	}
	if msg.UpdateResponse != nil {
		if err := p.processResponse(log, msg.UpdateResponse); err != nil {
			return fmt.Errorf("failed to process UpdateResponse: %w", err)
		}
		return nil
	}
	if msg.SnapshotResponse != nil {
		if err := p.processResponse(log, msg.SnapshotResponse); err != nil {
			return fmt.Errorf("failed to process SnapshotResponse: %w", err)
		}
		return nil
	}
	return errors.New("PushPortMessage does not contain any data")
}

func (p *Processor) processResponse(log *slog.Logger, resp *decoder.Response) error {
	log.Debug("processing Response", slog.String("updateOrigin", resp.UpdateOrigin), slog.String("requestSourceSystem", resp.RequestSourceSystem))
	for _, schedule := range resp.Schedules {
		if err := p.processSchedule(log, &schedule); err != nil {
			return fmt.Errorf("failed to process schedule: %w", err)
		}
	}
	/* for _, deactivation := range resp.Deactivations {*/
	/*processDeactivation(log, &deactivation)*/
	/*}*/
	/*for _, association := range resp.Associations {*/
	/*processAssociation(log, &association)*/
	/*}*/
	/*for _, formation := range resp.Formations {*/
	/*processFormation(log, &formation)*/
	/*}*/
	/*for _, forecastTime := range resp.ForecastTimes {*/
	/*processForecastTime(log, &forecastTime)*/
	/*}*/
	/*for _, serviceLoading := range resp.ServiceLoadings {*/
	/*processServiceLoading(log, &serviceLoading)*/
	/*}*/
	/*for _, formationLoading := range resp.FormationLoadings {*/
	/*processFormationLoading(log, &formationLoading)*/
	/*}*/
	/*for _, stationMessage := range resp.StationMessages {*/
	/*processStationMessage(log, &stationMessage)*/
	/*}*/
	/*for _, trainAlert := range resp.TrainAlerts {*/
	/*processTrainAlert(log, &trainAlert)*/
	/*}*/
	/*for _, trainOrder := range resp.TrainOrders {*/
	/*processTrainOrder(log, &trainOrder)*/
	/*}*/
	/*for _, headcodeChange := range resp.HeadcodeChanges {*/
	/*processHeadcodeChange(log, &headcodeChange)*/
	/*}*/
	/*for _, alarm := range resp.Alarms {*/
	/*processAlarm(log, &alarm)*/
	/*}*/
	return nil
}
