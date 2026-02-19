package dns

import (
	"log"
	"photobooth/internal/config"

	"github.com/miekg/dns"
)

type Server struct {
	server *dns.Server
	config config.WifiConfig
}

func NewServer(cfg config.WifiConfig) *Server {
	return &Server{
		config: cfg,
	}
}

func (s *Server) Start() {
	// Only start if WiFi and Captive Portal are enabled
	// But for now, we assume if this is called, it should run.
	// Ideally check config.

	dns.HandleFunc(".", s.handleDNSRequest)

	s.server = &dns.Server{Addr: ":53", Net: "udp"}

	go func() {
		log.Printf("🌐 DNS Server (Captive Portal) starting on :53 -> %s", s.config.IpAddress)
		if err := s.server.ListenAndServe(); err != nil {
			log.Printf("⚠️ DNS Server failed: %v (Port 53 might be in use or permissions missing)", err)
		}
	}()
}

func (s *Server) Stop() {
	if s.server != nil {
		s.server.Shutdown()
	}
}

func (s *Server) handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		for _, q := range m.Question {
			switch q.Qtype {
			case dns.TypeA:
				// Always return our IP for any A record query
				rr, err := dns.NewRR(q.Name + " 3600 IN A " + s.config.IpAddress)
				if err == nil {
					m.Answer = append(m.Answer, rr)
				}
			}
		}
	}

	w.WriteMsg(m)
}
