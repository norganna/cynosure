package common

import "encoding/json"

// ObjMap is a map[string]interface{} that has a JSON method.
type ObjMap map[string]interface{}

// JSON returns the json encoded string from the map (or "").
func (o *ObjMap) JSON() string {
	return toJSON(o)
}

// StringMap is a map[string]string that has a JSON method.
type StringMap map[string]string

// JSON returns the json encoded string from the map (or "").
func (o *StringMap) JSON() string {
	return toJSON(o)
}

func toJSON(o interface{}) string {
	if o == nil {
		return ""
	}

	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return ""
	}
	return string(data)
}
