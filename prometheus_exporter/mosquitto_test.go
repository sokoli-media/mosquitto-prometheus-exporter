package prometheus_exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"testing"
)

var sLogForTesting = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func getGaugeValue(t *testing.T, metric prometheus.Gauge) float64 {
	var m = &dto.Metric{}
	if err := metric.Write(m); err != nil {
		t.Fatalf("couldn't get metric value: %s", err)
	}
	return m.Gauge.GetValue()
}

func Test_Uptime(t *testing.T) {
	err := processMosquittoMessage(sLogForTesting, "$SYS/broker/uptime", "9284 seconds")
	require.NoError(t, err)

	assert.Equal(t, float64(9284), getGaugeValue(t, mosquittoUptime))
}

func Test_ClientsDisconnected(t *testing.T) {
	err := processMosquittoMessage(sLogForTesting, "$SYS/broker/clients/disconnected", "2")
	require.NoError(t, err)

	assert.Equal(t, float64(2), getGaugeValue(t, mosquittoDisconnectedClients))
}

func Test_ClientsConnected(t *testing.T) {
	err := processMosquittoMessage(sLogForTesting, "$SYS/broker/clients/connected", "2")
	require.NoError(t, err)

	assert.Equal(t, float64(2), getGaugeValue(t, mosquittoConnectedClients))
}

func Test_ClientsTotal(t *testing.T) {
	err := processMosquittoMessage(sLogForTesting, "$SYS/broker/clients/total", "5")
	require.NoError(t, err)

	assert.Equal(t, float64(5), getGaugeValue(t, mosquittoTotalClients))
}

func Test_ClientsExpired(t *testing.T) {
	err := processMosquittoMessage(sLogForTesting, "$SYS/broker/clients/expired", "5")
	require.NoError(t, err)

	assert.Equal(t, float64(5), getGaugeValue(t, mosquittoExpiredClients))
}

func Test_StoredMessages(t *testing.T) {
	err := processMosquittoMessage(sLogForTesting, "$SYS/broker/store/messages/count", "5")
	require.NoError(t, err)

	assert.Equal(t, float64(5), getGaugeValue(t, mosquittoStoredMessages))
}

func Test_StoredBytes(t *testing.T) {
	err := processMosquittoMessage(sLogForTesting, "$SYS/broker/store/messages/bytes", "5")
	require.NoError(t, err)

	assert.Equal(t, float64(5), getGaugeValue(t, mosquittoStoredBytes))
}

func Test_ActiveSubscriptions(t *testing.T) {
	err := processMosquittoMessage(sLogForTesting, "$SYS/broker/subscriptions/count", "10")
	require.NoError(t, err)

	assert.Equal(t, float64(10), getGaugeValue(t, mosquittoActiveSubscriptions))
}

func Test_RetainedMessages(t *testing.T) {
	err := processMosquittoMessage(sLogForTesting, "$SYS/broker/retained messages/count", "10")
	require.NoError(t, err)

	assert.Equal(t, float64(10), getGaugeValue(t, mosquittoRetainedMessages))
}

func Test_MessagesReceivedSinceBrokerStarted(t *testing.T) {
	valueBefore := getGaugeValue(t, mosquittoReceivedMessages)

	err := processMosquittoMessage(sLogForTesting, "$SYS/broker/messages/received", "123")
	require.NoError(t, err)

	valueAfter := getGaugeValue(t, mosquittoReceivedMessages)

	assert.Equal(t, float64(123), valueAfter-valueBefore)
}

func Test_MessagesSentSinceBrokerStarted(t *testing.T) {
	valueBefore := getGaugeValue(t, mosquittoSentMessages)

	err := processMosquittoMessage(sLogForTesting, "$SYS/broker/messages/sent", "123")
	require.NoError(t, err)

	valueAfter := getGaugeValue(t, mosquittoSentMessages)

	assert.Equal(t, float64(123), valueAfter-valueBefore)
}

func Test_BytesReceivedSinceBrokerStarted(t *testing.T) {
	valueBefore := getGaugeValue(t, mosquittoReceivedBytes)

	err := processMosquittoMessage(sLogForTesting, "$SYS/broker/bytes/received", "234")
	require.NoError(t, err)

	valueAfter := getGaugeValue(t, mosquittoReceivedBytes)

	assert.Equal(t, float64(234), valueAfter-valueBefore)
}

func Test_BytesSentSinceBrokerStarted(t *testing.T) {
	valueBefore := getGaugeValue(t, mosquittoSentBytes)

	err := processMosquittoMessage(sLogForTesting, "$SYS/broker/bytes/sent", "234")
	require.NoError(t, err)

	valueAfter := getGaugeValue(t, mosquittoSentBytes)

	assert.Equal(t, float64(234), valueAfter-valueBefore)
}
