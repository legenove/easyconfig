package parsers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/printer"
	jsoniter "github.com/json-iterator/go"
	"github.com/legenove/easyconfig/config_tools"
	"github.com/legenove/easyconfig/ifacer"
	"github.com/magiconair/properties"
	"github.com/pelletier/go-toml"
	"github.com/spf13/afero"
	"github.com/subosito/gotenv"
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v3"
)

var SupportedParsers map[string]Parser

func init() {
	SupportedParsers = make(map[string]Parser, 32)
	AddParser(&JSONParser{}, "json")
	AddParser(&TOMLParser{}, "toml")
	AddParser(&YAMLParser{}, "yaml", "yml")
	AddParser(&PROPSParser{}, "properties", "props", "prop")
	AddParser(&HCLParser{}, "hcl")
	AddParser(&DOTENVParser{}, "dotenv", "env")
	AddParser(&INIParser{}, "ini")
	AddParser(&XMLParser{}, "xml")
}

// add support parser
func AddParser(parser Parser, names ...string) {
	for _, n := range names {
		SupportedParsers[n] = parser
	}
}

type Parser interface {
	UnmarshalReader(v ifacer.Configer, in io.Reader, c map[string]interface{}) error // Unmarshal a Reader into a map.
	MarshalWriter(v ifacer.Configer, f afero.File, c map[string]interface{}) error   // Marshal a map into Writer.
}

// json parser
type JSONParser struct {
}

func (pp *JSONParser) UnmarshalReader(v ifacer.Configer, in io.Reader, c map[string]interface{}) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)
	return jsoniter.Unmarshal(buf.Bytes(), &c)
}
func (pp *JSONParser) MarshalWriter(v ifacer.Configer, f afero.File, c map[string]interface{}) error {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	_, err = f.WriteString(string(b))
	return err
}

// toml parser
type TOMLParser struct {
}

func (pp *TOMLParser) UnmarshalReader(v ifacer.Configer, in io.Reader, c map[string]interface{}) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)
	tree, err := toml.LoadReader(buf)
	if err != nil {
		return err
	}
	tmap := tree.ToMap()
	for k, v := range tmap {
		c[k] = v
	}
	return nil
}
func (pp *TOMLParser) MarshalWriter(v ifacer.Configer, f afero.File, c map[string]interface{}) error {
	t, err := toml.TreeFromMap(c)
	if err != nil {
		return err
	}
	s := t.String()
	if _, err := f.WriteString(s); err != nil {
		return err
	}
	return nil
}

// yaml parser
type YAMLParser struct {
}

func (pp *YAMLParser) UnmarshalReader(v ifacer.Configer, in io.Reader, c map[string]interface{}) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)
	return yaml.Unmarshal(buf.Bytes(), &c)
}
func (pp *YAMLParser) MarshalWriter(v ifacer.Configer, f afero.File, c map[string]interface{}) error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	if _, err = f.WriteString(string(b)); err != nil {
		return err
	}
	return nil
}

// ini parser
type INIParser struct {
}

func (pp *INIParser) UnmarshalReader(v ifacer.Configer, in io.Reader, c map[string]interface{}) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)
	cfg := ini.Empty()
	err := cfg.Append(buf.Bytes())
	if err != nil {
		return err
	}
	sections := cfg.Sections()
	for i := 0; i < len(sections); i++ {
		section := sections[i]
		keys := section.Keys()
		for j := 0; j < len(keys); j++ {
			key := keys[j]
			value := cfg.Section(section.Name()).Key(key.Name()).String()
			c[section.Name()+"."+key.Name()] = value
		}
	}
	return nil
}
func (pp *INIParser) MarshalWriter(v ifacer.Configer, f afero.File, c map[string]interface{}) error {
	keys := v.AllKeys()
	cfg := ini.Empty()
	ini.PrettyFormat = false
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		lastSep := strings.LastIndex(key, ".")
		sectionName := key[:(lastSep)]
		keyName := key[(lastSep + 1):]
		if sectionName == "default" {
			sectionName = ""
		}
		sv, err := v.GetString(key)
		if err != nil {
			return err
		}
		cfg.Section(sectionName).Key(keyName).SetValue(sv)
	}
	cfg.WriteTo(f)
	return nil
}

// hcl parser
type HCLParser struct {
}

func (pp *HCLParser) UnmarshalReader(v ifacer.Configer, in io.Reader, c map[string]interface{}) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)

	obj, err := hcl.Parse(buf.String())
	if err != nil {
		return err
	}
	if err = hcl.DecodeObject(&c, obj); err != nil {
		return err
	}
	return nil
}
func (pp *HCLParser) MarshalWriter(v ifacer.Configer, f afero.File, c map[string]interface{}) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	ast, err := hcl.Parse(string(b))
	if err != nil {
		return err
	}
	err = printer.Fprint(f, ast.Node)
	if err != nil {
		return err
	}
	return nil
}

// dot env parser
type DOTENVParser struct {
}

func (pp *DOTENVParser) UnmarshalReader(v ifacer.Configer, in io.Reader, c map[string]interface{}) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)

	env, err := gotenv.StrictParse(buf)
	if err != nil {
		return err
	}
	for k, v := range env {
		c[k] = v
	}

	return nil
}
func (pp *DOTENVParser) MarshalWriter(v ifacer.Configer, f afero.File, c map[string]interface{}) error {
	lines := []string{}
	for _, key := range v.AllKeys() {
		envName := strings.ToUpper(strings.Replace(key, ".", "_", -1))
		val, err := v.Get(key)
		if err == nil {
			return err
		}
		lines = append(lines, fmt.Sprintf("%v=%v", envName, val))
	}
	s := strings.Join(lines, "\n")
	if _, err := f.WriteString(s); err != nil {
		return err
	}
	return nil
}

// props parser
type PROPSParser struct {
}

func (pp *PROPSParser) UnmarshalReader(vv ifacer.Configer, in io.Reader, c map[string]interface{}) error {
	v, ok := vv.(ifacer.ConfigForPropertier)
	if !ok {
		return errors.New("not support properties")
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)
	prop := properties.NewProperties()
	var err error
	if prop, err = properties.Load(buf.Bytes(), properties.UTF8); err != nil {
		return err
	}
	for _, key := range prop.Keys() {
		value, _ := prop.Get(key)
		// recursively build nested maps
		path := strings.Split(key, ".")
		lastKey := strings.ToLower(path[len(path)-1])
		deepestMap := config_tools.DeepSearch(c, path[0:len(path)-1])
		// set innermost value
		deepestMap[lastKey] = value
	}
	v.SetProperties(prop)
	return nil
}
func (pp *PROPSParser) MarshalWriter(vv ifacer.Configer, f afero.File, c map[string]interface{}) error {
	v, ok := vv.(ifacer.ConfigForPropertier)
	if !ok {
		return errors.New("not support properties")
	}
	prop := v.GetProperties()
	if prop == nil {
		prop = properties.NewProperties()
		v.SetProperties(prop)
	}
	p := prop
	for _, key := range vv.AllKeys() {
		sv, err := vv.GetString(key)
		if err != nil {
			return err
		}
		_, _, err = p.Set(key, sv)
		if err != nil {
			return err
		}
	}
	_, err := p.WriteComment(f, "#", properties.UTF8)
	if err != nil {
		return err
	}
	return nil
}
