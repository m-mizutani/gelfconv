package gelfconv_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/m-mizutani/gelfconv"
	"github.com/stretchr/testify/require"
)

func TestEncoderSingleMessage(t *testing.T) {
	buf := bytes.Buffer{}
	enc := gelfconv.NewEncoder(&buf)

	data := map[string]interface{}{
		"k1": "v1",
		"k4": []int{1, 2, 3},
	}
	m := gelfconv.NewMessage("five")
	m.SetData(data)

	err := enc.Write(m)
	require.NoError(t, err)
	raw := buf.Bytes()
	require.NotEqual(t, 0, len(raw))
	assert.NotEqual(t, -1, bytes.Index(raw, []byte("k1")))
	assert.Equal(t, -1, bytes.Index(raw, []byte{0}))
}

func TestEncoderMultipleMessage(t *testing.T) {
	buf := bytes.Buffer{}
	enc := gelfconv.NewEncoder(&buf)

	d1 := map[string]interface{}{
		"k1": "v1",
	}
	m1 := gelfconv.NewMessage("five")
	m1.SetData(d1)
	d2 := map[string]interface{}{
		"k1": "v2",
	}
	m2 := gelfconv.NewMessage("six")
	m2.SetData(d2)

	err := enc.Write(m1)
	require.NoError(t, err)
	err = enc.Write(m2)
	require.NoError(t, err)

	raw := buf.Bytes()
	require.NotEqual(t, 0, len(raw))
	assert.NotEqual(t, -1, bytes.Index(raw, []byte("v1")))
	assert.NotEqual(t, -1, bytes.Index(raw, []byte("v2")))
	assert.NotEqual(t, -1, bytes.Index(raw, []byte{0}))
}
