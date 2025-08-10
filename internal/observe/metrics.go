package observe

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	onlineUsers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "chat_online_users",
		Help: "Number of online users",
	})

	messagesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chat_messages_total",
			Help: "Total chat messages by type",
		},
		[]string{"type"}, // local|remote
	)

	directMessagesTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chat_direct_messages_total",
		Help: "Total direct (private) messages",
	})

	droppedMessagesTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chat_dropped_messages_total",
		Help: "Total messages dropped due to client backpressure",
	})

	heartbeatsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chat_heartbeats_total",
		Help: "Total heartbeats received",
	})

	commandsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chat_commands_total",
			Help: "Total commands executed by name",
		},
		[]string{"name"},
	)

	commandErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chat_command_errors_total",
			Help: "Total command errors by reason",
		},
		[]string{"reason"}, // not_found|permission|handler|parse
	)
)

func init() {
	prometheus.MustRegister(
		onlineUsers,
		messagesTotal,
		directMessagesTotal,
		droppedMessagesTotal,
		heartbeatsTotal,
		commandsTotal,
		commandErrorsTotal,
	)
}

func IncMessage(kind string)        { messagesTotal.WithLabelValues(kind).Inc() }
func IncDirect()                    { directMessagesTotal.Inc() }
func IncDropped()                   { droppedMessagesTotal.Inc() }
func IncHeartbeat()                 { heartbeatsTotal.Inc() }
func AddOnline(delta float64)       { onlineUsers.Add(delta) }
func IncCommand(name string)        { commandsTotal.WithLabelValues(name).Inc() }
func IncCommandError(reason string) { commandErrorsTotal.WithLabelValues(reason).Inc() }
