package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"regexp"
)

type Config struct {
	Apps []*AppConfig

	original string
}

type AppConfig struct {
	Name   string `yaml:"name"`
	Format string `yaml:"format"`

	SourceFiles      []string          `yaml:"source_files"`
	StaticLabels     map[string]string `yaml:"static_labels"`
	RelabelConfig    *RelabelConfig    `yaml:"relabel_config"`
	HistogramBuckets []float64         `yaml:"histogram_buckets"`
}

type RelabelConfig struct {
	SourceLabels []string                `yaml:"source_labels"`
	Replacement  map[string]*Replacement `yaml:"replacement"`
}

type Replacement struct {
	Trim     string     `yaml:"trim"`
	Replaces []*Replace `yaml:"replaces"`
}

type Replace struct {
	Target string `yaml:"target"`
	Value  string `yaml:"value"`

	Rex *regexp.Regexp
}

// 加载并解析配置文件
func LoadFile(filename string) (*Config, error) {
	var apps []*AppConfig
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	s := string(content)
	err = yaml.Unmarshal([]byte(s), &apps)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		original: s,
		Apps:     apps,
	}

	return cfg, nil
}

// 返回静态label及其value
func (cfg *AppConfig) StaticLabelSets() ([]string, []string) {
	labels := make([]string, len(cfg.StaticLabels))
	values := make([]string, len(cfg.StaticLabels))

	i := 0
	for k, v := range cfg.StaticLabels {
		labels[i] = k
		values[i] = v
		i++
	}

	return labels, values
}

// 返回动态label
func (cfg *AppConfig) DynamicLabels() []string {
	return cfg.RelabelConfig.SourceLabels
}

// 生成正则对象
func (cfg *AppConfig) Prepare() {
	for _, r := range cfg.RelabelConfig.Replacement {
		for _, rItem := range r.Replaces {
			rItem.prepare()
		}
	}
}

func (rp *Replace) prepare() {
	replace, err := regexp.Compile(rp.Target)
	if err != nil {
		log.Panic(err)
	}

	rp.Rex = replace
}
