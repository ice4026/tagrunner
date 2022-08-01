# TagRunner

Handle the tags in pattern `{key}:"{value}={arg1} {arg2} ... {argN}"`.

## Init

e.g.

```go
package tagrunner

type hello struct {
	World string `json:"field_hello" doctor:"hello=v1 v2"`
}

func main() {
	target := &hello{World: "hi"}

	runner := NewRunner()
	runner.MustRegister("doctor", "hello", func(ctx context.Context, info *RunnerInfo) error {
		// 对字段的处理逻辑
		return nil
	})
}
```

## RunnerInfo

Contains all things needed to handle the field.

```go
type RunnerInfo struct {
    Key   string 
    Value string 
    Args  []string  
    Field *TagField 
    
    passBy interface{} 
}

type TagField struct {
    FieldValue  reflect.Value       
    StructField reflect.StructField 
    Parent      *TagField 
    }
```

Btw, you can use `SetPassBy()` to pass something within a tag key.

## Error Handling

Got error of type `*RunnerError`, can use `Unwrap()` to get the error thrown by handler.