package dsl

type Variable struct {
	Name string `yaml:"name" json:"name"`
	Type string `yaml:"type" json:"type"`
}

type TestVeriable struct {
	Name  string      `yaml:"name" json:"name"`
	Value interface{} `yaml:"value" json:"value"`
}

type TestCase struct {
	Name      string         `yaml:"name" json:"name"`
	Variables []TestVeriable `yaml:"variables" json:"variables"`
	Expected  bool           `yaml:"expected" json:"expected"`
}

type Validation struct {
	ID        string     `yaml:"id" json:"id"`
	Cels      []string   `yaml:"cels" json:"cels"`
	Variables []Variable `yaml:"variables" json:"variables"`
	TestCases []TestCase `yaml:"testCases" json:"testCases"`
}

type DSL struct {
	Validations []Validation `yaml:"validations" json:"validations"`
}
