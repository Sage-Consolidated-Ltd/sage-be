package validator

import (
	"errors"
	"sage-backend/internal/shield/providers/aws"
)

type AWSValidator struct{}
type WebhookValidator struct{}
type KafkaValidator struct{}

type ProviderValidator interface {
	Validate(config []byte) error
}

func (v AWSValidator) Validate(config []byte) error {
	parsed, err := aws.ParseConfig(config)
	if err != nil {
		return err
	}

	return aws.Validate(parsed)
}

func (v WebhookValidator) Validate(config map[string]interface{}) error {
	if config["endpoint"] == nil {
		return errors.New("missing endpoint")
	}
	return nil
}
func (v KafkaValidator) Validate(config map[string]interface{}) error {
	if config["brokers"] == nil {
		return errors.New("missing brokers")
	}
	return nil
}

var Validators = map[string]ProviderValidator{
	"aws":     AWSValidator{},
	// "webhook": WebhookValidator{},
	// "kafka":   KafkaValidator{},
}

