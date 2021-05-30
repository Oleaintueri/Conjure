package internal

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"net/url"
	"regexp"
)

func removeElement(s []string, i int) []string {
	s[i] = s[len(s)-1]
	// We do not need to put s[i] at the end, as it will be discarded anyway
	return s[:len(s)-1]
}

func recursiveRemove(s []string, index int) []string{
	var o []string

	if len(s) == 1 {
		return s
	}

	o = removeElement(s, index)

	return recursiveRemove(o, index + 1)
}

// ExtractIdTag
// A utility function to extract the id and tag values from a string
// e.g. someId, tagValue := extractIdTag("someId<tagValue>")
func ExtractIdTag(idTag string) (id string, tags []string) {
	re := regexp.MustCompile(`<(\w+)>`)

	tagArr := re.FindAllStringSubmatch(idTag, -1)

	id = re.ReplaceAllString(idTag, "")

	if len(tagArr) > 0 {
		for _, inner := range tagArr {
			tags = append(tags, recursiveRemove(inner, 0)...)
		}
	}

	return
}

func ParseFilePath(fileName string) (string, error) {
	targetFileUrl, err := url.Parse(fileName)
	if err != nil {
		return "", err
	}
	return targetFileUrl.EscapedPath(), nil
}

// go binary encoder
func ToGOB64(config interface{}) (string, error) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)

	if err := e.Encode(config); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b.Bytes()), nil
}

// go binary decoder
func FromGOB64(str string) (interface{}, error) {
	var config interface{}

	by, err := base64.StdEncoding.DecodeString(str)

	if err != nil {
		return nil, err
	}

	b := bytes.Buffer{}
	b.Write(by)
	d := gob.NewDecoder(&b)

	if err = d.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}