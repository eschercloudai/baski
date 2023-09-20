package data

import (
	"errors"
	"testing"
)

// TestGetNestedField enables the ability to step through a map[string]interface{} to get sub keys
func TestGetNestedField(t *testing.T) {
	testMap := map[string]interface{}{
		"this": map[string]interface{}{
			"is": map[string]interface{}{
				"a": map[string]interface{}{
					"test": map[string]interface{}{
						"item": "result1",
					},
				},
				"another": map[string]interface{}{
					"test": map[string]string{
						"item": "result2",
					},
				},
			},
		},
	}
	tests := []struct {
		Name     string
		Keys     []string
		Expected string
		Error    error
	}{
		{
			Name:     "Test map[string]interface",
			Keys:     []string{"this", "is", "a", "test", "item"},
			Expected: "result1",
			Error:    nil,
		},
		{
			Name:     "Test map[string]string",
			Keys:     []string{"this", "is", "another", "test", "item"},
			Expected: "",
			Error:    errors.New("key item not found or not a map\n"),
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			res, err := GetNestedField(testMap, test.Keys...)

			if err != nil {
				if test.Error != nil {
					if test.Error.Error() != err.Error() {
						t.Errorf("expected error \"%s\" got \"%s\"\n", test.Error.Error(), err.Error())
					}
					return
				} else {
					t.Error(err)
					return
				}
			}

			if res != test.Expected {
				t.Errorf("expected %s got %s\n", test.Expected, res)
			}
		})
	}

}
