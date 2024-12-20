package satellite

import (
	"context"

	"container-registry.com/harbor-satellite/internal/config"
	"container-registry.com/harbor-satellite/internal/notifier"
	"container-registry.com/harbor-satellite/internal/scheduler"
	"container-registry.com/harbor-satellite/internal/state"
	"container-registry.com/harbor-satellite/internal/utils"
	"container-registry.com/harbor-satellite/logger"
)

type Satellite struct {
	stateReader  state.StateReader
	schedulerKey scheduler.SchedulerKey
}

func NewSatellite(ctx context.Context, schedulerKey scheduler.SchedulerKey) *Satellite {
	return &Satellite{
		schedulerKey: schedulerKey,
	}
}

func (s *Satellite) Run(ctx context.Context) error {
	log := logger.FromContext(ctx)
	log.Info().Msg("Starting Satellite")
	var cronExpr string
	state_fetch_period := config.GetStateFetchPeriod()
	cronExpr, err := utils.FormatDuration(state_fetch_period)
	if err != nil {
		log.Warn().Msgf("Error formatting duration in seconds: %v", err)
		log.Warn().Msgf("Using default duration: %v", state.DefaultFetchAndReplicateStateTimePeriod)
		cronExpr = state.DefaultFetchAndReplicateStateTimePeriod
	}
	userName := config.GetHarborUsername()
	password := config.GetHarborPassword()
	zotURL := config.GetZotURL()
	sourceRegistry := utils.FormatRegistryURL(config.GetRemoteRegistryURL())
	useUnsecure := config.UseUnsecure()
	// Get the scheduler from the context
	scheduler := ctx.Value(s.schedulerKey).(scheduler.Scheduler)
	// Create a simple notifier and add it to the process
	notifier := notifier.NewSimpleNotifier(ctx)
	// Creating a process to fetch and replicate the state
	states := config.GetStates()
	fetchAndReplicateStateProcess := state.NewFetchAndReplicateStateProcess(scheduler.NextID(), cronExpr, notifier, userName, password, zotURL, sourceRegistry, useUnsecure, states)
	// Add the process to the scheduler
	scheduler.Schedule(fetchAndReplicateStateProcess)

	return nil
}
