package winshim

import (
	"strings"

	"golang.org/x/net/dns/dnsmessage"
)

// Answer replies to one DNS query: every name under tld is loopback, where nginx
// listens through mirrored networking, and anything else is refused. Answering
// DNS rather than writing hosts lines gives wildcards for free.
func Answer(query []byte, tld string) ([]byte, error) {
	var p dnsmessage.Parser
	h, err := p.Start(query)
	if err != nil {
		return nil, err
	}
	q, err := p.Question()
	if err != nil {
		return nil, err
	}
	resp := dnsmessage.Header{ID: h.ID, Response: true, Authoritative: true, RecursionDesired: h.RecursionDesired}
	owned := underTLD(q.Name.String(), tld)
	if !owned {
		resp.RCode = dnsmessage.RCodeRefused
	}
	b := dnsmessage.NewBuilder(nil, resp)
	b.EnableCompression()
	if err := b.StartQuestions(); err != nil {
		return nil, err
	}
	if err := b.Question(q); err != nil {
		return nil, err
	}
	if err := b.StartAnswers(); err != nil {
		return nil, err
	}
	rh := dnsmessage.ResourceHeader{Name: q.Name, Class: dnsmessage.ClassINET, TTL: 60}
	switch {
	case owned && q.Type == dnsmessage.TypeA:
		err = b.AResource(rh, dnsmessage.AResource{A: [4]byte{127, 0, 0, 1}})
	case owned && q.Type == dnsmessage.TypeAAAA:
		err = b.AAAAResource(rh, dnsmessage.AAAAResource{AAAA: [16]byte{15: 1}})
	}
	if err != nil {
		return nil, err
	}
	return b.Finish()
}

func underTLD(name, tld string) bool {
	name = strings.ToLower(strings.TrimSuffix(name, "."))
	tld = strings.ToLower(tld)
	return tld != "" && (name == tld || strings.HasSuffix(name, "."+tld))
}
