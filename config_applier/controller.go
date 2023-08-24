package config_applier

import (
	"context"
	"github.com/go-logr/logr"
	"math"
	"time"
)

const (
	sleepDuration = 20 * time.Second
	maxRetries    = 10 // 1024s or ~17 minutes at peak
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
	pollingTimer := backoffTicker{
		Ticker:        time.NewTicker(sleepDuration),
		sleepDuration: sleepDuration,
		maxRetries:    maxRetries,
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
			controller.logger.Error(err, "Configuration applier iteration failed, it will be retried on next iteration period")
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
