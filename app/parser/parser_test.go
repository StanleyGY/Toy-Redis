package parser

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseInteger(t *testing.T) {
	t.Log("Test parsing positive number")
	mockReader := bytes.NewReader([]byte{0x2b, 0x30, 0x31, 0x32, 0xd, 0xa})
	actualInt, err := parseInteger(mockReader)
	assert.NoError(t, err)
	assert.Equal(t, 12, actualInt)

	t.Log("Test parsing negative number")
	mockReader = bytes.NewReader([]byte{0x2d, 0x30, 0x31, 0x32, 0xd, 0xa})
	actualInt, err = parseInteger(mockReader)
	assert.NoError(t, err)
	assert.Equal(t, -12, actualInt)

	t.Log("Test parsing invalid number")
	mockReader = bytes.NewReader([]byte{0x2d, 0x30, 0x31, 0x32, 0x2d, 0xd, 0xa})
	_, err = parseInteger(mockReader)
	assert.Error(t, err)
}
