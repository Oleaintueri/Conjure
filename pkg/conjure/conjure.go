package conjure

import (
	internal2 "Conjure/internal"
	"Conjure/pkg/handler"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
)

type ConjureRecall struct {}

type Conjure struct {
	*handler.ConjureFileHandler
	recalls []handler.ConjureFile
}

func New(config *handler.ConjureFileHandler) (*Conjure, error) {
	return &Conjure{
		config,
		[]handler.ConjureFile{},
	}, nil
}

func (conjure *Conjure) Recall() error {
	for _, targetFile := range conjure.Config.Files {
		targetName, targetTags := internal2.ExtractIdTag(targetFile.Output)
		targetFile.CurrentName = targetName
		if len(targetTags) == 1 && targetTags[0] == "tags" {
			backupTemplate := targetFile.FileData
			for _, customTags := range conjure.Config.Tags {
				targetFile.CurrentTag = customTags.Id
				if err := conjure.recall(conjure.ConjureFileHandler, targetFile); err != nil {
					return err
				}
				/*if err := conjure.writeFile(targetFile); err != nil {
					return err
				}*/
				conjure.recalls = append(conjure.recalls, *targetFile)

				targetFile.FileData = backupTemplate
			}
		} else if len(targetTags) > 0 {
			backupTemplate := targetFile.FileData
			for _, tag := range targetTags {
				targetFile.CurrentTag = tag
				if err := conjure.recall(conjure.ConjureFileHandler, targetFile); err != nil {
					return err
				}
				/*if err := conjure.writeFile(targetFile); err != nil {
					return err
				}*/
				conjure.recalls = append(conjure.recalls, *targetFile)
				targetFile.FileData = backupTemplate
			}
		} else {
			backupTemplate := targetFile.FileData
			targetFile.CurrentTag = ""
			if err := conjure.recall(conjure.ConjureFileHandler, targetFile); err != nil {
				return err
			}
			/*if err := conjure.writeFile(targetFile); err != nil {
				return err
			}*/
			conjure.recalls = append(conjure.recalls, *targetFile)
			targetFile.FileData = backupTemplate
		}

	}
	return nil
}

func (conjure *Conjure) recall(inheritance *handler.ConjureFileHandler, targetFile *handler.ConjureFile) error {
	var err error

	if inheritance.Parent != nil {
		if err := conjure.recall(inheritance.Parent, targetFile); err != nil {
			return err
		}
	}

	if len(targetFile.FileData) == 0 {
		if targetFile.FileData, err = ioutil.ReadFile(targetFile.Path); err != nil {
			return err
		}
	}

	targetFile.FileData = conjure.extract(inheritance, targetFile.FileData, targetFile.CurrentTag)

	return err
}

// extract
// return the targetFile data and related tag if it exists
func (conjure *Conjure) extract(inheritance *handler.ConjureFileHandler, targetFile []byte, targetTag string) []byte {
	var tags []string
	var groupId string

	for _, group := range inheritance.Config.Groups {
		groupId, tags = internal2.ExtractIdTag(group.Id)

		if tags == nil {
			// and any groups with no tags
			tags = append(tags, "")
		}

		for _, tag := range tags {
			if tag == targetTag {
				for _, item := range group.Items {
					re := regexp.MustCompile(fmt.Sprintf(`\${%s\.%s}`, groupId, item.Id))
					targetFile = re.ReplaceAll(targetFile, []byte(item.Value))
				}
			}
		}

	}

	return targetFile
}

func (conjure *Conjure) writeFile(targetFile *handler.ConjureFile) error {
	var err error

	var targetFileName string

	if targetFile.CurrentTag != "" {
		tagPath := conjure.tagSearch(targetFile.CurrentTag)
		if tagPath == "" {
			tagPath = "."
		}

		if targetFileName, err = internal2.ParseFilePath(fmt.Sprintf("%s%s", tagPath, targetFile.CurrentName)); err != nil {
			return err
		}

	} else {
		if targetFileName, err = internal2.ParseFilePath(fmt.Sprintf("./%s", targetFile.CurrentName)); err != nil {
			return err
		}
	}

	if targetFileName == "./" || targetFileName == "." || targetFileName == "" {
		return fmt.Errorf("the specified targetFilename for fileId `%s` is not correctly specified", targetFile.Id)
	}

	parentPath := path.Dir(targetFileName)
	_ = os.MkdirAll(parentPath, 0700)

	err = ioutil.WriteFile(targetFileName, targetFile.FileData, 0700)

	return err
}

func (conjure *Conjure) tagSearch(search string) string {
	return tagSearchRecursive(conjure.ConjureFileHandler, search)
}

func tagSearchRecursive(inheritance *handler.ConjureFileHandler, search string) string {
	var found string
	if inheritance.Parent != nil {
		found = tagSearchRecursive(inheritance.Parent, search)
	}

	if found != "" {
		return found
	}

	for _, tag := range inheritance.Config.Tags {
		if tag.Id == search {
			return tag.Path
		}
	}

	return ""
}