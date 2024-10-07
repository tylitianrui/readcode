# RecordRule源码解析



## 数据结构

### 规则配置相关的数据结构

**RuleGroups**



``````go
// RuleGroups is a set of rule groups that are typically exposed in a file.
type RuleGroups struct {
	Groups []RuleGroup `yaml:"groups"`
}

``````



**RuleGroup**



``````go
// RuleGroup is a list of sequentially evaluated recording and alerting rules.
type RuleGroup struct {
	Name        string          `yaml:"name"`
	Interval    model.Duration  `yaml:"interval,omitempty"`
	QueryOffset *model.Duration `yaml:"query_offset,omitempty"`
	Limit       int             `yaml:"limit,omitempty"`
	Rules       []RuleNode      `yaml:"rules"`
}
``````

**RuleNode**

``````go
// RuleNode adds yaml.v3 layer to support line and column outputs for invalid rules.
type RuleNode struct {
	Record        yaml.Node         `yaml:"record,omitempty"`
	Alert         yaml.Node         `yaml:"alert,omitempty"`
	Expr          yaml.Node         `yaml:"expr"`
	For           model.Duration    `yaml:"for,omitempty"`
	KeepFiringFor model.Duration    `yaml:"keep_firing_for,omitempty"`
	Labels        map[string]string `yaml:"labels,omitempty"`
	Annotations   map[string]string `yaml:"annotations,omitempty"`
}

``````







