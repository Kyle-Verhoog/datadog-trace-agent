package main

// https://stackoverflow.com/questions/21856517/whats-the-practical-limit-on-the-size-of-single-packet-transmitted-over-domain

import (
	"net"

	log "github.com/cihub/seelog"

	"github.com/DataDog/datadog-trace-agent/config"
	"github.com/DataDog/datadog-trace-agent/info"
	"github.com/DataDog/datadog-trace-agent/model"
	"github.com/DataDog/datadog-trace-agent/sampler"
)

// TODO: This should be in it's own module/file
type UDSServer struct {
	listener   net.Listener
	handleFunc func(c net.Conn) // function to handle connections
	connBuf    chan net.Conn
	quitBuf    chan int
}

func NewUDSServer(ln net.Listener) *UDSServer {
	return &UDSServer{
		handleFunc: func(c net.Conn) {
		},
		listener: ln,
		connBuf:  make(chan net.Conn),
		quitBuf:  make(chan int),
	}
}

func (s *UDSServer) Accepter() {
	for {
		fd, err := s.listener.Accept()
		if err != nil {
			log.Errorf("Failed to accept incoming socket connection")
		}

		s.connBuf <- fd
	}
}

func (s *UDSServer) Serve(ln net.Listener) {
	s.listener = ln
	go s.Accepter()

L:
	for {
		select {
		case conn := <-s.connBuf:
			s.handleFunc(conn)
		case <-s.quitBuf:
			break L
		}
	}
}

func (s *UDSServer) Shutdown() {
	s.listener.Close()
	s.quitBuf <- 0
}

type UDSReceiver struct {
	traces   chan model.Trace
	services chan model.ServicesMetadata
	conf     *config.AgentConfig
	dynConf  *config.DynamicConfig
	server   *UDSServer

	stats      *info.ReceiverStats
	preSampler *sampler.PreSampler

	maxRequestBodyLength int64
	debug                bool
}

// NewUDSReceiver returns a pointer to a new UDSReceiver
func NewUDSReceiver(
	conf *config.AgentConfig, dynConf *config.DynamicConfig, traces chan model.Trace, services chan model.ServicesMetadata,
) *UDSReceiver {
	// use buffered channels so that handlers are not waiting on downstream processing
	return &UDSReceiver{
		conf:       conf,
		dynConf:    dynConf,
		stats:      info.NewReceiverStats(),
		preSampler: sampler.NewPreSampler(conf.PreSampleRate),

		traces:   traces,
		services: services,

		maxRequestBodyLength: maxRequestBodyLength,
	}
}

func (r *UDSReceiver) Run() {
	if err := r.Listen(); err != nil {
		panic(err)
	}
}

func (r *UDSReceiver) Stop() {
	r.server.Shutdown()
}

func (r *UDSReceiver) Listen() error {
	// TODO: if the agent crashes it will leave the socket file on the machine
	// this will break subsequent calls unless we unlink the file
	ln, err := net.Listen("unix", r.conf.UDSReceiverFile)
	if err != nil {
		ln.Close()
		return err
	}
	r.server = NewUDSServer(ln)

	go r.server.Serve(ln)
	return nil
}
