package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAllowedRecipients(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, os.Remove(f.Name()))
	}()

	_, err = f.WriteString(`identifier,number
phone_identifier1,0123456111
phone_identifier2,0123456222
`)
	assert.NoError(t, err)
	assert.NoError(t, f.Close())

	rs, err := parseAllowedRecipients(f.Name())
	assert.NoError(t, err)
	assert.Equal(
		t,
		map[string]string{
			"phone_identifier1": "0123456111",
			"phone_identifier2": "0123456222",
		},
		rs,
	)
}

func TestConvertTextToSpeech(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, os.Remove(f.Name()))
	}()
	assert.NoError(t, f.Close())

	assert.NoError(t, convertTextToSpeech("hey ho", f.Name()))
	in, err := ioutil.ReadFile(f.Name())
	assert.NoError(t, err)

	assert.Contains(t, string(in), "WAV")
}
