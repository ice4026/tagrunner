package tagrunner

import (
	"context"
	"errors"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRunner_RunStruct(t *testing.T) {
	Convey("Normal", t, func() {
		type hello struct {
			FieldHello string `json:"field_hello" doctor:"hello=world1 world2"`
			Foo        string `json:"f" bar:"update=v1 v2,validate"`
		}

		target := &hello{FieldHello: "hi", Foo: "bbb"}

		runner := NewRunner()
		labelFn := func(field *TagField) string {
			return field.FieldValue.String()
		}
		runner.SetLabelFn(labelFn)
		runner.MustRegister("doctor", "hello", func(ctx context.Context, info *RunnerInfo) (bool, error) {
			So(info, ShouldNotBeNil)
			So(info.Key, ShouldEqual, "doctor")
			So(info.Value, ShouldEqual, "hello")
			So(info.Args, ShouldResemble, []string{"world1", "world2"})
			So(info.PassBy(), ShouldBeNil)
			So(info.Field, ShouldNotBeNil)
			So(info.Field.jsonName(), ShouldEqual, "field_hello")
			So(info.Field.LabelName(), ShouldEqual, "hi")
			So(info.Field.StructField.Name, ShouldEqual, "FieldHello")
			So(info.Field.FieldValue.String(), ShouldEqual, "hi")
			So(info.Field.labelFn, ShouldEqual, labelFn)
			So(info.Field.Parent.FieldValue, ShouldResemble, reflect.Indirect(reflect.ValueOf(target)))
			So(info.Field.Parent.StructField, ShouldResemble, reflect.StructField{})
			So(info.Field.Parent.Parent, ShouldBeNil)
			So(info.Field.Parent.labelFn, ShouldEqual, labelFn)
			So(info.Field.IsTop(), ShouldBeFalse)
			So(info.Field.Parent.IsTop(), ShouldBeTrue)

			return true, nil
		})

		runner.MustRegister("bar", "update", func(ctx context.Context, info *RunnerInfo) (bool, error) {
			info.SetPassBy(1)
			info.Field.FieldValue.Set(reflect.ValueOf(info.Args[0]))

			return false, nil
		})

		runner.MustRegister("bar", "validate", func(ctx context.Context, info *RunnerInfo) (bool, error) {
			So(info.PassBy(), ShouldEqual, 1)

			return false, nil
		})

		err := runner.RunStruct(context.Background(), target)
		So(err, ShouldBeNil)
		So(target.Foo, ShouldEqual, "v1")
	})

	Convey("Recursion", t, func() {
		type hello struct {
			World struct {
				Field string `json:"field" doctor:"hello=world1 world2"`
			} `foo:"bar=arg1 arg2"`
		}

		target := &hello{
			World: struct {
				Field string `json:"field" doctor:"hello=world1 world2"`
			}{
				Field: "hi",
			}}

		runner := NewRunner()
		runner.MustRegister("doctor", "hello", func(ctx context.Context, info *RunnerInfo) (bool, error) {
			So(info, ShouldNotBeNil)
			So(info.Key, ShouldEqual, "doctor")
			So(info.Value, ShouldEqual, "hello")
			So(info.Args, ShouldResemble, []string{"world1", "world2"})
			So(info.PassBy(), ShouldBeNil)
			So(info.Field, ShouldNotBeNil)
			So(info.Field.jsonName(), ShouldEqual, "field")
			So(info.Field.StructField.Name, ShouldEqual, "Field")
			So(info.Field.FieldValue.String(), ShouldEqual, "hi")
			So(info.Field.Parent, ShouldResemble, &TagField{
				FieldValue:  reflect.ValueOf(target).Elem().Field(0),
				StructField: reflect.TypeOf(target).Elem().Field(0),
				Parent: &TagField{
					FieldValue:  reflect.Indirect(reflect.ValueOf(target)),
					StructField: reflect.StructField{},
					Parent:      nil,
				},
			})
			So(info.Field.IsTop(), ShouldBeFalse)
			So(info.Field.Parent.IsTop(), ShouldBeFalse)
			So(info.Field.Parent.Parent.IsTop(), ShouldBeTrue)

			return false, nil
		})

		runner.MustRegister("foo", "bar", func(ctx context.Context, info *RunnerInfo) (skip bool, err error) {
			So(info, ShouldNotBeNil)
			So(info.Key, ShouldEqual, "foo")
			So(info.Value, ShouldEqual, "bar")
			So(info.Args, ShouldResemble, []string{"arg1", "arg2"})
			So(info.PassBy(), ShouldBeNil)
			So(info.Field, ShouldNotBeNil)
			So(info.Field.jsonName(), ShouldEqual, "World")
			So(info.Field.StructField.Name, ShouldEqual, "World")
			So(info.Field.FieldValue.Interface(), ShouldResemble, struct {
				Field string `json:"field" doctor:"hello=world1 world2"`
			}{
				Field: "hi",
			})
			So(info.Field.Parent, ShouldResemble, &TagField{
				FieldValue:  reflect.ValueOf(target).Elem(),
				StructField: reflect.StructField{},
				Parent:      nil,
			})
			So(info.Field.IsTop(), ShouldBeFalse)
			So(info.Field.Parent.IsTop(), ShouldBeTrue)

			return false, nil
		})

		err := runner.RunStruct(context.Background(), target)
		So(err, ShouldBeNil)
	})

	Convey("Error", t, func() {
		type hello struct {
			World string `json:"world" doctor:"hello=a1 a2"`
		}

		target := &hello{World: "world"}

		runner := NewRunner()
		runner.MustRegister("doctor", "hello", func(ctx context.Context, info *RunnerInfo) (bool, error) {
			info.SetPassBy(1)
			return false, errors.New("hello error")
		})

		err := runner.RunStruct(context.Background(), target)
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, &RunnerError{
			TagKey:          "doctor",
			TagValue:        "hello",
			TagArgs:         []string{"a1", "a2"},
			FieldStructName: "World",
			FieldLabelName:  "world",
			PassBy:          1,
			raw:             errors.New("hello error"),
		})
	})

	Convey("Anonymous", t, func() {
		type world struct {
			hi string `doctor:"hello=a1 a2"`
		}

		type hello struct {
			world
		}

		target := &hello{
			world: world{hi: "what"},
		}

		runner := NewRunner()
		runner.MustRegister("doctor", "hello", func(ctx context.Context, info *RunnerInfo) (bool, error) {
			info.SetPassBy(1)

			return false, errors.New("hello error")
		})

		err := runner.RunStruct(context.Background(), target)
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, &RunnerError{
			TagKey:          "doctor",
			TagValue:        "hello",
			TagArgs:         []string{"a1", "a2"},
			FieldStructName: "hi",
			FieldLabelName:  "hi",
			PassBy:          1,
			raw:             errors.New("hello error"),
		})
	})

}

