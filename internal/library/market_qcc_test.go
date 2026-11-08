package library

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

func TestQCCMarketOAuthDiscoveryAndInstall(t *testing.T) {
	sandbox(t)
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"servers":[]}`))
	}))
	defer registry.Close()
	was := registryURL
	registryURL = registry.URL
	t.Cleanup(func() { registryURL = was })
	for _, query := range []string{"", "企查查", "QCC", "qcc-company"} {
		list, err := MarketServers(query)
		if err != nil {
			t.Fatal(err)
		}
		i := slices.IndexFunc(list, func(m MarketServer) bool { return m.ID == "qcc-company" })
		if i < 0 {
			t.Fatalf("Discover query %q did not offer QCC", query)
		}
		m := list[i]
		if !m.SignIn || !m.Featured || len(m.Inputs) != 0 || m.Homepage != "https://agent.qcc.com" {
			t.Fatalf("QCC must offer browser sign-in without a key form: %+v", m)
		}
	}
	if _, err := InstallServer("qcc-company", nil, []string{}); err != nil {
		t.Fatal(err)
	}
	l, err := load()
	if err != nil {
		t.Fatal(err)
	}
	s := l.server("qcc-company")
	if s == nil || s.Transport != "http" || s.URL != "https://agent.qcc.com/mcp/company/stream" || len(s.Headers) != 0 || len(s.Agents) != 0 {
		t.Fatalf("unexpected QCC install: %+v", s)
	}
	if _, err := InstallServer("qcc-company", nil, []string{}); err == nil {
		t.Fatal("installing QCC twice should not overwrite the existing entry")
	}
}
