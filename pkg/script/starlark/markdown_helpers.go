package starlark

import (
	"fmt"

	starlarkgo "go.starlark.net/starlark"
)

func titleBuiltin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	doc, err := documentArg("title", args, kwargs)
	if err != nil {
		return nil, err
	}

	return starlarkgo.String(doc.doc.MetadataString("title")), nil
}

func descriptionBuiltin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	doc, err := documentArg("description", args, kwargs)
	if err != nil {
		return nil, err
	}

	return starlarkgo.String(doc.doc.MetadataString("description")), nil
}

func h1Builtin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	return headingList("h1", args, kwargs)
}

func h2Builtin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	return headingList("h2", args, kwargs)
}

func h3Builtin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	return headingList("h3", args, kwargs)
}

func h4Builtin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	return headingList("h4", args, kwargs)
}

func h5Builtin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	return headingList("h5", args, kwargs)
}

func h6Builtin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	return headingList("h6", args, kwargs)
}

func headingsBuiltin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	doc, err := documentArg("headings", args, kwargs)
	if err != nil {
		return nil, err
	}

	return fromGoValue(doc.doc.MetadataValue("headings")), nil
}

func headingList(
	name string,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	doc, err := documentArg(name, args, kwargs)
	if err != nil {
		return nil, err
	}

	return fromGoValue(doc.doc.MetadataValue(name)), nil
}

func documentArg(
	name string,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (documentValue, error) {
	var value starlarkgo.Value
	if err := starlarkgo.UnpackArgs(name, args, kwargs, "doc", &value); err != nil {
		return documentValue{}, err
	}

	doc, ok := value.(documentValue)
	if !ok {
		return documentValue{}, fmt.Errorf("%s: want document, got %s", name, value.Type())
	}

	return doc, nil
}
