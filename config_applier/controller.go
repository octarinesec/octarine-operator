package config_applier

import (
	"context"
	"github.com/go-logr/logr"
	"math"
	"time"
)

type configurationApplier interface {
	RunIteration(ctx context.Context) error
}

type RemoteConfigurationController struct {
	applier configurationApplier
	logger  logr.Logger
}

func NewRemoteConfigurationController(applier configurationApplier, logger logr.Logger) *RemoteConfigurationController {
	return &RemoteConfigurationController{applier: applier, logger: logger}
}

func (controller *RemoteConfigurationController) RunLoop(signalsContext context.Context) {
	// TODO: Parameters vs consts

	pollingSleepDuration := 20 * time.Second
	pollingTimer := backoffTicker{
		Ticker:        time.NewTicker(pollingSleepDuration),
		sleepDuration: pollingSleepDuration,
		maxRetries:    10, // 1024s or ~17minutes max
	}
	defer pollingTimer.Stop()

	for {
		select {
		case <-signalsContext.Done():
			controller.logger.Info("Received cancel signal, turning off configuration applier")
			return
		case <-pollingTimer.C:
			// Nothing to do; this is the polling sleep case
		}
		err := controller.applier.RunIteration(signalsContext)

		if err != nil {
			controller.logger.Error(err, "Configuration applier iteration failed, will retry again")
			pollingTimer.resetErr()
		} else {
			controller.logger.Info("Completed configuration applier iteration, sleeping")
			pollingTimer.resetSuccess()
		}
	}
}

// backoffTicker is a ticker with exponential backoff for errors and static backoff for success cases
// Note: When calling resetErr or resetSuccess, the ticker will wait the full sleep duration again
type backoffTicker struct {
	*time.Ticker

	sleepDuration time.Duration
	maxRetries    int

	currentRetries int
}

func (b *backoffTicker) resetErr() {
	if b.currentRetries < b.maxRetries {
		b.currentRetries++
	}

	nextSleepDuration := time.Duration(math.Pow(2.0, float64(b.currentRetries)))*time.Second + b.sleepDuration
	b.Reset(nextSleepDuration)
}

func (b *backoffTicker) resetSuccess() {
	b.currentRetries = 0
	b.Reset(b.sleepDuration)
}
