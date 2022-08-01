package tagrunner

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRunnerError_Error(t *testing.T) {
	Convey("root", t, func() {
		err := &RunnerError{
			TagKey:          "key",
			TagValue:        "val",
			TagArgs:         nil,
			FieldStructName: "sname",
			FieldLabelName:  "jname",
			PassBy:          1,
			raw:             errors.New("test error"),
		}

		So(err.Error(), ShouldEqual, `test error
{"tag_key":"key","tag_value":"val","tag_args":null,"field_struct_name":"sname","field_label_name":"jname","pass_by":1}`)
	})
}

func TestRunnerError_Unwrap(t *testing.T) {
	Convey("root", t, func() {
		e := errors.New("test error")
		err := &RunnerError{
			TagKey:          "key",
			TagValue:        "val",
			TagArgs:         nil,
			FieldStructName: "sname",
			FieldLabelName:  "jname",
			PassBy:          1,
			raw:             e,
		}

		So(err.Unwrap(), ShouldEqual, e)
	})
}
