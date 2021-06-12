package handler

import (
	"github.com/Oleaintueri/Conjure/internal"
	"github.com/go-playground/validator"
	"log"
	"path"
	"strings"
)

type ConjureFile struct {
	Id          string `koanf:"id"`
	Path        string `koanf:"path"`
	Output      string `koanf:"output"`
	FileData    []byte
	CurrentTag  string
	CurrentName string
}

type ConjureTags struct {
	Id   string `koanf:"id" validate:"required"`
	Path string `koanf:"path" validate:"required"`
}

type ConjureGroups struct {
	Id    string `koanf:"id" validate:"required"`
	Items []struct {
		Id    string      `koanf:"id" validate:"required"`
		Value interface{} `koanf:"value" validate:"required"`
	} `koanf:"items"`
}

type ConjureConfig struct {
	Inherit string           `koanf:"inherit"`
	Files   []*ConjureFile   `koanf:"files" validate:"required"`
	Tags    []*ConjureTags   `koanf:"tags"`
	Groups  []*ConjureGroups `koanf:"groups" validate:"required"`
}

type ConjureFileHandler struct {
	Parent  *ConjureFileHandler
	Config  *ConjureConfig
	*internal.KoanfWrapper
}

type ConjureFileType internal.ConjureFileType
type ConjureParserType internal.ConjureParserType
type ConjureOptions internal.KoanfWrapperOptions

const (
	Yaml ConjureParserType = ConjureParserType(internal.Yaml)
	Json ConjureParserType = ConjureParserType(internal.Json)
	Toml ConjureParserType = ConjureParserType(internal.Toml)
)

const (
	FilePath ConjureFileType = ConjureFileType(internal.FilePath)
	FileBytes ConjureFileType = ConjureFileType(internal.FileBytes)
	FileStruct ConjureFileType = ConjureFileType(internal.FileStruct)
)

func WithFile(file interface{}, target interface{}) ConjureOptions {
	return internal.WithFile(file, target)
}

func WithParser(parser ConjureParserType) ConjureOptions {
	return internal.WithParser(internal.ConjureParserType(parser))
}

func WithFileType(fileType ConjureFileType) ConjureOptions {
	return internal.WithFileType(internal.ConjureFileType(fileType))
}

func New(opts ...ConjureOptions) (*ConjureFileHandler, error) {
	var koanfOpts []internal.KoanfWrapperOptions

	for _, opt := range opts {
		koanfOpts = append(koanfOpts, opt)
	}

	k, err := internal.New(koanfOpts...)

	if err != nil {
		return nil, err
	}

	return &ConjureFileHandler{
		KoanfWrapper: k,
	}, nil
}

// BuildConjureFile
// recursively build the conjure file
func (handler *ConjureFileHandler) BuildConjureFile() (*ConjureFileHandler, error) {
	// Create a new validator to validate the conjure file structural requirements
	validate := validator.New()

	config := &ConjureConfig{}

	log.Printf("reading conjure file...")

	// unmarshal the conjure file into a struct
	if err := handler.Unmarshal("", config); err != nil {
		return nil, err
	}

	// validate if the struct is correctly mapped
	if err := validate.Struct(config); err != nil {
		return nil, err
	}

	log.Printf("conjure file validated :)")

	// since a conjure file can inherit, we will need to recursively build the parents into the config
	var parent *ConjureFileHandler
	var parentFileHandler *ConjureFileHandler

	// if the inherit field is not empty
	if config.Inherit != "" {
		// extract the value from base64 if relevant
		parentAny, err := internal.FromGOB64(config.Inherit)
		if err != nil {
			// the parent is a file instead of a base64
			configPath := handler.FilePath()

			if !strings.Contains(config.Inherit, "/") {
				configPath = path.Join(configPath, config.Inherit)
			} else {
				configPath = config.Inherit
			}

			parentFileHandler, err = New(
				WithParser(ConjureParserType(handler.ParserType())),
				WithFileType(ConjureFileType(handler.ConjureType())),
				WithFile(configPath, nil))

			if err == nil {
				if parent, err = parentFileHandler.BuildConjureFile(); err != nil {
					return nil, err
				}
			} else {
				log.Printf("inheritance failed with error: %v\n", err)
			}

		} else {
			// the parent is a base64
			parent = parentAny.(*ConjureFileHandler)
		}

	}

	return &ConjureFileHandler{
		Parent:  parent,
		Config:  config,
		KoanfWrapper: handler.KoanfWrapper,
	}, nil
}
