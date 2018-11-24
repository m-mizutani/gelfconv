package gelfconv

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/pkg/errors"
)

var defaultHostname string

func init() {
	updateDefaultHostname()
}

func updateDefaultHostname() {
	hostname, err := os.Hostname()
	if err == nil {
		defaultHostname = hostname
	}
}

// Message is a structure to store GELF metadata and log values.
type Message struct {
	data interface{}

	Host         string
	ShortMessage string
	FullMessage  string
	Level        int
	Timestamp    time.Time
}

// NewMessage is a constructor of GELF message structure
func NewMessage(message string) Message {
	return Message{
		Timestamp:    time.Now().UTC(),
		Host:         defaultHostname,
		ShortMessage: message,
	}
}

// SetData is setter of GELF additional fields for structured data (struct or interface)
func (x *Message) SetData(v interface{}) {
	x.data = v
}

// SetJSON is setter of GELF additional fields for encoded data in JSON
func (x *Message) SetJSON(jdata []byte) error {
	var v interface{}
	err := json.Unmarshal(jdata, &v)
	if err != nil {
		return errors.Wrap(err, "Fail to unmarshal json string for GELF")
	}

	x.data = v
	return nil
}

// Gelf returns GELF encoded byte data
func (x *Message) Gelf() ([]byte, error) {
	v := map[string]interface{}{
		"version":       "1.1",
		"host":          x.Host,
		"short_message": x.ShortMessage,
		"timestamp":     x.Timestamp.Unix(),
	}

	if x.FullMessage != "" {
		v["full_message"] = x.FullMessage
	}
	if x.Level > 0 {
		v["level"] = x.Level
	}

	kvList := toKeyValuePairs(x.data, "")
	for _, kv := range kvList {
		if kv.key == "" {
			v["_value"] = kv.value
		} else {
			v[kv.key] = kv.value
		}
	}

	d, err := json.Marshal(v)
	if err != nil {
		return []byte{}, err
	}
	d = append(d, 0)
	return d, nil
}

func toKeyValuePairs(v interface{}, keyPrefix string) []keyValuePair {
	value := reflect.ValueOf(v)

	switch value.Kind() {
	case reflect.Bool:
		b, _ := v.(bool)
		// Convert to number.
		var d int
		if b {
			d = 1
		} else {
			d = 0
		}
		return []keyValuePair{{keyPrefix, d}}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr, reflect.Complex64, reflect.Complex128,
		reflect.Float32, reflect.Float64, reflect.String:
		return []keyValuePair{{keyPrefix, v}}

	case reflect.Map:
		var kvList []keyValuePair

		keys := value.MapKeys()
		for i := 0; i < value.Len(); i++ {
			mValue := value.MapIndex(keys[i])
			key := fmt.Sprintf("%s_%s", keyPrefix, keys[i])
			kvList = append(kvList, toKeyValuePairs(mValue.Interface(), key)...)
		}
		return kvList

	case reflect.Struct:
		jdata, err := json.Marshal(v)
		if err != nil {
			return []keyValuePair{}
		}
		var vdata interface{}
		err = json.Unmarshal(jdata, &vdata)
		if err != nil {
			return []keyValuePair{}
		}

		return toKeyValuePairs(vdata, keyPrefix)

	case reflect.Array, reflect.Slice:
		raw, err := json.Marshal(v)
		if err != nil {
			return []keyValuePair{}
		}
		return []keyValuePair{{keyPrefix, string(raw)}}

	default: // will be ignored
		// Expected:
		// reflect.Chan, reflect.Interface, reflect.Ptr, reflect.Func,
		// reflect.UnsafePointer, reflect.Invalid
		return []keyValuePair{} // returns empty list.
	}

}

type keyValuePair struct {
	key   string
	value interface{}
}
