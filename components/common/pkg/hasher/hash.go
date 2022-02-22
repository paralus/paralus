package hasher

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"sort"

	jsoniter "github.com/json-iterator/go"
	"github.com/speps/go-hashids"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sapijson "sigs.k8s.io/kustomize/pseudo/k8s/apimachinery/pkg/runtime/serializer/json"
)

var hd *hashids.HashID

const (
	// ObjectHash is the hash of the object processed by Rafay
	ObjectHash = "rafay.dev/object-hash"
)

func init() {
	hd, _ = hashids.NewWithData(&hashids.HashIDData{
		Alphabet:  "abcdefghijklmnopqrstuvwxyz1234567890",
		Salt:      "***REMOVED***",
		MinLength: 7,
	})
}

// IDFromString returns new RafayID for hash id string
func IDFromString(hashID string) (int64, error) {
	ids, err := hd.DecodeInt64WithError(hashID)
	if err != nil {
		return -1, err
	}

	if len(ids) < 1 {
		return -1, fmt.Errorf("no ids could be constructed from string %s", hashID)
	}
	return ids[0], nil
}

// HashFromInt64 returns new RafayID for hash id string
func HashFromInt64(id int64) (string, error) {
	stringHash, err := hd.EncodeInt64([]int64{id})
	if err != nil {
		return "", err
	}
	return stringHash, nil
}

// HashFromHex returns new Hash for uuid string
func HashFromHex(id string) (string, error) {
	stringHash, err := hd.EncodeHex(id)
	if err != nil {
		return "", err
	}
	return stringHash, nil
}

// HashFromHex returns new Hash for uuid string
func IDFromHash(hash string) (string, error) {
	id, err := hd.DecodeHex(hash)
	if err != nil {
		return "", err
	}
	return id, nil
}

var json = k8sapijson.CaseSensitiveJsonIterator()

// GetHash returns the hash of spec/data for a kubernetes style object
func GetHash(o interface{}) (string, error) {
	var b []byte
	var err error
	b, err = json.Marshal(o)
	if err != nil {
		return "", err
	}

	bb := new(bytes.Buffer)

	iter := jsoniter.ParseBytes(json, b)
	for field := iter.ReadObject(); field != ""; field = iter.ReadObject() {
		switch field {
		case "data", "binaryData", "spec":
			bb.Write(iter.SkipAndReturnBytes())
		default:
			iter.Skip()
		}
	}

	h := sha256.New()
	h.Write(bb.Bytes())

	return fmt.Sprintf("%x", h.Sum(nil)), nil

}

//Add adds object hash to rutime object
func Add(o metav1.Object) error {

	hash, err := GetHash(o)
	if err != nil {
		return err
	}

	annotations := o.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[ObjectHash] = hash
	o.SetAnnotations(annotations)

	return nil
}

// Node Hash should take Labels into the hash calculation since it can be
// set from either side: From core or from cluster
func GetNodeHashFrom(labels map[string]string, taints []v1.Taint, unscheduleable bool) (string, error) {
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
	taintMap := make(map[string]v1.Taint)
	for _, taint := range taints {
		taintKeys = append(taintKeys, taint.Key)
		taintMap[taint.Key] = taint
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
