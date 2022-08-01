package tagrunner

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

func JsonTagNameFromField(field reflect.StructField) string {
	j := strings.Split(field.Tag.Get("json"), ",")
	if len(j) == 0 || j[0] == "-" || j[0] == "" {
		return field.Name
	}

	return j[0]
}

type valAndArgs struct {
	val  string
	args []string
}

// deconstructTag compile the tag in pattern `{key}:"valA=argA1 argA2,valB=argB1 argB2"`, with returning
// []valAndArgs{
// 	  {val: "valA", args: []string{"argA1", "argA2"}}
// 	  {val: "valB", args: []string{"argB1", "argB2"}}
// }
func deconstructTag(tag string) ([]valAndArgs, error) {
	result := make([]valAndArgs, 0)

	for _, sec := range strings.Split(tag, ",") {
		if sec == "" {
			continue
		}

		vas := strings.Split(sec, "=")
		if len(vas) > 2 {
			return nil, errors.Wrap(ErrInvalidTagPattern, "sections more than 2")
		}

		val := vas[0]
		if val == "" {
			continue
		}

		args := make([]string, 0)
		if len(vas) == 2 {
			for _, a := range strings.Split(vas[1], " ") {
				if a == "" {
					continue
				}

				args = append(args, a)
			}
		}

		result = append(result, valAndArgs{
			val:  val,
			args: args,
		})
	}

	return result, nil
}
