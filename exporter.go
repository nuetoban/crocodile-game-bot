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
	startTotal          *prometheus.Desc
	ratingTotal         *prometheus.Desc
	globalRatingTotal   *prometheus.Desc
	cstatTotal          *prometheus.Desc
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
		startTotal: prometheus.NewDesc("start_total",
			"Shows how many times /start command has been called",
			nil, nil,
		),
		ratingTotal: prometheus.NewDesc("rating_total",
			"Shows how many times /rating command has been called",
			nil, nil,
		),
		globalRatingTotal: prometheus.NewDesc("globalrating_total",
			"Shows how many times /globalrating command has been called",
			nil, nil,
		),
		cstatTotal: prometheus.NewDesc("cstat_total",
			"Shows how many times /cstat command has been called",
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
	ch <- c.startTotal
	ch <- c.ratingTotal
	ch <- c.globalRatingTotal
	ch <- c.cstatTotal
}

// Collect implements required collect function for all promehteus collectors
func (c *metricsCollector) Collect(ch chan<- prometheus.Metric) {
	stats, _ := c.sg.GetStatistics()

	// TODO Get rid of textUpdatesRecieved global variable
	ch <- prometheus.MustNewConstMetric(c.textUpdatesRecieved, prometheus.CounterValue, textUpdatesRecieved)
	ch <- prometheus.MustNewConstMetric(c.chatsTotal, prometheus.CounterValue, float64(stats.Chats))
	ch <- prometheus.MustNewConstMetric(c.usersTotal, prometheus.CounterValue, float64(stats.Users))
	ch <- prometheus.MustNewConstMetric(c.gamesTotal, prometheus.CounterValue, float64(stats.GamesPlayed))
	ch <- prometheus.MustNewConstMetric(c.startTotal, prometheus.CounterValue, startTotal)
	ch <- prometheus.MustNewConstMetric(c.ratingTotal, prometheus.CounterValue, ratingTotal)
	ch <- prometheus.MustNewConstMetric(c.globalRatingTotal, prometheus.CounterValue, globalRatingTotal)
	ch <- prometheus.MustNewConstMetric(c.cstatTotal, prometheus.CounterValue, cstatTotal)
}
