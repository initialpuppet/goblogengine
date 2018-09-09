package envae

import (
	"strconv"
	"strings"
	"testing"
)

var testVals = map[string]string{
	"string_config_val":        "string conf val",
	"int_config_val":           "10",
	"string_slice_config_val":  "foo,bar",
	"nested_string_config_val": "nested string conf val",
}

func mockLookupEnv(key string) (string, bool) {
	v, ok := testVals[key]
	return v, ok
}

type correctConfig struct {
	TestString      string   `envae:"string_config_val"`
	TestInt         int      `envae:"int_config_val"`
	TestStringSlice []string `envae:"string_slice_config_val"`
	Nest            moreCorrectConfig
}

type moreCorrectConfig struct {
	TestNestedString string `envae:"nested_string_config_val"`
}

func TestPopulate(t *testing.T) {
	lookupEnv = mockLookupEnv
	conf := &correctConfig{}
	err := Populate(conf)

	expInt, _ := strconv.Atoi(testVals["int_config_val"])
	if conf.TestInt != expInt {
		t.Errorf("Configuration integer values don't match, got %d need %d", conf.TestInt, expInt)
	}

	expString := testVals["string_config_val"]
	if strings.Compare(conf.TestString, expString) != 0 {
		t.Errorf("Configuration string values don't match, got %s need %s", conf.TestString, expString)
	}

	if strings.Compare(conf.TestStringSlice[1], "bar") != 0 {
		t.Errorf("Configuration string slice values don't match, got %s need bar", conf.TestStringSlice[1])
	}

	if err != nil {
		t.Errorf("Error populating configuration struct: %s", err)
	}

}

func TestNonPointerValuesShouldBeRejected(t *testing.T) {
	conf := correctConfig{}
	err := Populate(conf)
	if err == nil {
		t.Error("Populate function accepted struct value, should only accept struct pointers.")
	}
}

func TestNonStructShouldBeRejected(t *testing.T) {
	conf := "foo"
	err := Populate(conf)
	if err == nil {
		t.Error("Populate function accepted string, should only accept struct pointers.")
	}
}

// type configWithStructPointer struct {
// 	o          *configPointed
// 	TestString string `envae:"string_config_val"`
// }

// type configPointed struct {
// 	o                 configWithStructPointer
// 	TestPointedString string `envae:"nested_string_config_val"`
// }

// func TestInnerStructsCannotBePointers(t *testing.T) {
// 	lookupEnv = mockLookupEnv
// 	conf := &configWithStructPointer{}
// 	err := Populate(conf)
// 	if err == nil {
// 		t.Error("Accepted struct containing a pointer to another struct")
// 	}
// 	log.Print(err)
// }
