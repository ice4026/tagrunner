package tagrunner

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
)

func NewRunner(opts ...Opt) *Runner {
	r := &Runner{}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// RunFn the handler for the registered key and value. If returns skip == true, then skip the following values. Only goes into effect when err is nil.
type RunFn func(ctx context.Context, info *RunnerInfo) (skip bool, err error)
type LabelFn func(*TagField) string

type Runner struct {
	dive bool // if true, runner would traverse all elements in an array

	fns     map[string]map[string]RunFn
	labelFn LabelFn
}

func (runner *Runner) MustRegister(key string, value string, fn RunFn) {
	if err := runner.Register(key, value, fn); err != nil {
		panic(err)
	}
}

// Register the handler in key and value. Not concurrent safe.
func (runner *Runner) Register(key string, value string, fn RunFn) error {
	if runner == nil {
		return errors.Wrapf(ErrRunnerNil, "register fail, key is %s, value is %s", key, value)
	}

	if runner.fns == nil {
		runner.fns = make(map[string]map[string]RunFn)
	}

	if runner.fns[key] == nil {
		runner.fns[key] = make(map[string]RunFn)
	}

	runner.fns[key][value] = fn

	return nil
}

// SetLabelFn get label name. Mainly used in error message, with priority: label fn > json name > field name.
func (runner *Runner) SetLabelFn(fn LabelFn) {
	runner.labelFn = fn
}

func (runner *Runner) Run(ctx context.Context, source interface{}) error {
	return runner.runAny(ctx, nil, reflect.StructField{}, reflect.ValueOf(source))
}

func (runner *Runner) RunStruct(ctx context.Context, source interface{}) error {
	return runner.runStruct(
		ctx,
		newTagField(
			runner,
			getVal(reflect.ValueOf(source)),
			reflect.StructField{},
			nil,
		),
		reflect.ValueOf(source),
	)
}

func (runner *Runner) runAny(ctx context.Context, parent *TagField, field reflect.StructField, fieldValue reflect.Value) error {
	// if fieldValue is nil, then return.
	if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
		return nil
	}

	switch vi := getVal(fieldValue); vi.Kind() {
	case reflect.Struct:
		if err := runner.runStruct(
			ctx,
			newTagField(runner, vi, field, parent),
			vi); err != nil {
			return err
		}
	case reflect.Slice, reflect.Array:
		if err := runner.runSlice(ctx, vi); err != nil {
			return err
		}
	}

	return nil
}

func (runner *Runner) runStruct(ctx context.Context, parent *TagField, fieldValue reflect.Value) error {
	if runner == nil {
		return errors.Wrap(ErrRunnerNil, "run struct failed")
	}

	val := getVal(fieldValue)
	if val.Kind() != reflect.Struct {
		return errors.Wrapf(ErrInvalidInput, "expected type is struct, which is %v", val.Kind())
	}

	for i := 0; i < val.NumField(); i++ {
		v := val.Field(i)

		f := val.Type().Field(i)
		if err := runner.runField(ctx, parent, f, v); err != nil {
			return err
		}

		if err := runner.runAny(ctx, parent, f, v); err != nil {
			return err
		}
	}

	return nil
}

func (runner *Runner) runSlice(ctx context.Context, fieldValue reflect.Value) error {
	if !runner.dive {
		return nil
	}
	val := getVal(fieldValue)

	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return errors.Wrapf(ErrInvalidInput, "expected type is array or slice, which is %v", val.Kind())
	}

	for i := 0; i < val.Len(); i++ {
		if err := runner.runAny(
			ctx,
			nil,
			reflect.StructField{},
			val.Index(i),
		); err != nil {
			return err
		}
	}

	return nil
}

func (runner *Runner) runField(ctx context.Context, parent *TagField, structField reflect.StructField, fieldValue reflect.Value) error {
	if runner == nil {
		return errors.Wrap(ErrRunnerNil, "run field failed")
	}

	for key := range runner.fns {
		if err := runner.runTag(ctx, parent, structField, fieldValue, key); err != nil {
			return err
		}
	}

	return nil
}

func (runner *Runner) runTag(
	ctx context.Context,
	parent *TagField,
	structField reflect.StructField,
	fieldValue reflect.Value,
	tagKey string,
) error {
	tag, ok := structField.Tag.Lookup(tagKey)
	if !ok {
		return nil
	}

	vas, err := deconstructTag(tag)
	if err != nil {
		return err
	}

	runnerInfo := &RunnerInfo{
		Key:   tagKey,
		Field: newTagField(runner, fieldValue, structField, parent),
	}

	for _, p := range vas {
		runnerInfo.Value = p.val
		runnerInfo.Args = p.args

		if fns := runner.fns; fns == nil || fns[tagKey] == nil || fns[tagKey][runnerInfo.Value] == nil {
			continue
		}

		skip, err := runner.fns[tagKey][runnerInfo.Value](ctx, runnerInfo)
		if err != nil {
			return runnerInfo.wrapError(err)
		}

		// if skip is true, then skip the followings.
		if skip {
			return nil
		}
	}

	return nil
}

func getVal(val reflect.Value) reflect.Value {
	if val.Kind() == reflect.Ptr {
		return getVal(val.Elem())
	}

	return val
}
