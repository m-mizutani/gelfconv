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
var recursiveLimit = 3

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
	data         interface{}
	kvList       []keyValuePair
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

// AddField adds additional pair of key and value. The pair will overwrite key from .data
func (x *Message) AddField(key string, value interface{}) {
	x.kvList = append(x.kvList, keyValuePair{key, value})
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

	kvList := toKeyValuePairs(x.data, "", 0)
	for _, kv := range kvList {
		if kv.key == "" {
			v["_value"] = kv.value
		} else {
			key := kv.key
			if key == "_version" || key == "_host" || key == "_short_message" ||
				key == "_timestamp" || key == "_full_message" || key == "_level" ||
				key == "_message" {
				key = "_" + key
			}

			v[key] = kv.value
		}
	}

	for _, kv := range x.kvList {
		v["_"+kv.key] = kv.value
	}

	d, err := json.Marshal(v)
	if err != nil {
		return []byte{}, err
	}

	return d, nil
}

func toKeyStringPairs(v interface{}, key string) []keyValuePair {
	raw, err := json.Marshal(v)
	if err != nil {
		return []keyValuePair{}
	}
	return []keyValuePair{{key, string(raw)}}

}

func toKeyValuePairs(v interface{}, keyPrefix string, depth int) []keyValuePair {
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
		if depth >= recursiveLimit {
			return toKeyStringPairs(v, keyPrefix)
		}

		var kvList []keyValuePair

		keys := value.MapKeys()
		for i := 0; i < value.Len(); i++ {
			mValue := value.MapIndex(keys[i])
			key := fmt.Sprintf("%s_%s", keyPrefix, keys[i])
			kvList = append(kvList, toKeyValuePairs(mValue.Interface(), key, depth+1)...)
		}
		return kvList

	case reflect.Struct:
		if depth >= recursiveLimit {
			return toKeyStringPairs(v, keyPrefix)
		}

		t := value.Type()
		var pList []keyValuePair

		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			vdata := value.FieldByName(f.Name)
			if !vdata.CanInterface() {
				continue
			}

			jsonTag := f.Tag.Get("json")

			fname := f.Name
			if jsonTag != "" {
				fname = jsonTag
			}
			newKeyPrefix := fmt.Sprintf("%s_%s", keyPrefix, fname)
			pList = append(pList, toKeyValuePairs(vdata.Interface(), newKeyPrefix, depth+1)...)
		}

		return pList

	case reflect.Array, reflect.Slice:
		return toKeyStringPairs(v, keyPrefix)

	case reflect.Ptr, reflect.UnsafePointer:
		return toKeyValuePairs(value.Elem().Interface(), keyPrefix, depth)

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
