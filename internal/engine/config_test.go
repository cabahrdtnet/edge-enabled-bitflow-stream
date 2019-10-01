package engine

import (
	"testing"
)

func TestReplaceIO_EmptyScript_ShouldReturnEmptyStringAndError(t *testing.T) {
	// arrange
	str := ""

	// act
	result, err := ReplaceIO(str)

	// assert
	expected := ""
	if result != expected || err == nil {
		t.Errorf("Can't handle empty script!\n-Result was: %s\n-Expected was: %s", result, expected)
	}
}

func TestReplaceIO_ErrorInValuesToReplace_ShouldReturnFaultyScriptAndError(t *testing.T) {
	// arrange
	str := "iput -> avg() -> outpit"

	// act
	result, err := ReplaceIO(str)

	// assert
	expected := "iput -> avg() -> outpit"
	if result != expected || err == nil {
		t.Errorf("Can't handle faulty script!\n-Result was: %s\n-Expected was: %s", result, expected)
	}
}

func TestReplaceIO_MinimalScript_ShouldSucceed(t *testing.T) {
	// arrange
	str := "input -> output"

	// act
	result, err := ReplaceIO(str)

	// assert
	expected := "std://- -> std://-"
	if result != expected || err != nil {
		t.Errorf("Can't handle minimal script!\n-Result was: %s\n-Expected was: %s", result, expected)
	}
}

func TestReplaceIO_MinimalScriptWithWhiteSpaces_ShouldSucceed(t *testing.T) {
	// arrange
	str := `	
			    input -> output	
	              
			`

	// act
	result, err := ReplaceIO(str)

	// assert
	expected := "std://- -> std://-"
	if result != expected || err != nil {
		t.Errorf("Can't handle whitespaced script!\n-Result was: %s\n-Expected was: %s", result, expected)
	}
}


func TestReplaceIO_AverageScript_ShouldSucceed(t *testing.T) {
	// arrange
	str := `input -> avg() -> do(expr='set("input1",output * 10, set("field2", now(), set_tag("input", "output")))) -> output`

	// act
	result, err := ReplaceIO(str)

	// assert
	expected := `std://- -> avg() -> do(expr='set("input1",output * 10, set("field2", now(), set_tag("input", "output")))) -> std://-`
	if result != expected || err != nil {
		t.Errorf("Can't handle average script!\n-Result was: %s\n-Expected was: %s", result, expected)
	}
}