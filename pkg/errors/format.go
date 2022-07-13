package errors

import (
	"fmt"
	"reflect"
	"strings"
	"text/template"
)

var (
	// FormatWithoutTrace is a template format for formatting errors without a
	// stack trace.
	//
	// Format:
	//   <error1>: <error2>: <errorN>
	FormatWithoutTrace = `{{ if ge (len .) 1 }}{{ (index . 0).Msg }}{{ range (slice . 1) }}: {{ .Msg }}{{ end }}{{ end }}`

	// FormatWithTrace is a template format for formatting errors with a partial
	// stack trace for each wrapped error.
	//
	// Format:
	//   <error1>:
	//       <file1>:<line1> in <function1>
	//       <file2>:<line2> in <function2>
	//       <fileN>:<lineN> in <functionN>
	//   <error2>:
	//       <file1>:<line1> in <function1>
	//       <file2>:<line2> in <function2>
	//       <fileN>:<lineN> in <functionN>
	//   <errorN>:
	//       <file1>:<line1> in <function1>
	//       <file2>:<line2> in <function2>
	//       <fileN>:<lineN> in <functionN>
	FormatWithTrace = `{{ range . }}{{ .Msg }}{{ if .Stack }}{{ range .Stack }}
    {{ .File }}:{{ .Line }} in {{ .Name }}{{ end }}{{ end }}
{{end}}`

	// FormatWithCombinedTrace is a template format for formatting errors with
	// a single combined stack trace.
	//
	// Format:
	//   <error1>: <error2>: <errorN>
	//       <file1>:<line1> in <function1>
	//       <file2>:<line2> in <function2>
	//       <fileN>:<lineN> in <functionN>
	FormatWithCombinedTrace = `{{ if ge (len .) 1 }}{{ (index . 0).Msg }}{{ range (slice . 1) }}: {{ .Msg }}{{ end }}
{{ with (index . 0) }}{{ if gt (len .Stack) 0 }}{{ range .Err.FullStack }}    {{ .File }}:{{ .Line }} in {{ .Name }}
{{ end }}{{ end }}{{ end }}{{ end }}`

	// FormatPythonTraceback is a template format for formatting errors similar
	// to Python's tracebacks.
	//
	// Format:
	//   Traceback (most recent call last):
	//     File "<file3>", line <line3>, in <function3>
	//     File "<file2>", line <line2>, in <function2>
	//     File "<file1>", line <line1>, in <function1>
	//   <typeName1>: <error1>
	//
	//   The above exception was the direct cause of the following exception:
	//
	//   Traceback (most recent call last):
	//     File "<file3>", line <line3>, in <function3>
	//     File "<file2>", line <line2>, in <function2>
	//     File "<file1>", line <line1>, in <function1>
	//   <typeName2>: <error2>
	FormatPythonTraceback = `{{ $l := len . }}{{ range $i, $e := (reversed .) }}{{ $stack := reversed .PartialStack }}
{{ if gt (len $stack) 0 }}Traceback (most recent call last):{{ range $stack }}
  File "{{ .File }}", line {{ .Line }}, in {{ .Name }}{{ end }}{{ end }}
{{ .TypeName }}: {{ .Msg }}{{ if not (last $i $l) }}

The above exception was the direct cause of the following exception:
{{ end }}{{ end }}
`

	templateFuncs = template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
		"dec": func(i int) int {
			return i + 1
		},
		"last": func(i, length int) bool {
			return i == length-1
		},
		"reversed": func(s interface{}) []interface{} {
			sv := reflect.ValueOf(s)
			if sv.Kind() != reflect.Slice {
				return nil
			}
			// create a new reversed list
			reversed := make([]interface{}, 0, sv.Len())
			for i := sv.Len() - 1; i >= 0; i-- {
				reversed = append(reversed, sv.Index(i).Interface())
			}
			return reversed
		},
	}
)

// ToString returns a default formatted string for a given error.
func ToString(err error, withTrace bool) string {
	if withTrace {
		return ToCustomString(err, FormatWithTrace)
	}
	return ToCustomString(err, FormatWithoutTrace)
}

// ToCustomString returns a custom formatted string for a given error. The
// format is defined by the given Go template.
func ToCustomString(err error, tplStr string) string {
	tpl := template.Must(template.New("").Funcs(templateFuncs).Parse(tplStr))
	strBld := strings.Builder{}
	upkErr := Unpack(err, true)
	terr := tpl.Execute(&strBld, upkErr)
	if terr != nil {
		panic(terr)
	}
	return strings.TrimSpace(strBld.String())
}

// ToJSON returns a JSON formatted slice for a given error.
func ToJSON(err error, withTrace bool) []interface{} {
	upkErr := Unpack(err, true)

	jsonList := make([]interface{}, 0, len(upkErr))
	for _, errElm := range upkErr {
		jsonList = append(jsonList, errElm.formatJSON(withTrace))
	}

	return jsonList
}

// Unpack returns a human-readable UnpackedError type for a given error.
func Unpack(err error, unwrapExternal bool) UnpackedError {
	upkErr := make([]UnpackedElement, 0, 20)
	prvStack := Stack{}

	for err != nil {
		// check if it behaves like a base error
		if e, ok := err.(interface {
			Stack() Stack
			FullStack() Stack
		}); ok {
			stack := e.Stack()
			upkErr = append(upkErr, UnpackedElement{
				Err:          err,
				Msg:          err.Error(),
				Stack:        stack,
				PartialStack: stack.RelativeTo(prvStack),
				FullStack:    e.FullStack(),
				TypeName:     TypeName(err),
			})
			prvStack = stack

		} else {
			// continue unwrapping external errors
			upkErr = append(upkErr, UnpackedElement{
				Err:      err,
				Msg:      err.Error(),
				TypeName: TypeName(err),
			})
			if !unwrapExternal {
				break
			}
		}

		err = Unwrap(err)
	}

	return upkErr
}

// UnpackedError represents an unpacked error which is quite useful for
// formatting purposes and other error processing. Use Unpack() to unpack any
// kind of error that supports it.
type UnpackedError []UnpackedElement

// String formatter for external errors.
func formatExternalStr(err error, withTrace bool) string {
	if withTrace {
		return fmt.Sprintf("%+v", err)
	}
	return fmt.Sprint(err)
}

// UnpackedElement represents a single error frame and the accompanying message.
type UnpackedElement struct {
	// Err is the error instance.
	Err error
	// Msg is the message contained in the error.
	Msg string
	// TypeName is the type name of the error.
	TypeName string
	// Stack is the error stack for this particular error instance.
	Stack Stack
	// PartialStack is the error stack with parts cut off that are already in
	// the previous error stack.
	PartialStack Stack
	// FullStack is the combined error stack of all errors in err's chain.
	FullStack Stack
}

// JSON formatter for wrap error chains.
func (ue *UnpackedElement) formatJSON(withTrace bool) map[string]interface{} {
	wrapMap := make(map[string]interface{})
	wrapMap["message"] = fmt.Sprint(ue.Msg)
	stackList := make([]interface{}, 0, len(ue.PartialStack))

	if withTrace {
		for _, frame := range ue.PartialStack {
			stackMap := make(map[string]interface{})
			stackMap["function"] = frame.Name
			stackMap["file"] = frame.File
			stackMap["line"] = frame.Line
			stackMap["location"] = fmt.Sprintf("%s:%s:%d", frame.Name, frame.File, frame.Line)
			stackList = append(stackList, stackMap)
		}
		wrapMap["stack"] = stackList
	}

	return wrapMap
}
