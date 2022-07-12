package errors

import (
	"fmt"
	"strings"
	"text/template"
)

// ToString returns a default formatted string for a given error.
func ToString(err error, withTrace bool) string {
	if withTrace {
		return ToCustomString(err, DefaultFormatWithTrace)
	}
	return ToCustomString(err, DefaultFormat)
}

// DefaultFormat is the default template for formatting errors.
var DefaultFormat = `{{if ge (len .) 1 }}{{(index . 0).Msg}}{{range (slice . 1) }}: {{.Msg}}{{ end }}{{ end }}`

// DefaultFormatWithTrace is the default template for formatting errors that include stack traces.
var DefaultFormatWithTrace = `{{range .}}{{.Msg}}{{if .Stack}}{{range .Stack}}
    {{.File}}:{{.Line}} in {{.Name}}{{end}}{{end}}
{{end}}`

// ToCustomString returns a custom formatted string for a given error. The
// format is defined by the given template.
func ToCustomString(err error, tplStr string) string {
	tpl := template.Must(template.New("").Parse(tplStr))
	strBld := strings.Builder{}
	upkErr := Unpack(err, true)
	terr := tpl.Execute(&strBld, upkErr)
	if terr != nil {
		panic(err)
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
			Message() string
		}); ok {
			fullStack := e.Stack()
			upkErr = append(upkErr, UnpackedElement{
				Err:       err,
				Msg:       e.Message(),
				Stack:     relativeStack(prvStack, fullStack),
				FullStack: fullStack,
			})
			prvStack = fullStack

		} else {
			// continue unwrapping external errors
			upkErr = append(upkErr, UnpackedElement{
				Err: err,
				Msg: err.Error(),
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
	Err       error
	Msg       string
	Stack     Stack
	FullStack Stack
}

// JSON formatter for wrap error chains.
func (ue *UnpackedElement) formatJSON(withTrace bool) map[string]interface{} {
	wrapMap := make(map[string]interface{})
	wrapMap["message"] = fmt.Sprint(ue.Msg)
	stackList := make([]interface{}, 0, len(ue.Stack))

	if withTrace {
		for _, frame := range ue.Stack {
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
