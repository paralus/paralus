package hasher

import (
	"crypto/sha256"
	"fmt"
	"sort"

	commonv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
)

// Node Hash should take Labels into the hash calculation since it can be
// set from either side: From core or from cluster
func GetNodeHashFrom(labels map[string]string, taints []*commonv3.Taint, unscheduleable bool) (string, error) {
	//add sorted labels
	labelsKeys := make([]string, 0)
	for k, _ := range labels {
		labelsKeys = append(labelsKeys, k)
	}
	sort.Strings(labelsKeys)
	finalLabelsAsString := ""
	for _, k := range labelsKeys {
		finalLabelsAsString += fmt.Sprintf("%s:%s,", k, labels[k])
	}
	//add sorted taints
	taintKeys := make([]string, 0)
	taintMap := make(map[string]commonv3.Taint)
	for _, taint := range taints {
		taintKeys = append(taintKeys, taint.Key)
		taintMap[taint.Key] = *taint
	}
	sort.Strings(taintKeys)
	finalTaintsAsString := ""
	for _, k := range taintKeys {
		finalTaintsAsString += fmt.Sprintf("%s:%s:%s,", k, taintMap[k].Value, taintMap[k].Effect)
	}
	finalHashString := fmt.Sprintf("labels:%s,taints:%s,unschedulable:%v", finalLabelsAsString, finalTaintsAsString, unscheduleable)
	h := sha256.New()
	h.Write([]byte(finalHashString))
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
