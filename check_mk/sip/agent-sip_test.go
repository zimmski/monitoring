package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSIPRegistryStatus(t *testing.T) {
	rs, err := parseSIPRegistryStatus(`Host                                    dnsmgr Username       Refresh State                Reg.Time
somedomain:5060                    N      098765         105 Registered           Tue, 13 Dec 2016 18:23:39
somedomain:5060                    N      0123445             105 Registered           Tue, 13 Dec 2016 18:23:39
2 SIP registrations.
`)
	assert.NoError(t, err)
	assert.Equal(
		t,
		[]string{
			"somedomain:5060->0123445",
			"somedomain:5060->098765",
		},
		rs,
	)
}
