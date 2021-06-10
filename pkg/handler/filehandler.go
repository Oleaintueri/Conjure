package handler

import (
	"fmt"
	"github.com/Oleaintueri/Conjure/internal"
	"github.com/go-playground/validator"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/providers/structs"
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
		Id    string `koanf:"id" validate:"required"`
		Value interface{} `koanf:"value" validate:"required"`
	} `koanf:"items"`
}

type ConjureConfig struct {
	Inherit string           `koanf:"inherit"`
	Files   []*ConjureFile   `koanf:"files" validate:"required"`
	Tags    []*ConjureTags   `koanf:"tags"`
	Groups  []*ConjureGroups `koanf:"groups" validate:"required"`
}

type ConjureFileType int

const (
	FilePath ConjureFileType = iota
	FileBytes
	FileStruct
)

type ConjureParserType int

const (
	Yaml ConjureParserType = iota
	Json
	Toml
)

type options struct {
	conjureType ConjureFileType
	parserType  ConjureParserType
	conjureFile interface{}
	targetType  interface{}
}

type Options interface {
	apply(*options)
}

func (c ConjureFileType) apply(opts *options) {
	opts.conjureType = c
}

func (p ConjureParserType) apply(opts *options) {
	opts.parserType = p
}

type conjureFileOption struct {
	file       interface{}
	targetType interface{}
}

func (c conjureFileOption) apply(opts *options) {
	opts.conjureFile = c.file
	opts.targetType = c.targetType
}

func WithFile(file interface{}, target interface{}) Options {
	return conjureFileOption{
		file:       file,
		targetType: target,
	}
}

func WithParser(parser ConjureParserType) Options {
	return parser
}

func WithFileType(fileType ConjureFileType) Options {
	return fileType
}

type ConjureFileHandler struct {
	Parent  *ConjureFileHandler
	Config  *ConjureConfig
	k       *koanf.Koanf
	options *options
}

func New(opts ...Options) (*ConjureFileHandler, error) {
	options := &options{}

	for _, o := range opts {
		o.apply(options)
	}

	if options.conjureFile == nil {
		return nil, fmt.Errorf("the conjure file cannot be nil")
	}

	var conjureType ConjureFileType
	var parserType ConjureParserType

	var parser koanf.Parser
	var provider koanf.Provider

	switch conjureType {
	case FilePath:
		provider = file.Provider(options.conjureFile.(string))
	case FileBytes:
		provider = rawbytes.Provider(options.conjureFile.([]byte))
	case FileStruct:
		provider = structs.Provider(options.conjureFile, "koanf")
	default:
		return nil, fmt.Errorf("passing unsupported conjureType")
	}

	switch parserType {
	case Yaml:
		parser = yaml.Parser()
	case Json:
		parser = json.Parser()
	case Toml:
		parser = toml.Parser()
	default:
		parser = nil
	}

	k := koanf.New(".")

	if err := k.Load(provider, parser); err != nil {
		return nil, err
	}

	return &ConjureFileHandler{
		k:       k,
		options: options,
	}, nil
}

func (handler *ConjureFileHandler) ReadAny() (interface{}, error) {
	target := &handler.options.targetType
	if err := handler.k.Unmarshal("", target); err != nil {
		return nil, err
	}
	return target, nil
}

// BuildConjureFile
// recursively build the conjure file
func (handler *ConjureFileHandler) BuildConjureFile() (*ConjureFileHandler, error) {
	validate := validator.New()

	config := &ConjureConfig{}

	if err := handler.k.Unmarshal("", config); err != nil {
		return nil, err
	}

	// validate if the struct is correctly mapped
	if err := validate.Struct(config); err != nil {
		return nil, err
	}

	var parent *ConjureFileHandler
	var parentFileHandler *ConjureFileHandler

	if config.Inherit != "" {
		parentAny, err := internal.FromGOB64(config.Inherit)
		if err != nil {
			configPath := handler.options.conjureFile.(string)

			if !strings.Contains(config.Inherit, "/") {
				configPath = path.Join(path.Dir(configPath), config.Inherit)
			} else {
				configPath = config.Inherit
			}

			if parentFileHandler, err = New(WithParser(handler.options.parserType), WithFileType(handler.options.conjureType), WithFile(configPath, nil)); err != nil {
				return nil, err
			}

			if parent, err = parentFileHandler.BuildConjureFile(); err != nil {
				return nil, err
			}

		} else {
			parent = parentAny.(*ConjureFileHandler)
		}

	}

	return &ConjureFileHandler{
		Parent:  parent,
		Config:  config,
		k:       handler.k,
		options: handler.options,
	}, nil
}
