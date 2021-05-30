package internal

import (
	"net/url"
	"regexp"
)

// ExtractIdTag
// A utility function to extract the id and tag values from a string
// e.g. someId, tagValue := extractIdTag("someId<tagValue>")
func ExtractIdTag(idTag string) (id string, tags []string) {
	re := regexp.MustCompile(`<(.*)>`)

	tagArr := re.FindStringSubmatch(idTag)

	id = re.ReplaceAllString(idTag, "")

	if len(tagArr) > 1 {
		tags = tagArr
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