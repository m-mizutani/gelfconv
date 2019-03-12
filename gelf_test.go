package gelfconv_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/m-mizutani/gelfconv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	data := map[string]interface{}{
		"k1": "v1",
		"k2": map[string]string{
			"k3": "v3",
		},
		"k4": []int{1, 2, 3},
	}
	m := gelfconv.NewMessage("five")
	m.SetData(data)
	raw, err := m.Gelf()

	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(raw))
	assert.NotEqual(t, uint8(0), raw[len(raw)-1])

	jdata := raw
	var v map[string]interface{}
	err = json.Unmarshal(jdata, &v)
	require.NoError(t, err)

	v0, ok := v["short_message"].(string)
	assert.Equal(t, true, ok)
	assert.Equal(t, "five", v0)

	v1, ok := v["_k1"].(string)
	assert.Equal(t, true, ok)
	assert.Equal(t, v1, "v1")

	v2, ok := v["_k4"].(string)
	assert.Equal(t, true, ok)
	assert.Equal(t, "[1,2,3]", v2)
}

func toMap(t *testing.T, v interface{}) map[string]interface{} {
	m := gelfconv.NewMessage("test")
	m.SetData(v)
	raw, err := m.Gelf()

	require.NoError(t, err)
	jdata := raw

	var vmap map[string]interface{}
	err = json.Unmarshal(jdata, &vmap)
	require.NoError(t, err)

	return vmap
}

func TestInteger(t *testing.T) {
	var d int = 10
	vmap := toMap(t, d)
	// pp.Println(vmap)

	v, ok := vmap["_value"].(float64)
	require.Equal(t, true, ok)
	require.Equal(t, float64(10), v)
}

func TestFloat(t *testing.T) {
	f := 3.14
	vmap := toMap(t, f)
	// pp.Println(vmap)

	v, ok := vmap["_value"].(float64)
	require.Equal(t, true, ok)
	require.Equal(t, 3.14, v)
}

func TestStruct(t *testing.T) {
	type sample struct {
		Str      string `json:"str"`
		Integer  int    `json:"integer"`
		NoTag    string
		noExport string
	}
	s := sample{"blue", 5, "no_name", "omg"}
	vmap := toMap(t, s)

	v1, ok := vmap["_str"].(string)
	require.Equal(t, true, ok)
	require.Equal(t, "blue", v1)

	v2, ok := vmap["_integer"].(float64)
	require.Equal(t, true, ok)
	require.Equal(t, 5.0, v2)

	v3, ok := vmap["_NoTag"].(string)
	require.Equal(t, true, ok)
	require.Equal(t, "no_name", v3)

	v4, ok := vmap["_noExport"].(string)
	require.Equal(t, false, ok)
	require.Equal(t, "", v4)
}

func TestAddrOfStruct(t *testing.T) {
	type sample struct {
		Str      string `json:"str"`
		Integer  int    `json:"integer"`
		NoTag    string
		noExport string
	}

	s := sample{"blue", 5, "no_name", "omg"}
	vmap := toMap(t, &s)

	v1, ok := vmap["_str"].(string)
	require.Equal(t, true, ok)
	require.Equal(t, "blue", v1)

	v2, ok := vmap["_integer"].(float64)
	require.Equal(t, true, ok)
	require.Equal(t, 5.0, v2)

	v3, ok := vmap["_NoTag"].(string)
	require.Equal(t, true, ok)
	require.Equal(t, "no_name", v3)

	v4, ok := vmap["_noExport"].(string)
	require.Equal(t, false, ok)
	require.Equal(t, "", v4)
}

func TestSetJSON(t *testing.T) {
	jdata := []byte(`{"k1": "blue", "k2": 5}`)
	m := gelfconv.NewMessage("test")

	err := m.SetJSON(jdata)
	require.NoError(t, err)
	raw, err := m.Gelf()
	require.NoError(t, err)
	jdata2 := raw

	var vmap map[string]interface{}
	err = json.Unmarshal(jdata2, &vmap)
	require.NoError(t, err)

	v1, ok := vmap["_k1"].(string)
	require.Equal(t, true, ok)
	require.Equal(t, "blue", v1)

	v2, ok := vmap["_k2"].(float64)
	require.Equal(t, true, ok)
	require.Equal(t, 5.0, v2)
}

func TestAddField(t *testing.T) {
	data := map[string]interface{}{
		"color": "red",
		"count": "four",
	}
	m := gelfconv.NewMessage("five")
	m.SetData(data)
	m.AddField("color", "orange")
	m.AddField("color2", "blue")
	raw, err := m.Gelf()

	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(raw))
	assert.NotEqual(t, uint8(0), raw[len(raw)-1])

	var v map[string]interface{}
	err = json.Unmarshal(raw, &v)
	require.NoError(t, err)

	v0, ok := v["short_message"].(string)
	assert.Equal(t, true, ok)
	assert.Equal(t, "five", v0)

	// Overwrite existing key
	v1, ok := v["_color"].(string)
	assert.Equal(t, true, ok)
	assert.Equal(t, "orange", v1)

	// Add a new key
	v2, ok := v["_color2"].(string)
	assert.Equal(t, true, ok)
	assert.Equal(t, "blue", v2)

	// Not changed
	v3, ok := v["_count"].(string)
	assert.Equal(t, true, ok)
	assert.Equal(t, "four", v3)

}

func TestReservedKey(t *testing.T) {
	data := map[string]interface{}{
		"timestamp": 12345,
	}
	m := gelfconv.NewMessage("five")
	m.SetData(data)
	raw, err := m.Gelf()

	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(raw))

	var v map[string]interface{}
	err = json.Unmarshal(raw, &v)
	require.NoError(t, err)

	_, ok := v["_timestamp"]
	assert.False(t, ok)
	v1, ok := v["__timestamp"]
	assert.True(t, ok)
	i1, ok := v1.(float64)
	assert.True(t, ok)
	assert.Equal(t, 12345.0, i1)
}

func TestDeepNestedData(t *testing.T) {
	data := map[string]interface{}{
		"d1": map[string]interface{}{
			"d2": map[string]interface{}{
				"d3": map[string]interface{}{
					"d4": map[string]interface{}{
						"d5": "five",
					},
				},
			},
		},
	}

	vmap := toMap(t, data)
	_, ok := vmap["_d1_d2_d3_d4"]
	assert.False(t, ok)
	v, ok := vmap["_d1_d2_d3"].(string)
	assert.True(t, ok)
	assert.True(t, strings.Index(v, "five") >= 0)
}
