package sqlconditions

import (
	"fmt"
	"os"

	//"strings"
	"testing"
)

type TestCase struct {
	ArgNames    []string
	ExpectedSQL string
}

type TestFile struct {
	FileName string

	TestCases []TestCase
}

func ArgNames(names ...string) []string {
	return names
}

var (
	TestDataDir = "./testdata"

	TestFiles = []TestFile{
		TestFile{
			"basic_00.yaml",

			[]TestCase{
				{
					ArgNames("clientID"),
					"(client_id = @clientID)",
				},
				{
					ArgNames("orgID"),
					"(consumer_org_id = @orgID)",
				},
				{
					ArgNames("clientID"),
					"(client_id = @clientID)",
				},
				{ArgNames("userID"),
					"(consumer_id = @userID)",
				},
			},
		},

		TestFile{
			"basic_01.yaml",

			[]TestCase{
				{
					ArgNames("orgID"),
					"(consumer_org_id = @orgID) AND (r.audiences = 'student' OR (r.audiences LIKE '%teacher%' AND r.audiences LIKE '%student%'))",
				},
			},
		},
	}
)

func ConfigFromFile(fileName string) (Config, error) {
	var c Config

	filePath := fmt.Sprintf("%v/%v", TestDataDir, fileName)
	bb, err := os.ReadFile(filePath)
	if err != nil {
		return c, fmt.Errorf("read test file err: %v", err)
	}

	if err := c.FromYAML(bb); err != nil {
		return c, fmt.Errorf("failed to parse yaml with err: %v", err)
	}

	if err := c.Parse(); err != nil {
		return c, fmt.Errorf("failed to parse config with err: %v", err)
	}

	return c, nil
}

func TestRunner(t *testing.T) {
	for _, testFile := range TestFiles {
		testFileName := testFile.FileName
		c, err := ConfigFromFile(testFileName)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		for _, tc := range testFile.TestCases {
			//fmt.Printf("config: %v", c)

			args := FilterArgs{}
			for _, argName := range tc.ArgNames {
				args[argName] = 123456
			}

			opName := "get-resources"
			exp, err := c.GetOperation(opName, []string{"default"})
			if err != nil {
				t.Errorf("GetOperation <%v> err: %v", opName, err)
				t.FailNow()
			}

			gotSQL, err := ToSQL(exp.CondExpr, args)
			if err != nil {
				t.Errorf("ToSQL err: %v", err)
				t.FailNow()
			}

			if gotSQL != tc.ExpectedSQL {
				t.Errorf("Expected: %v\nGot     : %v", tc.ExpectedSQL, gotSQL)
			}
		}
	}
}
