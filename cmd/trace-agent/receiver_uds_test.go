package main

import (
	"testing"

	"github.com/DataDog/datadog-trace-agent/config"
	"github.com/DataDog/datadog-trace-agent/model"
	"github.com/stretchr/testify/assert"
)

func NewTestUDSReceiverConfig() *config.AgentConfig {
	conf := config.New()
	return conf
}

func NewTestUDSReceiverFromConfig(conf *config.AgentConfig) *UDSReceiver {
	dynConf := config.NewDynamicConfig()

	rawTraceChan := make(chan model.Trace, 5000)
	serviceChan := make(chan model.ServicesMetadata, 50)
	receiver := NewUDSReceiver(conf, dynConf, rawTraceChan, serviceChan)

	return receiver
}

func TestUDSReceiverRun(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(uint64(1), uint64(1))
	conf := NewTestUDSReceiverConfig()
	receiver := NewTestUDSReceiverFromConfig(conf)

	receiver.Run()
	receiver.Stop()
}

func TestUDSReceiverClient(t *testing.T) {
	conf := NewTestUDSReceiverConfig()
	receiver := NewTestUDSReceiverFromConfig(conf)

	receiver.Run()
	receiver.Stop()
}
