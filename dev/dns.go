package dev

import (
	"net"
	"time"

	"github.com/miekg/dns"
	"gopkg.in/tomb.v2"
)

type DNSResponder struct {
	Address string
	Domains []string

	udpServer *dns.Server
	tcpServer *dns.Server
}

func NewDNSResponder(address string, domains []string) *DNSResponder {
	udp := &dns.Server{Addr: address, Net: "udp", TsigSecret: nil}
	tcp := &dns.Server{Addr: address, Net: "tcp", TsigSecret: nil}

	d := &DNSResponder{Address: address, Domains: domains, udpServer: udp, tcpServer: tcp}

	return d
}

func (d *DNSResponder) handleDNS(w dns.ResponseWriter, r *dns.Msg) {
	var (
		v4 bool
		rr dns.RR
		a  net.IP
	)

	dom := r.Question[0].Name

	m := new(dns.Msg)
	m.SetReply(r)
	if ip, ok := w.RemoteAddr().(*net.UDPAddr); ok {
		a = ip.IP
		v4 = a.To4() != nil
	}
	if ip, ok := w.RemoteAddr().(*net.TCPAddr); ok {
		a = ip.IP
		v4 = a.To4() != nil
	}

	if v4 {
		rr = new(dns.A)
		rr.(*dns.A).Hdr = dns.RR_Header{Name: dom, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0}
		rr.(*dns.A).A = a.To4()
	} else {
		rr = new(dns.AAAA)
		rr.(*dns.AAAA).Hdr = dns.RR_Header{Name: dom, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 0}
		rr.(*dns.AAAA).AAAA = a
	}

	switch r.Question[0].Qtype {
	case dns.TypeAAAA, dns.TypeA:
		m.Answer = append(m.Answer, rr)
	}

	if r.IsTsig() != nil {
		if w.TsigStatus() == nil {
			m.SetTsig(r.Extra[len(r.Extra)-1].(*dns.TSIG).Hdr.Name, dns.HmacMD5, 300, time.Now().Unix())
		}
	}

	w.WriteMsg(m)
}

func (d *DNSResponder) Serve() error {
	for _, domain := range d.Domains {
		dns.HandleFunc(domain+".", d.handleDNS)
	}

	var t tomb.Tomb

	t.Go(func() error {
		return d.udpServer.ListenAndServe()
	})

	t.Go(func() error {
		return d.tcpServer.ListenAndServe()
	})

	return t.Wait()
}
