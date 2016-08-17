package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseServicesStatus(t *testing.T) {
	assert.Equal(
		t,
		map[string]string{
			"acpid.service":       "dead",
			"after-local.service": "exited",
			"apache2.service":     "running",
			"ido2db.service":      "failed",
		},
		parseServicesStatus(`
			acpid.service                        not-found inactive dead    acpid.service
after-local.service                  loaded    active   exited  /etc/init.d/after.local Compatibility
	apache2.service                      loaded    active   running The Apache Webserver
				ido2db.service                       loaded    failed   failed  Icinga Data Out Utilities (IDOUtils)
`),
	)
}
