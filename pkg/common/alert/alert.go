package alert

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/metrics"
	"github.com/slack-go/slack"
	"github.com/spf13/viper"
)

// MetricAlerts is an array of LogMetric types with an easier lookup method
type MetricAlerts []MetricAlert

var once = sync.Once{}

var metricAlerts = MetricAlerts{}
var slackChannelCache = make(map[string]slack.Channel)

// GetMetricAlerts will return the log metrics.
func GetMetricAlerts() MetricAlerts {
	once.Do(func() {
		viper.Set("metricAlerts", metricAlerts)
	})

	tmp := viper.Get("metricAlerts")
	ma, ok := tmp.(MetricAlerts)
	if !ok {
		log.Println("Error casting metricAlerts from Viper")
	}

	return ma
}

// AddAlert adds an alert to an existing MetricAlerts object
func (mas MetricAlerts) AddAlert(alert MetricAlert) MetricAlerts {
	mas = append(mas, alert)
	viper.Set("metricAlerts", mas)
	return mas
}

// MetricAlert lets you define a test name and the criteria to alert
// an owner via an alert channel of some sort.
type MetricAlert struct {
	// --- Description of Test ---
	// Name of the metric to look for
	Name string

	// -- Description of Test Owner ---
	// TeamOwner describes which RedHat team may own this test
	TeamOwner string
	// PrimaryContact is a point person or SME for this set of tests.
	// If there isn't one, it should default to the person committing these tests.
	PrimaryContact string

	// --- Description of Alert Channels ---
	// SlackChannel is the channel in slack to message with an alert
	SlackChannel string
	// Email is the email address to send alerts to.
	// TODO: Make this work.
	// This does not work yet.
	Email string

	// --- Description of Alert Triggers ---
	// FailureThreshold is the number of failures in a rolling window
	FailureThreshold int
}

// Notify prepares and then iterates through MetricAlerts to generate notifications
func (mas MetricAlerts) Notify() error {
	client, err := metrics.NewClient()
	if err != nil {
		return fmt.Errorf("unable to create Prometheus client: %v", err)
	}

	for _, ma := range mas {
		log.Printf("Checking %s", ma.Name)
		if err := ma.Check(client); err != nil {
			return err
		}
	}

	return nil
}

// Check will query and notify depending on query results
func (ma MetricAlert) Check(client *metrics.Client) error {
	results, err := client.ListFailedJUnitResultsByTestName(ma.QuerySafeName(), time.Now().Add(-24*time.Hour), time.Now())
	if err != nil {
		return err
	}

	if len(results) >= ma.FailureThreshold {
		log.Printf("Alert triggered for %s: %d >= %d", ma.Name, len(results), ma.FailureThreshold)
		sendSlackMessage(ma.SlackChannel, fmt.Sprintf("%s has seen %d failures in the last 24h", ma.Name, len(results)))
	}

	return nil
}

// QuerySafeName is a helper function that returns a regex prometheus safe query string
func (ma MetricAlert) QuerySafeName() string {
	tmp := strings.Replace(ma.Name, "[", "\\\\[", -1)
	tmp = strings.Replace(tmp, "]", "\\\\]", -1)
	tmp = strings.Replace(tmp, "(", "\\\\(", -1)
	tmp = strings.Replace(tmp, ")", "\\\\)", -1)
	tmp = strings.Replace(tmp, "-", "\\\\-", -1)
	tmp = strings.Replace(tmp, ".", "\\\\.", -1)
	tmp = strings.Replace(tmp, ":", "\\\\:", -1)
	return tmp

}

// RegisterGinkgoAlert will retrieve the ginkgo test info and register an alert given
// the supplied arguments
func RegisterGinkgoAlert(test, team, contact, slack, email string, threshold int) {
	ma := GetMetricAlerts()
	testAlert := MetricAlert{
		Name:             test,
		TeamOwner:        team,
		PrimaryContact:   contact,
		SlackChannel:     slack,
		Email:            email,
		FailureThreshold: 4,
	}
	ma.AddAlert(testAlert)
}

func sendSlackMessage(channel, message string) error {
	slackAPI := slack.New(viper.GetString(config.Alert.SlackAPIToken))
	var slackChannel slack.Channel
	var ok bool

	if slackChannel, ok = slackChannelCache[channel]; !ok {
		channels, _, err := slackAPI.GetConversations(&slack.GetConversationsParameters{})
		if err != nil {
			return err
		}
		for _, c := range channels {
			slackChannelCache[c.Name] = c
			if c.Name == channel {
				slackChannel = c
			}
		}
	}

	if slackChannel.ID == "" {
		return fmt.Errorf("no slack channel named `%s` found", channel)
	}

	_, _, err := slackAPI.PostMessage(slackChannel.ID, slack.MsgOptionText(message, false))
	return err
}
