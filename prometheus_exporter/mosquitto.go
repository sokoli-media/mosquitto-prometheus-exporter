package prometheus_exporter

import (
	"errors"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"log/slog"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

var mosquittoUptime = promauto.NewGauge(prometheus.GaugeOpts{Name: "mosquitto_uptime"})
var mosquittoDisconnectedClients = promauto.NewGauge(prometheus.GaugeOpts{Name: "mosquitto_disconnected_clients"})
var mosquittoConnectedClients = promauto.NewGauge(prometheus.GaugeOpts{Name: "mosquitto_connected_clients"})
var mosquittoTotalClients = promauto.NewGauge(prometheus.GaugeOpts{Name: "mosquitto_total_clients"})
var mosquittoExpiredClients = promauto.NewGauge(prometheus.GaugeOpts{Name: "mosquitto_expired_clients"})

var mosquittoActiveSubscriptions = promauto.NewGauge(prometheus.GaugeOpts{Name: "mosquitto_active_subscriptions"})

var mosquittoRetainedMessages = promauto.NewGauge(prometheus.GaugeOpts{Name: "mosquitto_retained_messages"})
var mosquittoStoredMessages = promauto.NewGauge(prometheus.GaugeOpts{Name: "mosquitto_stored_messages"})
var mosquittoStoredBytes = promauto.NewGauge(prometheus.GaugeOpts{Name: "mosquitto_stored_bytes"})
var mosquittoReceivedMessages = promauto.NewGauge(prometheus.GaugeOpts{Name: "mosquitto_received_messages_total"})
var mosquittoSentMessages = promauto.NewGauge(prometheus.GaugeOpts{Name: "mosquitto_sent_messages_total"})

var mosquittoReceivedBytes = promauto.NewGauge(prometheus.GaugeOpts{Name: "mosquitto_received_bytes_total"})
var mosquittoSentBytes = promauto.NewGauge(prometheus.GaugeOpts{Name: "mosquitto_sent_bytes_total"})

var mosquittoLastUpdate = promauto.NewGauge(prometheus.GaugeOpts{Name: "mosquitto_last_update"})
var mosquittoUnknownTopic = promauto.NewCounterVec(prometheus.CounterOpts{Name: "mosquitto_unknown_topic"}, []string{"topic"})

func isTopicIgnored(topic string) bool {
	ignoredPrefixes := []string{
		"$SYS/broker/publish/",
		"$SYS/broker/load/",
	}
	for _, prefix := range ignoredPrefixes {
		if strings.HasPrefix(topic, prefix) {
			return true
		}
	}

	ignoredTopics := []string{
		"$SYS/broker/version",
		"$SYS/broker/clients/active",
		"$SYS/broker/clients/inactive",
		"$SYS/broker/clients/maximum",
		"$SYS/broker/publish/messages/received",
		"$SYS/broker/publish/messages/sent",
		"$SYS/broker/messages/stored",
		"$SYS/broker/shared_subscriptions/count",
	}
	return slices.Contains(ignoredTopics, topic)
}

func processMosquittoMessage(logger *slog.Logger, topic string, payload string) error {
	logger.Info("received message", "topic", topic, "payload", payload)

	if isTopicIgnored(topic) {
		return nil
	}

	switch topic {
	case "$SYS/broker/uptime":
		if found := regexp.MustCompile("^([0-9]+) seconds$").FindStringSubmatch(payload); len(found) > 0 {
			if s, err := strconv.ParseFloat(found[1], 64); err == nil {
				mosquittoUptime.Set(s)
			} else {
				return err
			}
		} else {
			return errors.New("didn't find expected pattern")
		}
	case "$SYS/broker/clients/disconnected":
		if s, err := strconv.ParseFloat(payload, 64); err == nil {
			mosquittoDisconnectedClients.Set(s)
		} else {
			return err
		}
	case "$SYS/broker/clients/connected":
		if s, err := strconv.ParseFloat(payload, 64); err == nil {
			mosquittoConnectedClients.Set(s)
		} else {
			return err
		}
	case "$SYS/broker/clients/total":
		if s, err := strconv.ParseFloat(payload, 64); err == nil {
			mosquittoTotalClients.Set(s)
		} else {
			return err
		}
	case "$SYS/broker/clients/expired":
		if s, err := strconv.ParseFloat(payload, 64); err == nil {
			mosquittoExpiredClients.Set(s)
		} else {
			return err
		}
	case "$SYS/broker/messages/received":
		if s, err := strconv.ParseFloat(payload, 64); err == nil {
			mosquittoReceivedMessages.Set(s)
		} else {
			return err
		}
	case "$SYS/broker/messages/sent":
		if s, err := strconv.ParseFloat(payload, 64); err == nil {
			mosquittoSentMessages.Set(s)
		} else {
			return err
		}
	case "$SYS/broker/bytes/received":
		if s, err := strconv.ParseFloat(payload, 64); err == nil {
			mosquittoReceivedBytes.Set(s)
		} else {
			return err
		}
	case "$SYS/broker/bytes/sent":
		if s, err := strconv.ParseFloat(payload, 64); err == nil {
			mosquittoSentBytes.Set(s)
		} else {
			return err
		}
	case "$SYS/broker/store/messages/count":
		if s, err := strconv.ParseFloat(payload, 64); err == nil {
			mosquittoStoredMessages.Set(s)
		} else {
			return err
		}
	case "$SYS/broker/store/messages/bytes":
		if s, err := strconv.ParseFloat(payload, 64); err == nil {
			mosquittoStoredBytes.Set(s)
		} else {
			return err
		}
	case "$SYS/broker/subscriptions/count":
		if s, err := strconv.ParseFloat(payload, 64); err == nil {
			mosquittoActiveSubscriptions.Set(s)
		} else {
			return err
		}
	case "$SYS/broker/retained messages/count":
		if s, err := strconv.ParseFloat(payload, 64); err == nil {
			mosquittoRetainedMessages.Set(s)
		} else {
			return err
		}
	default:
		mosquittoUnknownTopic.With(prometheus.Labels{"topic": topic}).Inc()
		return fmt.Errorf("unknown topic: %s", topic)
	}

	mosquittoLastUpdate.SetToCurrentTime()
	return nil
}

func CollectMosquittoMetrics(logger *slog.Logger, config MosquittoConfig, wg *sync.WaitGroup, quit chan bool) {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(config.Broker)
	opts.SetClientID(config.ClientId)
	opts.SetUsername(config.Username)
	opts.SetPassword(config.Password)
	opts.SetCleanSession(false)

	choke := make(chan MQTT.Message)

	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		choke <- msg
	})

	logger.Info("connecting to mqtt")
	client := MQTT.NewClient(opts)

	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		logger.Error("couldn't connect to mqtt", "error", token.Error())
		return
	}

	logger.Info("subscribing to mqtt topics")
	token = client.Subscribe("$SYS/#", byte(2), nil)
	if token.WaitTimeout(5*time.Second) && token.Error() != nil {
		logger.Error("couldn't subscribe to mqtt topics", "error", token.Error())
		return
	}

	logger.Info("waiting for mosquitto updates on mqtt")
	for {
		select {
		case message := <-choke:
			topic := message.Topic()
			payload := string(message.Payload())

			err := processMosquittoMessage(logger, topic, payload)
			if err != nil {
				logger.Error(
					"couldn't process mqtt message",
					"topic",
					message.Topic(),
					"payload",
					string(message.Payload()),
					"error",
					err,
				)
			}
		case <-quit:
			logger.Info("disconnecting from mqtt gracefully")
			client.Disconnect(250)
			wg.Done()
			return
		}
	}
}
