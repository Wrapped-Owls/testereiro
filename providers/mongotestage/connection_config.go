package mongotestage

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	defaultMongoHost             = "localhost"
	defaultMongoPort             = 27017
	defaultConnectTimeout        = 10 * time.Second
	defaultDisconnectTimeout     = 10 * time.Second
	defaultConnectionPingTimeout = 5 * time.Second
)

// MongoCredential aliases mongo driver credentials for connection auth.
type MongoCredential = options.Credential

// ConnectionConfig defines MongoDB connection and timeout settings for the provider.
type ConnectionConfig struct {
	MongoCredential
	Context           context.Context
	URI               string
	Host              string
	Port              int
	ConnectTimeout    time.Duration
	PingTimeout       time.Duration
	DisconnectTimeout time.Duration
}

func (cfg ConnectionConfig) normalize() (ConnectionConfig, error) {
	normalized := cfg

	if normalized.URI == "" {
		if normalized.Host == "" {
			normalized.Host = defaultMongoHost
		}
		if normalized.Port == 0 {
			normalized.Port = defaultMongoPort
		}
	}
	if normalized.ConnectTimeout <= 0 {
		normalized.ConnectTimeout = defaultConnectTimeout
	}
	if normalized.PingTimeout <= 0 {
		normalized.PingTimeout = defaultConnectionPingTimeout
	}
	if normalized.DisconnectTimeout <= 0 {
		normalized.DisconnectTimeout = defaultDisconnectTimeout
	}
	if normalized.Context == nil {
		normalized.Context = context.Background()
	}

	return normalized, nil
}
