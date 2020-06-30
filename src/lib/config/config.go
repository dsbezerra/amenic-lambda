package config

import (
	"log"
	"os"

	"github.com/dsbezerra/amenic-lambda/src/lib/env"
)

// DBType ...
type DBType string

const (
	// MongoDB name
	MongoDB DBType = "MongoDB"

	// DefaultDBConnection is the default URI used to connect to the database
	DefaultDBConnection = "mongodb://localhost/amenic"

	// DefaultDBLoggingConnection is the default URI used to connect to the logging database
	DefaultDBLoggingConnection = "mongodb://localhost/amenic-logs"

	DefaultMessageBrokerType = "amqp"
	DefaultAMQPMessageBroker = "amqp://guest:guest@localhost:5672"
)

// ServiceConfig ...
type ServiceConfig struct {
	DBType       DBType `json:"database_type"`
	DBConnection string `json:"database_connection"`
	RESTEndpoint string `json:"rest_endpoint"`
	// RESTTLSEndpoint    string
	IsProduction           bool   `json:"is_production"`
	MessageBrokerType      string `json:"message_broker_type"`
	AMQPMessageBroker      string `json:"amqp_message_broker"`
	ImageServiceConnection string `json:"imageservice_connection"`
	AWSLambda              bool   `json:"aws_lambda"`
}

// LoadConfiguration initializes the required configuration
// for this service
func LoadConfiguration() (*ServiceConfig, error) {
	// Start config with default values
	config := &ServiceConfig{
		DBType:            MongoDB,
		DBConnection:      DefaultDBConnection,
		RESTEndpoint:      "0.0.0.0:8000", // Default REST endpoint.
		IsProduction:      false,
		MessageBrokerType: DefaultMessageBrokerType,
		AMQPMessageBroker: DefaultAMQPMessageBroker,
	}

	config.AWSLambda = os.Getenv("AWS_LAMBDA") == "true"

	if !config.AWSLambda {
		_, err := env.LoadEnv()
		if err != nil {
			return config, err
		}
	}

	release := os.Getenv("AMENIC_MODE") == "release"

	var dbEnv string
	if !release {
		dbEnv = "DB_DEV"
	} else {
		dbEnv = "DB_PROD"
	}
	connection := os.Getenv(dbEnv)
	if connection == "" {
		log.Fatal("Main database URI is missing")
	}

	// config.AMQPMessageBroker = vars["AMQP_URL"]
	// config.RESTEndpoint = vars["LISTEN_URL"]
	// NOTE: If we use other database type, update this.
	config.DBConnection = connection
	config.IsProduction = release
	config.ImageServiceConnection = os.Getenv("CLOUDINARY_URL")
	return config, nil
}
