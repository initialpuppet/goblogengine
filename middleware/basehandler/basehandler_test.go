package basehandler_test

import (
	"errors"
	"reflect"
	"testing"

	"goblogengine/middleware/basehandler"
)

func TestAppErrorf(t *testing.T) {
	terr := errors.New("Test error")
	tmsg := "Test message"
	tcode := 500

	need := &basehandler.AppError{
		Error:      terr,
		Message:    tmsg,
		StatusCode: tcode,
	}

	have := basehandler.AppErrorf(tmsg, tcode, terr)

	if !reflect.DeepEqual(need, have) {
		t.Errorf("AppError struct incorrectly generated, need %v have %v", need, have)
	}
}
