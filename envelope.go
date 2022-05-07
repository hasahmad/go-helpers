package helpers

import "encoding/json"

type Envelope map[string]interface{}

func (e Envelope) Marshal() ([]byte, error) {
	js, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	return js, nil
}
