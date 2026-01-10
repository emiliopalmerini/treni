package converters

import (
	"github.com/emiliopalmerini/treni/internal/observation"
	"github.com/emiliopalmerini/treni/internal/voyage"
	"github.com/emiliopalmerini/treni/internal/web/views"
)

// GlobalStats converts observation.GlobalStats to view model.
func GlobalStats(s *observation.GlobalStats) views.GlobalStatsView {
	if s == nil {
		return views.GlobalStatsView{}
	}
	return views.GlobalStatsView{
		TotalObservations: s.TotalObservations,
		AverageDelay:      s.AverageDelay,
		OnTimePercentage:  s.OnTimePercentage,
		CancelledCount:    s.CancelledCount,
	}
}

// CategoryStats converts a slice of observation.CategoryStats to view models.
func CategoryStats(stats []*observation.CategoryStats) []views.CategoryStatsView {
	result := make([]views.CategoryStatsView, len(stats))
	for i, s := range stats {
		result[i] = views.CategoryStatsView{
			Category:         s.Category,
			ObservationCount: s.ObservationCount,
			AverageDelay:     s.AverageDelay,
			OnTimePercentage: s.OnTimePercentage,
		}
	}
	return result
}

// TrainStats converts a slice of observation.TrainStats to view models.
func TrainStats(stats []*observation.TrainStats) []views.TrainStatsView {
	result := make([]views.TrainStatsView, len(stats))
	for i, s := range stats {
		result[i] = views.TrainStatsView{
			TrainNumber:      s.TrainNumber,
			Category:         s.Category,
			OriginName:       s.OriginName,
			DestinationName:  s.DestinationName,
			ObservationCount: s.ObservationCount,
			AverageDelay:     s.AverageDelay,
			MaxDelay:         s.MaxDelay,
			OnTimePercentage: s.OnTimePercentage,
		}
	}
	return result
}

// StationStats converts a slice of observation.StationStats to view models.
func StationStats(stats []*observation.StationStats) []views.StationStatsView {
	result := make([]views.StationStatsView, len(stats))
	for i, s := range stats {
		result[i] = StationStat(s)
	}
	return result
}

// StationStat converts a single observation.StationStats to view model.
func StationStat(s *observation.StationStats) views.StationStatsView {
	if s == nil {
		return views.StationStatsView{}
	}
	return views.StationStatsView{
		StationID:        s.StationID,
		StationName:      s.StationName,
		ObservationCount: s.ObservationCount,
		AverageDelay:     s.AverageDelay,
		OnTimePercentage: s.OnTimePercentage,
	}
}

// Observations converts a slice of observation.TrainObservation to view models.
func Observations(obs []*observation.TrainObservation) []views.ObservationView {
	result := make([]views.ObservationView, len(obs))
	for i, o := range obs {
		result[i] = views.ObservationView{
			ObservedAt:      o.ObservedAt,
			StationName:     o.StationName,
			ObservationType: string(o.ObservationType),
			TrainNumber:     o.TrainNumber,
			TrainCategory:   o.TrainCategory,
			OriginName:      o.OriginName,
			DestinationName: o.DestinationName,
			Delay:           o.Delay,
			IsCancelled:     o.CirculationState == 1,
		}
	}
	return result
}

// VoyageDetail converts voyage.VoyageWithStops to view model.
func VoyageDetail(v *voyage.VoyageWithStops) views.VoyageDetailView {
	stops := make([]views.VoyageStopView, len(v.Stops))
	for i, stop := range v.Stops {
		stops[i] = views.VoyageStopView{
			StationID:          stop.StationID,
			StationName:        stop.StationName,
			StopSequence:       stop.StopSequence,
			StopType:           stop.StopType,
			ScheduledArrival:   stop.ScheduledArrival,
			ScheduledDeparture: stop.ScheduledDeparture,
			ActualArrival:      stop.ActualArrival,
			ActualDeparture:    stop.ActualDeparture,
			ArrivalDelay:       stop.ArrivalDelay,
			DepartureDelay:     stop.DepartureDelay,
			Platform:           stop.Platform,
			IsSuppressed:       stop.IsSuppressed,
			LastObservationAt:  stop.LastObservationAt,
		}
	}

	return views.VoyageDetailView{
		VoyageID:        v.ID.String(),
		TrainNumber:     v.TrainNumber,
		TrainCategory:   v.TrainCategory,
		OriginName:      v.OriginName,
		DestinationName: v.DestinationName,
		ScheduledDate:   v.ScheduledDate,
		Stops:           stops,
	}
}

// VoyageList converts a slice of voyage.Voyage to view models.
func VoyageList(voyages []*voyage.Voyage) []views.VoyageListView {
	result := make([]views.VoyageListView, len(voyages))
	for i, v := range voyages {
		result[i] = views.VoyageListView{
			VoyageID:        v.ID.String(),
			TrainNumber:     v.TrainNumber,
			TrainCategory:   v.TrainCategory,
			OriginName:      v.OriginName,
			DestinationName: v.DestinationName,
			ScheduledDate:   v.ScheduledDate,
			UpdatedAt:       v.UpdatedAt,
		}
	}
	return result
}
