package refs

import (
	"encoding/json"
	"time"
)

// Millisecs is used to marshal and unmarshal time as a JSON number representing
// a timestamp in milliseconds.
type Millisecs time.Time

func (t *Millisecs) UnmarshalJSON(in []byte) error {
	var milliseconds float64
	if err := json.Unmarshal(in, &milliseconds); err != nil {
		return err
	}
	*t = Millisecs(time.UnixMilli(int64(milliseconds)))
	return nil
}

func (t Millisecs) MarshalJSON() ([]byte, error) {
	milliseconds := time.Time(t).UnixMilli()
	return json.Marshal(milliseconds)
}
