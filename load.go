package sqlconditions

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// FromYAML loads a Config and parses it
func (c *Config) FromYAML(bb []byte) error {
	if err := yaml.Unmarshal(bb, &c); err != nil {
		return err
	}

	if err := c.Parse(); err != nil {
		return err
	}

	return nil
}

// Parse may be called directly without load when the Config was declared directly in Go
func (c *Config) Parse() error {
	for opName, opConfig := range c.Operations {
		for variantName, customQuery := range opConfig.VariantsByTag {
			var err error
			if customQuery == nil {
				return fmt.Errorf("Invalid (nil) customQuery def for SQL OpName:%v variant:%v", opName, variantName) // customQuery.Tokens)
			}
			if customQuery.Tokens == nil {
				return fmt.Errorf("Invalid (nil) tokens for SQL OpName:%v variant:%v", opName, variantName) // customQuery.Tokens)
			}
			customQuery.CondExpr, err = customQuery.Tokens.Parse()
			if err != nil {
				return fmt.Errorf("Parse err: %v for expr: %v", err, customQuery.Tokens)
			}
		}
	}
	return nil
}

// Strings renders a Config as a string
func (c *Config) String() string {
	var sb strings.Builder

	sb.WriteString("Operations:\n")
	for opName, opConfig := range c.Operations {
		sb.WriteString(fmt.Sprintf("\n %v:\n", opName))
		for tags, customQuery := range opConfig.VariantsByTag {
			sb.WriteString(fmt.Sprintf("   tags : %v\n", tags))
			sb.WriteString(fmt.Sprintf("   cond : %v\n\n", customQuery.CondExpr))
		}
	}

	return sb.String()
}

// Print prints a Config to standard output
func (c *Config) Print() {
	fmt.Println(c.String())
}
