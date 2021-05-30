package tests

import (
	"Conjure/pkg/handler"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/structs"
	"github.com/stretchr/testify/require"
	"testing"
)

type TestYaml struct {
	Host    string   `yaml:"host"`
	Port    string   `yaml:"port"`
	Options []string `yaml:"options"`
	Secret  string   `yaml:"secret"`
}

func getTestFile() ([]byte, error) {
	t := TestYaml{
		Host: "${test-1.host}",
		Port: "${test-1.port}",
		Options: []string{
			"${test-1.option1}",
		},
		Secret: "${test-inherit.secret}",
	}

	k := koanf.New(".")

	if err := k.Load(structs.Provider(t, "koanf"), yaml.Parser()); err != nil {
		return nil, err
	}

	return k.Marshal(yaml.Parser())
}

func TestSingleYaml(t *testing.T) {

	config := handler.ConjureFileHandler{
		Config: &handler.ConjureConfig{
			Files: []*handler.ConjureFile{
				{
					Id:     "some-file-id",
					Output: "",
					Path:   "",
				},
			},
			Tags: nil,
			Groups: []*handler.ConjureGroups{
				{
					Id: "test-1",
					Items: []struct {
						Id    string `koanf:"id" validate:"required"`
						Value string `koanf:"value" validate:"required"`
					}{
						{
							Id:    "host",
							Value: "0.0.0.0",
						},
						{
							Id:    "port",
							Value: "8000",
						},
						{
							Id:    "option1",
							Value: "someOption",
						},
					},
				},
			},
		},
	}

	t.Run("file without inheritance", func(t *testing.T) {
		h, err := handler.New(handler.WithFile(config, nil), handler.WithFileType(handler.FileStruct), handler.WithParser(handler.Yaml))
		require.Nil(t, err)
		require.NotNil(t, h)

		h, err = h.BuildConjureFile()
		require.Nil(t, err)
		require.NotNil(t, h)

	})

}
