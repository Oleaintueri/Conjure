package internal

import (
	"fmt"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/providers/structs"
	"log"
	"path"
	"reflect"
	"strconv"
)

type ConjureFileType int

const (
	FilePath ConjureFileType = iota
	FileBytes
	FileStruct
	FileMap
)

type ConjureParserType int

const (
	Yaml ConjureParserType = iota
	Json
	Toml
)

type options struct {
	conjureType *ConjureFileType
	parserType  *ConjureParserType
	conjureFile interface{}
	targetType  interface{}
}

type KoanfWrapperOptions interface {
	apply(*options)
}

func (c ConjureFileType) apply(opts *options) {
	opts.conjureType = &c
}

func (p ConjureParserType) apply(opts *options) {
	opts.parserType = &p
}

type conjureFileOption struct {
	file       interface{}
	targetType interface{}
}

func (c conjureFileOption) apply(opts *options) {
	opts.conjureFile = c.file
	opts.targetType = c.targetType
}

func WithFile(file interface{}, target interface{}) KoanfWrapperOptions {
	return conjureFileOption{
		file:       file,
		targetType: target,
	}
}

func WithParser(parser ConjureParserType) KoanfWrapperOptions {
	return parser
}

func WithFileType(fileType ConjureFileType) KoanfWrapperOptions {
	return fileType
}

type KoanfWrapper struct {
	options *options
	*koanf.Koanf
}

func New(opts ...KoanfWrapperOptions) (*KoanfWrapper, error) {
	options := &options{
		conjureType: nil,
		parserType: nil,
		conjureFile: nil,
		targetType: nil,
	}

	for _, o := range opts {
		o.apply(options)
	}

	if options.conjureFile == nil {
		return nil, fmt.Errorf("the conjure file cannot be nil")
	}

	var parser koanf.Parser
	var provider koanf.Provider

	if options.parserType != nil {
		parser = getParserType(*options.parserType)
	}

	if options.conjureType != nil {
		provider = getProvider(*options.conjureType, options.conjureFile)
	}

	k := koanf.New(".")

	if err := k.Load(provider, parser); err != nil {
		return nil, err
	}

	return &KoanfWrapper{
		options: options,
		Koanf:   k,
	}, nil
}

func getParserType(parserType ConjureParserType) koanf.Parser {
	switch parserType {
	case Yaml:
		return yaml.Parser()
	case Json:
		return json.Parser()
	case Toml:
		return toml.Parser()
	default:
		return nil
	}
}

func getProvider(conjureType ConjureFileType, conjureFile interface{}) koanf.Provider {
	switch conjureType {
	case FilePath:
		return file.Provider(conjureFile.(string))
	case FileBytes:
		return rawbytes.Provider(conjureFile.([]byte))
	case FileStruct:
		return structs.Provider(conjureFile, "")
	case FileMap:
		return confmap.Provider(conjureFile.(map[string]interface{}), "")
	default:
		return nil
	}
}

func (k *KoanfWrapper) ReadAny() (interface{}, error) {
	target := &k.options.targetType
	if err := k.Koanf.Unmarshal("", target); err != nil {
		return nil, err
	}
	return target, nil
}

func (k *KoanfWrapper) FilePath() string {
	val := reflect.ValueOf(k.options.conjureFile)
	switch val.Kind() {
	case reflect.String:
		return path.Dir(k.options.conjureFile.(string))
	default:
		return ""
	}
}

func (k *KoanfWrapper) ParserType() ConjureParserType {
	return *k.options.parserType
}

func (k *KoanfWrapper) ConjureType() ConjureFileType {
	return *k.options.conjureType
}

func (k *KoanfWrapper) ToBytes(value interface{}) ([]byte, error) {
	val := reflect.ValueOf(value)
	kind := val.Kind().String()
	log.Printf("kind %s", kind)

	switch val.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32:
		return []byte(strconv.FormatInt(val.Int(), 10)), nil
	case reflect.String:
		return []byte(val.String()), nil
	case reflect.Map:
		innerK, err := New(
			WithFile(value, nil),
			WithFileType(FileMap))

		if err != nil {
			return nil, err
		}

		return innerK.Marshal(getParserType(k.ParserType()))
	case reflect.Slice:
		return nil, fmt.Errorf("values cannot be a slice alone, it needs to be a map")
	}
	return nil, fmt.Errorf("could not get value bytes")
}