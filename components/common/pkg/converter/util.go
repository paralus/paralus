package converter

import (
	"encoding/json"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/log"
)

var _log = log.GetLogger()

func ConvertToJsonRawMessage(data interface{}) json.RawMessage {
	bytes, err := json.Marshal(data)
	if err != nil {
		_log.Errorw("failed to marshal", "err", err, "data", data)
	}
	return json.RawMessage(bytes)
}

func ConvertToObject(data []byte, dest interface{}) interface{} {
	err := json.Unmarshal(data, &dest)
	if err != nil {
		_log.Errorw("failed to unmarshal", "err", err, "data", data)
	}
	return dest
}
