package health

import "fmt"
import "encoding/json"

// Stream is a representation of a health stream for health synchronization
type Stream struct {
	Urn       string `json:"urn"`
	SubStream string `json:"sub_stream_id,omitempty"`
}

// UnmarshalJSON decodes Stream in a way to be compatible with Python check (sub_stream -> SubStream)
// while keeping original encoding (SubStream `json:sub_stream_id`) compatible with intake API
func (i *Stream) UnmarshalJSON(buf []byte) error {
	data := map[string]string{}
	if err := json.Unmarshal(buf, &data); err != nil {
		return err
	}
	urn, ok := data["urn"]
	if !ok {
		return fmt.Errorf("urn is missing")
	}
	i.Urn = urn
	subStream, ok := data["sub_stream_id"]
	if !ok {
		subStream, _ = data["sub_stream"] // override if something is there
	}
	i.SubStream = subStream
	return nil
}

// GoString prints as string, can also be used in maps
func (i *Stream) GoString() string {
	b, err := json.Marshal(i)
	if err != nil {
		fmt.Println(err)
		return fmt.Sprintf("{\"error\": \"%s\"}", err.Error())
	}
	return string(b)
}
