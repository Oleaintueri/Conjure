package conjure

import (
	"Conjure/pkg/internal"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
)

type ConfigFile struct {
	Id          string `yaml:"id"`
	Path        string `yaml:"path"`
	Output      string `yaml:"output"`
	fileData    []byte
	currentTag  string
	currentName string
}

type ConfigTags struct {
	Id   string `yaml:"id"`
	Path string `yaml:"path"`
}

type ConfigGroups struct {
	Id    string `yaml:"id"`
	Items []struct {
		Id    string `yaml:"id"`
		Value string `yaml:"value"`
	} `yaml:"items"`
}

type Config struct {
	Inherit string          `yaml:"inherit"`
	Files   []*ConfigFile   `yaml:"files"`
	Tags    []*ConfigTags   `yaml:"tags"`
	Groups  []*ConfigGroups `yaml:"groups"`
}

type configInheritance struct {
	parent *configInheritance
	config *Config
}

type Conjure struct {
	config *configInheritance
}

func New(configPath string) (*Conjure, error) {
	inheritance, err := buildConfigInheritance(configPath, nil)

	if err != nil {
		return nil, err
	}

	return &Conjure{
		config: inheritance,
	}, nil
}

func buildConfigInheritance(configPath string, inheritance *configInheritance) (*configInheritance, error) {
	config := &Config{}

	err := internal.ParseYaml(configPath, config)

	if err != nil {
		return nil, err
	}

	if config.Inherit != "" {
		if !strings.Contains(config.Inherit, "/") {
			configPath = path.Join(path.Dir(configPath), config.Inherit)
		} else {
			configPath = config.Inherit
		}

		if inheritance, err = buildConfigInheritance(configPath, inheritance); err != nil {
			return nil, err
		}
	}

	return &configInheritance{
		parent: inheritance,
		config: config,
	}, nil
}

func (conjure *Conjure) Recall() error {
	for _, targetFile := range conjure.config.config.Files {
		targetName, targetTags := internal.ExtractIdTag(targetFile.Output)
		targetFile.currentName = targetName
		if len(targetTags) == 1 && targetTags[0] == "tags" {
			backupTemplate := targetFile.fileData
			for _, customTags := range conjure.config.config.Tags {
				targetFile.currentTag = customTags.Id
				if err := recall(conjure.config, targetFile); err != nil {
					return err
				}
				if err := conjure.writeFile(targetFile); err != nil {
					return err
				}
				targetFile.fileData = backupTemplate
			}
		} else {
			for _, tag := range targetTags {
				targetFile.currentTag = tag
				if err := recall(conjure.config, targetFile); err != nil {
					return err
				}
				if err := conjure.writeFile(targetFile); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func recall(inheritance *configInheritance, targetFile *ConfigFile) error {
	var err error

	if inheritance.parent != nil {
		if err := recall(inheritance.parent, targetFile); err != nil {
			return err
		}
	}

	if len(targetFile.fileData) == 0 {
		if targetFile.fileData, err = ioutil.ReadFile(targetFile.Path); err != nil {
			return err
		}
	}

	targetFile.fileData = inheritance.extract(targetFile.fileData, targetFile.currentTag)

	return err
}

// extract
// return the targetFile data and related tag if it exists
func (inheritance *configInheritance) extract(targetFile []byte, targetTag string) []byte {
	var tags []string
	var groupId string

	for _, group := range inheritance.config.Groups {
		groupId, tags = internal.ExtractIdTag(group.Id)

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

func (conjure *Conjure) writeFile(targetFile *ConfigFile) error {
	var err error

	var targetFileName string

	if targetFile.currentTag != "" {
		tagPath := conjure.tagSearch(targetFile.currentTag)
		if tagPath == "" {
			tagPath = "."
		}

		if targetFileName, err = internal.ParseFilePath(fmt.Sprintf("%s%s", tagPath, targetFile.currentName)); err != nil {
			return err
		}

	} else {
		if targetFileName, err = internal.ParseFilePath(fmt.Sprintf("./%s", targetFile.currentName)); err != nil {
			return err
		}
	}

	if targetFileName == "./" || targetFileName == "." || targetFileName == "" {
		return fmt.Errorf("the specified targetFilename for fileId `%s` is not correctly specified", targetFile.Id)
	}

	parentPath := path.Dir(targetFileName)
	_ = os.MkdirAll(parentPath, 0700)

	err = ioutil.WriteFile(targetFileName, targetFile.fileData, 0700)

	return err
}

func (conjure *Conjure) tagSearch(search string) string {
	return tagSearchRecursive(conjure.config, search)
}

func tagSearchRecursive(inheritance *configInheritance, search string) string {
	var found string
	if inheritance.parent != nil {
		found = tagSearchRecursive(inheritance.parent, search)
	}

	if found != "" {
		return found
	}

	for _, tag := range inheritance.config.Tags {
		if tag.Id == search {
			return tag.Path
		}
	}

	return ""
}
