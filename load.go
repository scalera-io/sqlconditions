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
		for variantName, cond := range opConfig.VariantsByTag {
			var err error
			if cond == nil {
				return fmt.Errorf("Invalid (nil) cond def for SQL OpName:%v variant:%v", opName, variantName) // cond.Tokens)
			}
			if cond.Tokens == nil {
				return fmt.Errorf("Invalid (nil) tokens for SQL OpName:%v variant:%v", opName, variantName) // cond.Tokens)
			}
			cond.Expr, err = cond.Tokens.Parse()
			if err != nil {
				return fmt.Errorf("Parse err: %v for expr: %v", err, cond.Tokens)
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
		for tags, cond := range opConfig.VariantsByTag {
			sb.WriteString(fmt.Sprintf("   tags : %v\n", tags))
			sb.WriteString(fmt.Sprintf("   cond : %v\n\n", cond.Expr))
		}
	}

	return sb.String()
}

// Print prints a Config to standard output
func (c *Config) Print() {
	fmt.Println(c.String())
}
