package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type metricsCollector struct {
	sg StatisticsGetter

	textUpdatesRecieved *prometheus.Desc
	chatsTotal          *prometheus.Desc
	usersTotal          *prometheus.Desc
	gamesTotal          *prometheus.Desc
	startsTotal         *prometheus.Desc
}

func newMetricsCollector(sg StatisticsGetter) *metricsCollector {
	return &metricsCollector{
		sg: sg,
		textUpdatesRecieved: prometheus.NewDesc("text_updates_total",
			"Shows how many text updates has been recieved",
			nil, nil,
		),
		chatsTotal: prometheus.NewDesc("chats_total",
			"Shows how many chats are in the bot",
			nil, nil,
		),
		usersTotal: prometheus.NewDesc("users_total",
			"Shows how many users are in the bot",
			nil, nil,
		),
		gamesTotal: prometheus.NewDesc("games_total",
			"Shows how many games has been played",
			nil, nil,
		),
		startsTotal: prometheus.NewDesc("starts_total",
			"Shows how many times start command has been called",
			nil, nil,
		),
	}
}

// Writes all descriptors to the prometheus desc channel
func (c *metricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.textUpdatesRecieved
	ch <- c.chatsTotal
	ch <- c.usersTotal
	ch <- c.gamesTotal
	ch <- c.startsTotal
}

// Collect implements required collect function for all promehteus collectors
func (c *metricsCollector) Collect(ch chan<- prometheus.Metric) {
	stats, _ := c.sg.GetStatistics()

	// TODO Get rid of textUpdatesRecieved global variable
	ch <- prometheus.MustNewConstMetric(c.textUpdatesRecieved, prometheus.CounterValue, textUpdatesRecieved)
	ch <- prometheus.MustNewConstMetric(c.chatsTotal, prometheus.CounterValue, float64(stats.Chats))
	ch <- prometheus.MustNewConstMetric(c.usersTotal, prometheus.CounterValue, float64(stats.Users))
	ch <- prometheus.MustNewConstMetric(c.gamesTotal, prometheus.CounterValue, float64(stats.GamesPlayed))
	ch <- prometheus.MustNewConstMetric(c.startsTotal, prometheus.CounterValue, startsTotal)
}
