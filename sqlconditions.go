package sqlconditions

import (
	"fmt"
	"strings"
)

// Config is used as the central container holding all SQL operations with their conditions.
// It can be loaded from a YAML file.
type Config struct {
	Version    uint
	Operations map[string]OperationConfig `yaml:"operations"`
}

func StrSliceContains(ss []string, searched string) bool {
	for _, s := range ss {
		if s == searched {
			return true
		}
	}

	return false
}

func (c Config) GetCondition(opName string, searchedTagNames []string) (Condition, error) {
	var opParam Condition

	opConfig, ok := c.Operations[opName]
	if !ok {
		return opParam, ErrNotFound
	}

	if searchedTagNames == nil {
		searchedTagNames = []string{"default"}
	}

	for opTagsSpaceSeparated, op := range opConfig.VariantsByTag {
		// TODO design question: space separated or comma separated with space allowed before after comma ?
		opTags := strings.Split(opTagsSpaceSeparated, " ")

		for _, searchedTag := range searchedTagNames {
			// match, _ := HasRole(ctx, RoleName(role)); match {
			if StrSliceContains(opTags, searchedTag) {
				return *op, nil
			}
		}
	}

	return opParam, ErrNotFound
}

// OperationConfig holds one or many Condition
type OperationConfig struct {
	VariantsByTag map[string]*Condition `yaml:"variants"`
}

// Condition holds the parameters of a condition
type Condition struct {
	Joins []string

	Tokens `yaml:"condition"`

	Expr `yaml:"ignore"`
}

// FilterArgs is used to pass the set of available named argument to ToSQL methods
// so that a Condition can be rendered with or without certains subconditions
type FilterArgs map[string]any

// An ExprElt is a constituent of an Expr
type ExprElt interface {
	ToSQL(h ParseHint, argsMap FilterArgs) (string, error)
}

// Expr is a list of ExprElt : either a BinaryCondition or another sub Expr.
type Expr []ExprElt

func (se Expr) String() string {
	s := ""
	for _, exprElt := range se {
		s += fmt.Sprintf("%v\n", exprElt)
	}
	return s
}

func ToSQL(op Condition, args FilterArgs) (string, error) {
	h := ParseHint{}
	return op.Expr.ToSQL(h, args)
}

type ParseHint struct {
	ExprNotEmpty bool
}

// ToSQL renders a Expr to an SQL string
func (se Expr) ToSQL(h ParseHint, argsMap FilterArgs) (string, error) {

	sql := ""

	for idx, exprElt := range se {
		if idx == 0 {
			h.ExprNotEmpty = false
		}

		s, err := exprElt.ToSQL(h, argsMap)
		if err != nil {
			return "", err
		}

		if s != "" {
			sql += s

			if _, ok := exprElt.(BinaryCondition); ok {
				h.ExprNotEmpty = true
			}
		}
	}
	return sql, nil
}

// BinaryCondition holds all the parameters of an SQL conditional expression and can be rendered as an SQL string
// Current implementation expects ArgName to be a named argument prefixed with a @ character.
type BinaryCondition struct {
	// Modality describes whether the condition is optional or mandatory
	// If set to "if_present" the condition is optional : it will be rendered only if a named argument
	// having the same name as ArgName is found at expression evaluation time
	// Else the condition is always rendered.
	Modality string

	// LinkOperator describes how the condition should be chained
	// Currently only AND or OR operators are supported
	LinkOperator string

	// ColumnName is the name of column tested in the condition
	ColumnName string

	// Operator used to render the condition (AND, OR, LIKE expr).
	// Current parser implementation expects the operator to not contain any space characters.
	Operator string

	// ArgName is the expected named argument
	ArgName string
}

func (c BinaryCondition) ToSQL(h ParseHint, argsMap FilterArgs) (string, error) {
	s := ""

	if argsMap == nil {
		return "", fmt.Errorf("invalid (nil) param")
	}

	if len(c.ArgName) < 2 {
		return "", fmt.Errorf("BinaryCondition ArgName invalid (too short): %v", c.ArgName)
	}

	if c.Modality == "if_present" {
		// remove @ prefix from argName
		expectedArgName := c.ArgName[1:]

		if _, found := argsMap[expectedArgName]; !found {
			return "", nil
		}
	}

	if c.LinkOperator != "" && h.ExprNotEmpty {
		s += " " + c.LinkOperator + " "
	}
	return s + fmt.Sprintf("%v %v %v", c.ColumnName, c.Operator, c.ArgName), nil
}

func (c BinaryCondition) String() string {
	return fmt.Sprintf("%v %v %v %v", c.Modality, c.ColumnName, c.Operator, c.ArgName)
}
