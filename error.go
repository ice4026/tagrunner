package tagrunner

import (
	"errors"
	"fmt"

	jsoniter "github.com/json-iterator/go"
)

var (
	ErrRunnerNil         = errors.New("runner is nil")
	ErrInvalidInput      = errors.New("invalid input")
	ErrInvalidTagPattern = errors.New("invalid tag pattern")
)

type RunnerError struct {
	TagKey          string      `json:"tag_key"`
	TagValue        string      `json:"tag_value"`
	TagArgs         []string    `json:"tag_args"`
	FieldStructName string      `json:"field_struct_name"`
	FieldLabelName  string      `json:"field_label_name"`
	PassBy          interface{} `json:"pass_by"`

	raw error
}

func (r RunnerError) Error() string {
	return fmt.Sprintf("%s\n%s", r.raw.Error(), marshal(r))
}

func (r *RunnerError) Unwrap() error {
	if r == nil {
		return nil
	}

	return r.raw
}

func marshal(o interface{}) string {
	s, _ := jsoniter.MarshalToString(o)

	return s
}
