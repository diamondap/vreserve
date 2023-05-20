package core_test

import (
	"fmt"
	"testing"

	"github.com/diamondap/vreserve/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVolumeClient(t *testing.T) {
	client := core.NewVolumeClient(serviceUrl)
	require.NotNil(t, client)
	expectedUrl := fmt.Sprintf("http://127.0.0.1:%d", port)
	assert.Equal(t, expectedUrl, client.BaseURL())
}

func TestVolumeReserve(t *testing.T) {
	runService(t)
	client := core.NewVolumeClient(serviceUrl)
	require.NotNil(t, client)

	ok, err := client.Reserve("/tmp/some_file", uint64(800))
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = client.Reserve("", uint64(800))
	assert.NotNil(t, err) // path required
	assert.False(t, ok)

	ok, err = client.Reserve("", uint64(0))
	assert.NotNil(t, err) // > 0 bytes required
	assert.False(t, ok)
}

func TestVolumeRelease(t *testing.T) {
	runService(t)
	client := core.NewVolumeClient(serviceUrl)
	require.NotNil(t, client)

	err := client.Release("/tmp/some_file")
	assert.Nil(t, err)

	err = client.Release("")
	assert.NotNil(t, err) // path required
}

func TestVolumeReport(t *testing.T) {
	runService(t)
	client := core.NewVolumeClient(serviceUrl)
	require.NotNil(t, client)

	data, err := client.Report("/tmp/some_file")
	assert.Nil(t, err)
	assert.NotNil(t, data)

	data, err = client.Report("")
	assert.NotNil(t, err) // path required
	assert.Nil(t, data)
}
