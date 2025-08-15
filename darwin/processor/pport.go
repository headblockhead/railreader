package processor

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/headblockhead/railreader/darwin/decoder"
)

func (p *Processor) processPushPortMessage(log *slog.Logger, msg *decoder.PushPortMessage) error {
	if msg == nil {
		return errors.New("PushPortMessage is nil")
	}

	timestamp, err := time.Parse(time.RFC3339Nano, msg.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp %q: %w", msg.Timestamp, err)
	}
	log.Debug("processing PushPortMessage", slog.Time("timestamp", timestamp), slog.String("version", msg.Version))

	/* if msg.NewTimetableFiles != nil {*/
	/*if err := p.processNewTimetableFiles(log, msg, msg.NewTimetableFiles); err != nil {*/
	/*return fmt.Errorf("failed to process StatusUpdate: %w", err)*/
	/*}*/
	/*return nil*/
	/*}*/
	/*if msg.StatusUpdate != nil {*/
	/*if err := p.processStatusUpdate(log, msg, msg.StatusUpdate); err != nil {*/
	/*return fmt.Errorf("failed to process StatusUpdate: %w", err)*/
	/*}*/
	/*return nil*/
	/*}*/
	if msg.UpdateResponse != nil {
		if err := p.processResponse(log, timestamp, false, msg.UpdateResponse); err != nil {
			return fmt.Errorf("failed to process UpdateResponse: %w", err)
		}
		return nil
	}
	/* if msg.SnapshotResponse != nil {*/
	/*if err := p.processResponse(log, timestamp, true, msg.SnapshotResponse); err != nil {*/
	/*return fmt.Errorf("failed to process SnapshotResponse: %w", err)*/
	/*}*/
	/*return nil*/
	/*}*/
	return errors.New("PushPortMessage does not contain any data")
}

func (p *Processor) processResponse(log *slog.Logger, lastUpdated time.Time, snapshot bool, resp *decoder.Response) error {
	if resp == nil {
		return errors.New("Response is nil")
	}

	log.Debug("processing Response", slog.String("updateOrigin", resp.Source), slog.String("requestSourceSystem", resp.SourceSystem), slog.Bool("snapshot", snapshot))

	for _, schedule := range resp.Schedules {
		if err := p.processSchedule(log, lastUpdated, resp.Source, resp.SourceSystem, &schedule); err != nil {
			return fmt.Errorf("failed to process Schedule %s: %w", schedule.RID, err)
		}
	}
	/* for _, deactivation := range resp.Deactivations {*/
	/*if err := p.processDeactivation(log, msg, resp, &deactivation); err != nil {*/
	/*return fmt.Errorf("failed to process Deactivation: %w", err)*/
	/*}*/
	/*}*/
	/*for _, association := range resp.Associations {*/
	/*if err := p.processAssociation(log, msg, resp, &association); err != nil {*/
	/*return fmt.Errorf("failed to process Association: %w", err)*/
	/*}*/
	/*}*/
	/*for _, formation := range resp.Formations {*/
	/*if err := p.processFormation(log, msg, resp, &formation); err != nil {*/
	/*return fmt.Errorf("failed to process Formation: %w", err)*/
	/*}*/
	/*}*/
	/*for _, forecastTime := range resp.ForecastTimes {*/
	/*if err := p.processForecastTime(log, msg, resp, &forecastTime); err != nil {*/
	/*return fmt.Errorf("failed to process ForecastTime: %w", err)*/
	/*}*/
	/*}*/
	/*for _, serviceLoading := range resp.ServiceLoadings {*/
	/*if err := p.processServiceLoading(log, msg, resp, &serviceLoading); err != nil {*/
	/*return fmt.Errorf("failed to process ServiceLoading: %w", err)*/
	/*}*/
	/*}*/
	/*for _, formationLoading := range resp.FormationLoadings {*/
	/*if err := p.processFormationLoading(log, msg, resp, &formationLoading); err != nil {*/
	/*return fmt.Errorf("failed to process FormationLoading: %w", err)*/
	/*}*/
	/*}*/
	/*for _, stationMessage := range resp.StationMessages {*/
	/*if err := p.processStationMessage(log, msg, resp, &stationMessage); err != nil {*/
	/*return fmt.Errorf("failed to process StationMessage: %w", err)*/
	/*}*/
	/*}*/
	/*for _, trainAlert := range resp.TrainAlerts {*/
	/*if err := p.processTrainAlert(log, msg, resp, &trainAlert); err != nil {*/
	/*return fmt.Errorf("failed to process TrainAlert: %w", err)*/
	/*}*/
	/*}*/
	/*for _, trainOrder := range resp.TrainOrders {*/
	/*if err := p.processTrainOrder(log, msg, resp, &trainOrder); err != nil {*/
	/*return fmt.Errorf("failed to process TrainOrder: %w", err)*/
	/*}*/
	/*}*/
	/*for _, headcodeChange := range resp.HeadcodeChanges {*/
	/*if err := p.processHeadcodeChange(log, msg, resp, &headcodeChange); err != nil {*/
	/*return fmt.Errorf("failed to process HeadcodeChange: %w", err)*/
	/*}*/
	/*}*/
	/*for _, alarm := range resp.Alarms {*/
	/*if err := p.processAlarm(log, msg, resp, &alarm); err != nil {*/
	/*return fmt.Errorf("failed to process Alarm: %w", err)*/
	/*}*/
	/*}*/
	return nil
}
