package cucumber

import "testing"

func TestParseRequestFromBytes(t *testing.T) {
	body := `name: yuppie
nodes_num: 3
dns_name: rock
`
	r, err := ParseRequestFromBytes([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	if r.Name != "yuppie" {
		t.Fatalf("expected name to be yubbie got %s", r.Name)
	}

	if r.DNSName == "" {
		t.Fatalf("expected dns name to be populated, but it is empty")
	}
}

func TestParseRequestFromBytesWithDnsNameEmpty(t *testing.T) {
	body := `name: yuppie
nodes_num: 3
`
	r, err := ParseRequestFromBytes([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	if r.DNSName != "" {
		t.Fatalf("expected dns name to be empty we got %s", r.DNSName)
	}
}
