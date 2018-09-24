package cwl

import (
	"encoding/json"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
)

func mapstringer(in string) (out common.MapStr, err error) {
	i := map[string]interface{}{}
	if err = json.Unmarshal([]byte(in), &i); err != nil {
		logp.Debug("", "Failed unmarshalling: %s", err)
		return
	}

	out = common.MapStr{}
	for k, v := range i {
		out[k] = v
	}

	return
}