func TestDiveSlice(t *testing.T) {
	Convey("Dive slice", t, func() {
		type World struct {
			Text string `test_key:"test_value=arg1 arg2"`
		}

		type Hello struct {
			Worlds []*World
		}

		hello := &Hello{Worlds: []*World{
			{Text: "world"},
			{Text: "world"},
		}}

		runner := NewRunner(WithDive(true))
		runner.MustRegister("test_key", "test_value", func(ctx context.Context, info *RunnerInfo) (bool, error) {
			So(info, ShouldNotBeNil)
			So(info.Key, ShouldEqual, "test_key")
			So(info.Value, ShouldEqual, "test_value")
			So(info.Field.Parent.IsTop(), ShouldBeTrue)
			So(info.Args, ShouldResemble, []string{"arg1", "arg2"})
			So(info.Field.FieldValue.String(), ShouldEqual, "world")
			So(info.Field.StructField, ShouldResemble, reflect.TypeOf(World{}).Field(0))

			info.Field.FieldValue.Set(reflect.ValueOf("work"))

			return false, nil
		})

		So(runner.RunStruct(context.Background(), hello), ShouldBeNil)
		So(hello, ShouldResemble, &Hello{Worlds: []*World{
			{Text: "work"},
			{Text: "work"},
		}})
	})
}

func TestRunSlice(t *testing.T) {
	Convey("Run slice", t, func() {
		type Hello struct {
			World string `test_key:"test_value=arg1 arg2"`
		}

		hellos := []*Hello{
			{World: "world"},
			{World: "world"},
		}

		runner := NewRunner(WithDive(true))
		runner.MustRegister("test_key", "test_value", func(ctx context.Context, info *RunnerInfo) (bool, error) {
			So(info, ShouldNotBeNil)
			So(info.Key, ShouldEqual, "test_key")
			So(info.Value, ShouldEqual, "test_value")
			So(info.Field.Parent.IsTop(), ShouldBeTrue)
			So(info.Args, ShouldResemble, []string{"arg1", "arg2"})
			So(info.Field.FieldValue.String(), ShouldEqual, "world")
			So(info.Field.StructField, ShouldResemble, reflect.TypeOf(Hello{}).Field(0))

			info.Field.FieldValue.Set(reflect.ValueOf("work"))

			return false, nil
		})

		So(runner.Run(context.Background(), hellos), ShouldBeNil)
		So(hellos, ShouldResemble, []*Hello{
			{World: "work"},
			{World: "work"},
		})
	})
}
