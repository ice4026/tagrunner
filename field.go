package tagrunner

import (
	"reflect"
)

// RunnerInfo supplies all the values needed to handle current value.
// tag pattern must be `{key}:"{value}={arg1} {arg2} ... {argN}"`
type RunnerInfo struct {
	Key   string    // key of current tag
	Value string    // value of current tag
	Args  []string  // args of current tag
	Field *TagField // field info of current tag

	passBy interface{}
	runner *Runner
}

func (info *RunnerInfo) PassBy() interface{} {
	if info == nil {
		return nil
	}

	return info.passBy
}

func (info *RunnerInfo) SetPassBy(passBy interface{}) {
	if info == nil {
		return
	}

	info.passBy = passBy
}

func (info *RunnerInfo) wrapError(rawErr error) error {
	return &RunnerError{
		TagKey:          info.Key,
		TagValue:        info.Value,
		TagArgs:         info.Args,
		FieldStructName: info.Field.FieldName(),
		FieldLabelName:  info.Field.LabelName(),
		PassBy:          info.PassBy(),
		raw:             rawErr,
	}
}

func newTagField(runner *Runner, fieldValue reflect.Value, structField reflect.StructField, parent *TagField) *TagField {
	return &TagField{
		FieldValue:  fieldValue,
		StructField: structField,
		Parent:      parent,
		labelFn:     runner.labelFn,
	}
}

// TagField contains field info of current field
type TagField struct {
	FieldValue  reflect.Value       // reflect.Value of current field
	StructField reflect.StructField // reflect.StructField of current field
	Parent      *TagField           // parent of current field

	labelFn LabelFn // function for getting current label name
}

func (f *TagField) IsTop() bool {
	return f == nil || f.Parent == nil
}

func (f *TagField) LabelName() string {
	if f.labelFn == nil {
		return f.jsonName()
	}

	return f.labelFn(f)
}

func (f *TagField) FieldName() string {
	return f.StructField.Name
}

// jsonName get json name for current field
func (f *TagField) jsonName() string {
	return JsonTagNameFromField(f.StructField)
}
