package output

import (
	"time"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/collection"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
)

// BenchmarkFixtures holds representative MQ-shaped datasets for rendering benchmarks.
var BenchmarkFixtures = struct {
	QueuePage    collection.Page[mqadmin.QueueSummary]
	MessagePage  collection.Page[messaging.MessageRecord]
	ProfilePage  collection.Page[application.ProfileSummary]
	ChannelPage  collection.Page[mqadmin.ChannelSummary]
	QueueDetail  mqadmin.QueueDetail
	Reason       mqadmin.ReasonExplanation
	Connectivity mqadmin.ConnectivityReport
	QMStatus     mqadmin.QueueManagerStatus
}{
	QueuePage: collection.Page[mqadmin.QueueSummary]{
		Items: []mqadmin.QueueSummary{
			{Name: "DEV.QUEUE.1", Type: "local"},
			{Name: "DEV.QUEUE.2", Type: "local"},
			{Name: "SYSTEM.ADMIN.COMMAND.QUEUE", Type: "alias"},
			{Name: "SYSTEM.DEFAULT.MODEL.QUEUE", Type: "model"},
			{Name: "APP.IN", Type: "local"},
			{Name: "APP.OUT", Type: "local"},
			{Name: "APP.ERROR", Type: "local"},
			{Name: "AUDIT.LOG", Type: "local"},
			{Name: "INT.REQ", Type: "remote"},
			{Name: "INT.REPLY", Type: "remote"},
		},
		Limit:            50,
		Truncated:        true,
		TruncationReason: collection.TruncationLimitReached,
		NextCursor:       "INT.REPLY",
	},
	MessagePage: collection.Page[messaging.MessageRecord]{
		Items: []messaging.MessageRecord{
			{
				MessageID: "ID:41424344", CorrelationID: "CID:1", Format: "MQSTR",
				MessageLength: 128, PutDate: "2026-08-05", PutTime: "12:00:00",
				Priority: 5, Persistence: "persistent", Encoding: messaging.EncodingUTF8,
			},
			{
				MessageID: "ID:45464748", Format: "MQSTR", MessageLength: 64,
				PutDate: "2026-08-05", PutTime: "12:00:01", Priority: 4,
				Persistence: "nonPersistent", Encoding: messaging.EncodingOmitted,
			},
			{
				MessageID: "ID:494a4b4c", CorrelationID: "CID:2", Format: "MQHRF2",
				MessageLength: 4096, PutDate: "2026-08-05", PutTime: "12:00:02",
				Priority: 9, Persistence: "persistent", Encoding: messaging.EncodingBase64,
				PayloadTruncated: true,
			},
		},
		Limit: 10,
	},
	ProfilePage: collection.Page[application.ProfileSummary]{
		Items: []application.ProfileSummary{
			{
				Name: "prod", QueueManager: "QM1", Endpoint: "https://mq.prod.example:9443",
				Capabilities: []string{"inspect", "browse", "produce", "consume"}, Valid: true,
			},
			{
				Name: "dr", QueueManager: "QM2", Endpoint: "https://mq.dr.example:9443",
				Capabilities: []string{"inspect"}, Valid: true,
			},
			{
				Name: "lab", QueueManager: "QM_LAB", Endpoint: "https://mq.lab.example:9443",
				Capabilities: []string{"inspect", "browse"}, Valid: false,
			},
		},
		Limit: 50,
	},
	ChannelPage: collection.Page[mqadmin.ChannelSummary]{
		Items: []mqadmin.ChannelSummary{
			{Name: "DEV.APP.SVRCONN", Type: "SVRCONN"},
			{Name: "TO.REMOTE.QM", Type: "SDR"},
			{Name: "FROM.REMOTE.QM", Type: "RCVR"},
			{Name: "CLNT.CONN", Type: "CLNTCONN"},
		},
		Limit: 50,
	},
	QueueDetail: mqadmin.QueueDetail{
		Name:            "DEV.QUEUE.1",
		Type:            "local",
		MaxDepth:        5000,
		CurrentDepth:    42,
		OpenInputCount:  1,
		OpenOutputCount: 2,
		InhibitGet:      "allowed",
		InhibitPut:      "allowed",
	},
	Reason: mqadmin.ExplainReasonCode(2035),
	Connectivity: mqadmin.ConnectivityReport{
		Profile:       "prod",
		Endpoint:      "https://mq.prod.example:9443",
		Reachable:     true,
		IdentityMatch: true,
		LatencyMs:     18,
		Identity:      mqadmin.Identity{Configured: "QM1", Observed: "QM1"},
		CheckedAt:     time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC),
	},
	QMStatus: mqadmin.QueueManagerStatus{
		Profile:      "prod",
		Identity:     mqadmin.Identity{Configured: "QM1", Observed: "QM1"},
		Availability: mqadmin.Available,
		Running:      true,
		StatusText:   "Running",
		LastChecked:  time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC),
	},
}
