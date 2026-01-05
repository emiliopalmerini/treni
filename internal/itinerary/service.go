package itinerary

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/emiliopalmerini/treni/internal/viaggiatreno"
)

const (
	minConnectionTime    = 10 * time.Minute
	maxResults           = 10
	maxDeparturesToCheck = 20
)

type Service struct {
	vt viaggiatreno.Client
}

func NewService(vt viaggiatreno.Client) *Service {
	return &Service{vt: vt}
}

func (s *Service) FindSolutions(ctx context.Context, fromID, toID string) ([]Solution, error) {
	departures, err := s.vt.Partenze(ctx, fromID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("fetching departures: %w", err)
	}

	if len(departures) > maxDeparturesToCheck {
		departures = departures[:maxDeparturesToCheck]
	}

	var (
		solutions []Solution
		mu        sync.Mutex
		wg        sync.WaitGroup
	)

	for _, dep := range departures {
		if dep.IsCancelled() {
			continue
		}

		wg.Add(1)
		go func(dep viaggiatreno.Departure) {
			defer wg.Done()

			found := s.findSolutionsForDeparture(ctx, dep, fromID, toID)
			if len(found) > 0 {
				mu.Lock()
				solutions = append(solutions, found...)
				mu.Unlock()
			}
		}(dep)
	}

	wg.Wait()

	sort.Slice(solutions, func(i, j int) bool {
		return solutions[i].ArrivalAt.Before(solutions[j].ArrivalAt)
	})

	if len(solutions) > maxResults {
		solutions = solutions[:maxResults]
	}

	return solutions, nil
}

func (s *Service) findSolutionsForDeparture(ctx context.Context, dep viaggiatreno.Departure, fromID, toID string) []Solution {
	trainNum := strconv.Itoa(dep.TrainNumber)

	status, err := s.vt.AndamentoTreno(ctx, dep.OriginID, trainNum, dep.DepartureTime)
	if err != nil || status == nil {
		return nil
	}

	var solutions []Solution

	fromIdx, toIdx := -1, -1
	for i, stop := range status.Stops {
		if stop.StationID == fromID {
			fromIdx = i
		}
		if stop.StationID == toID {
			toIdx = i
		}
	}

	if fromIdx >= 0 && toIdx > fromIdx {
		sol := s.buildDirectSolution(status, fromIdx, toIdx)
		solutions = append(solutions, *sol)
	}

	if toIdx < 0 && fromIdx >= 0 {
		connections := s.findConnections(ctx, status, fromIdx, toID)
		solutions = append(solutions, connections...)
	}

	return solutions
}

func (s *Service) buildDirectSolution(status *viaggiatreno.TrainStatus, fromIdx, toIdx int) *Solution {
	fromStop := status.Stops[fromIdx]
	toStop := status.Stops[toIdx]

	leg := Leg{
		TrainNumber: strconv.Itoa(status.TrainNumber),
		TrainType:   status.Category,
		From: Station{
			ID:   fromStop.StationID,
			Name: fromStop.StationName,
		},
		To: Station{
			ID:   toStop.StationID,
			Name: toStop.StationName,
		},
		DepartureAt: time.UnixMilli(fromStop.ScheduledDeparture),
		ArrivalAt:   time.UnixMilli(toStop.ScheduledArrival),
		Platform:    fromStop.EffectivePlatform(),
		Delay:       status.Delay,
	}

	sol := NewSolution()
	sol.AddLeg(leg)
	return sol
}

func (s *Service) findConnections(ctx context.Context, firstTrain *viaggiatreno.TrainStatus, fromIdx int, toID string) []Solution {
	var solutions []Solution
	fromStop := firstTrain.Stops[fromIdx]

	for i := fromIdx + 1; i < len(firstTrain.Stops); i++ {
		interStop := firstTrain.Stops[i]
		if interStop.ActualStopType == 3 {
			continue
		}

		arrivalTime := time.UnixMilli(interStop.ScheduledArrival)
		minDeparture := arrivalTime.Add(minConnectionTime)

		departures, err := s.vt.Partenze(ctx, interStop.StationID, minDeparture)
		if err != nil || len(departures) == 0 {
			continue
		}

		for _, connDep := range departures {
			if connDep.IsCancelled() {
				continue
			}

			connDepTime := connDep.DepartureTimeUTC()
			if connDepTime.Before(minDeparture) {
				continue
			}

			connTrainNum := strconv.Itoa(connDep.TrainNumber)
			connStatus, err := s.vt.AndamentoTreno(ctx, interStop.StationID, connTrainNum, connDep.DepartureTime)
			if err != nil || connStatus == nil {
				continue
			}

			connFromIdx, connToIdx := -1, -1
			for j, stop := range connStatus.Stops {
				if stop.StationID == interStop.StationID {
					connFromIdx = j
				}
				if stop.StationID == toID {
					connToIdx = j
				}
			}

			if connFromIdx >= 0 && connToIdx > connFromIdx {
				sol := s.buildConnectionSolution(firstTrain, fromStop, interStop, connStatus, connFromIdx, connToIdx)
				solutions = append(solutions, *sol)
				break
			}
		}

		if len(solutions) >= 3 {
			break
		}
	}

	return solutions
}

func (s *Service) buildConnectionSolution(
	firstTrain *viaggiatreno.TrainStatus,
	fromStop, interStop viaggiatreno.Stop,
	secondTrain *viaggiatreno.TrainStatus,
	connFromIdx, connToIdx int,
) *Solution {
	connFromStop := secondTrain.Stops[connFromIdx]
	connToStop := secondTrain.Stops[connToIdx]

	leg1 := Leg{
		TrainNumber: strconv.Itoa(firstTrain.TrainNumber),
		TrainType:   firstTrain.Category,
		From: Station{
			ID:   fromStop.StationID,
			Name: fromStop.StationName,
		},
		To: Station{
			ID:   interStop.StationID,
			Name: interStop.StationName,
		},
		DepartureAt: time.UnixMilli(fromStop.ScheduledDeparture),
		ArrivalAt:   time.UnixMilli(interStop.ScheduledArrival),
		Platform:    fromStop.EffectivePlatform(),
		Delay:       firstTrain.Delay,
	}

	leg2 := Leg{
		TrainNumber: strconv.Itoa(secondTrain.TrainNumber),
		TrainType:   secondTrain.Category,
		From: Station{
			ID:   connFromStop.StationID,
			Name: connFromStop.StationName,
		},
		To: Station{
			ID:   connToStop.StationID,
			Name: connToStop.StationName,
		},
		DepartureAt: time.UnixMilli(connFromStop.ScheduledDeparture),
		ArrivalAt:   time.UnixMilli(connToStop.ScheduledArrival),
		Platform:    connFromStop.EffectivePlatform(),
		Delay:       secondTrain.Delay,
	}

	sol := NewSolution()
	sol.AddLeg(leg1)
	sol.AddLeg(leg2)
	return sol
}
