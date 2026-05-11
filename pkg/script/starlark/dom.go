package starlark

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/kai687/mini-crawler/pkg/script"
	starlarkgo "go.starlark.net/starlark"
)

// documentValue is the Starlark-facing wrapper around a parsed document.
//
// It exposes only the DOM operations scripts need, keeping goquery internals out
// of the public script API.
type documentValue struct {
	doc script.Document
}

// newDocumentValue wraps a neutral script document for Starlark code.
func newDocumentValue(doc script.Document) starlarkgo.Value {
	return documentValue{doc: doc}
}

func (v documentValue) String() string { return fmt.Sprintf("<document %q>", v.doc.URL()) }
func (v documentValue) Type() string   { return "document" }
func (v documentValue) Freeze()        {}
func (v documentValue) Truth() starlarkgo.Bool {
	return true
}

func (v documentValue) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", v.Type())
}

func (v documentValue) AttrNames() []string {
	return []string{"select", "select_first", "url"}
}

func (v documentValue) Attr(name string) (starlarkgo.Value, error) {
	switch name {
	case "url":
		return starlarkgo.String(v.doc.URL()), nil
	case "select":
		return starlarkgo.NewBuiltin("document.select", v.selectNodes), nil
	case "select_first":
		return starlarkgo.NewBuiltin("document.select_first", v.selectFirst), nil
	default:
		return nil, nil
	}
}

func (v documentValue) selectNodes(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	var css string
	if err := starlarkgo.UnpackArgs("select", args, kwargs, "css", &css); err != nil {
		return nil, err
	}

	return nodesFromSelection(v.doc.GoqueryDocument().Find(css)), nil
}

func (v documentValue) selectFirst(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	var css string
	if err := starlarkgo.UnpackArgs("select_first", args, kwargs, "css", &css); err != nil {
		return nil, err
	}

	return firstNodeOrNone(v.doc.GoqueryDocument().FindMatcher(goquery.Single(css))), nil
}

// nodeValue is the Starlark-facing wrapper around one goquery selection.
type nodeValue struct {
	selection *goquery.Selection
}

// newNodeValue wraps a goquery selection for Starlark code.
func newNodeValue(selection *goquery.Selection) starlarkgo.Value {
	return nodeValue{selection: selection}
}

func (v nodeValue) String() string { return fmt.Sprintf("<node %s>", goquery.NodeName(v.selection)) }
func (v nodeValue) Type() string   { return "node" }
func (v nodeValue) Freeze()        {}
func (v nodeValue) Truth() starlarkgo.Bool {
	return starlarkgo.Bool(v.selection != nil && v.selection.Length() > 0)
}

func (v nodeValue) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", v.Type())
}

func (v nodeValue) AttrNames() []string {
	return []string{"next", "select", "select_first"}
}

func (v nodeValue) Attr(name string) (starlarkgo.Value, error) {
	switch name {
	case "next":
		return starlarkgo.NewBuiltin("node.next", v.next), nil
	case "select":
		return starlarkgo.NewBuiltin("node.select", v.selectNodes), nil
	case "select_first":
		return starlarkgo.NewBuiltin("node.select_first", v.selectFirst), nil
	default:
		return nil, nil
	}
}

func (v nodeValue) next(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	var css string
	if err := starlarkgo.UnpackArgs("next", args, kwargs, "css", &css); err != nil {
		return nil, err
	}

	return firstNodeOrNone(v.selection.NextFiltered(css)), nil
}

func (v nodeValue) selectNodes(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	var css string
	if err := starlarkgo.UnpackArgs("select", args, kwargs, "css", &css); err != nil {
		return nil, err
	}

	return nodesFromSelection(v.selection.Find(css)), nil
}

func (v nodeValue) selectFirst(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	var css string
	if err := starlarkgo.UnpackArgs("select_first", args, kwargs, "css", &css); err != nil {
		return nil, err
	}

	return firstNodeOrNone(v.selection.FindMatcher(goquery.Single(css))), nil
}

// nodesFromSelection converts a goquery selection into a Starlark node list.
func nodesFromSelection(selection *goquery.Selection) starlarkgo.Value {
	items := make([]starlarkgo.Value, 0, selection.Length())
	selection.Each(func(_ int, item *goquery.Selection) {
		items = append(items, newNodeValue(item))
	})

	return starlarkgo.NewList(items)
}

// firstNodeOrNone returns the first matching node, or Starlark None if empty.
func firstNodeOrNone(selection *goquery.Selection) starlarkgo.Value {
	first := selection.First()
	if first.Length() == 0 {
		return starlarkgo.None
	}

	return newNodeValue(first)
}

// predeclared returns builtins available to all extraction scripts.
func predeclared() starlarkgo.StringDict {
	return starlarkgo.StringDict{
		"attr":               starlarkgo.NewBuiltin("attr", attrBuiltin),
		"clone_without_text": starlarkgo.NewBuiltin("clone_without_text", cloneWithoutTextBuiltin),
		"collapse_space":     starlarkgo.NewBuiltin("collapse_space", collapseSpaceBuiltin),
		"first_attr":         starlarkgo.NewBuiltin("first_attr", firstAttrBuiltin),
		"first_text":         starlarkgo.NewBuiltin("first_text", firstTextBuiltin),
		"has_parent":         starlarkgo.NewBuiltin("has_parent", hasParentBuiltin),
		"node_name":          starlarkgo.NewBuiltin("node_name", nodeNameBuiltin),
		"path":               starlarkgo.NewBuiltin("path", pathBuiltin),
		"regex_match":        starlarkgo.NewBuiltin("regex_match", regexMatchBuiltin),
		"regex_replace":      starlarkgo.NewBuiltin("regex_replace", regexReplaceBuiltin),
		"safe_text":          starlarkgo.NewBuiltin("safe_text", safeTextBuiltin),
		"sha1":               starlarkgo.NewBuiltin("sha1", sha1Builtin),
		"text":               starlarkgo.NewBuiltin("text", textBuiltin),
		"trim":               starlarkgo.NewBuiltin("trim", trimBuiltin),
		"url_join":           starlarkgo.NewBuiltin("url_join", urlJoinBuiltin),
		"url_without_anchor": starlarkgo.NewBuiltin("url_without_anchor", urlWithoutAnchorBuiltin),
	}
}

// textBuiltin returns visible text for a node.
func textBuiltin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	var value starlarkgo.Value
	if err := starlarkgo.UnpackArgs("text", args, kwargs, "node", &value); err != nil {
		return nil, err
	}

	node, ok := value.(nodeValue)
	if !ok {
		return nil, fmt.Errorf("text: want node, got %s", value.Type())
	}

	return starlarkgo.String(node.selection.Text()), nil
}

// safeTextBuiltin returns visible text for a node, or an empty string for None.
func safeTextBuiltin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	var value starlarkgo.Value
	if err := starlarkgo.UnpackArgs("safe_text", args, kwargs, "node", &value); err != nil {
		return nil, err
	}

	if value == starlarkgo.None {
		return starlarkgo.String(""), nil
	}

	node, ok := value.(nodeValue)
	if !ok {
		return nil, fmt.Errorf("safe_text: want node or None, got %s", value.Type())
	}

	return starlarkgo.String(node.selection.Text()), nil
}

// firstTextBuiltin selects one descendant and returns its text, or empty string when absent.
func firstTextBuiltin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	var (
		value starlarkgo.Value
		css   string
	)
	if err := starlarkgo.UnpackArgs(
		"first_text",
		args,
		kwargs,
		"root",
		&value,
		"css",
		&css,
	); err != nil {
		return nil, err
	}

	node, err := selectFirstFromValue("first_text", value, css)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return starlarkgo.String(""), nil
	}

	return starlarkgo.String(node.Text()), nil
}

// firstAttrBuiltin selects one descendant and returns an attribute, or None when absent.
func firstAttrBuiltin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	var (
		value starlarkgo.Value
		css   string
		name  string
	)
	if err := starlarkgo.UnpackArgs(
		"first_attr",
		args,
		kwargs,
		"root",
		&value,
		"css",
		&css,
		"name",
		&name,
	); err != nil {
		return nil, err
	}

	node, err := selectFirstFromValue("first_attr", value, css)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return starlarkgo.None, nil
	}

	attrValue, exists := node.Attr(name)
	if !exists {
		return starlarkgo.None, nil
	}

	return starlarkgo.String(attrValue), nil
}

// selectFirstFromValue selects one node from a document or node wrapper.
func selectFirstFromValue(
	name string,
	value starlarkgo.Value,
	css string,
) (*goquery.Selection, error) {
	switch value := value.(type) {
	case documentValue:
		first := value.doc.GoqueryDocument().FindMatcher(goquery.Single(css))
		if first.Length() == 0 {
			return nil, nil
		}

		return first, nil
	case nodeValue:
		first := value.selection.FindMatcher(goquery.Single(css))
		if first.Length() == 0 {
			return nil, nil
		}

		return first, nil
	default:
		return nil, fmt.Errorf("%s: want document or node, got %s", name, value.Type())
	}
}

// attrBuiltin returns one attribute value for a node, or None when absent.
func attrBuiltin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	var (
		value starlarkgo.Value
		name  string
	)

	if err := starlarkgo.UnpackArgs(
		"attr",
		args,
		kwargs,
		"node",
		&value,
		"name",
		&name,
	); err != nil {
		return nil, err
	}

	node, ok := value.(nodeValue)
	if !ok {
		return nil, fmt.Errorf("attr: want node, got %s", value.Type())
	}

	attrValue, exists := node.selection.Attr(name)
	if !exists {
		return starlarkgo.None, nil
	}

	return starlarkgo.String(attrValue), nil
}

// hasParentBuiltin reports whether a node has an ancestor matching CSS.
func hasParentBuiltin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	var (
		value starlarkgo.Value
		css   string
	)
	if err := starlarkgo.UnpackArgs(
		"has_parent",
		args,
		kwargs,
		"node",
		&value,
		"css",
		&css,
	); err != nil {
		return nil, err
	}

	node, ok := value.(nodeValue)
	if !ok {
		return nil, fmt.Errorf("has_parent: want node, got %s", value.Type())
	}

	return starlarkgo.Bool(node.selection.ParentsFiltered(css).Length() > 0), nil
}

// cloneWithoutTextBuiltin removes matching descendants from a clone and returns text.
func cloneWithoutTextBuiltin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	var (
		value starlarkgo.Value
		css   string
	)
	if err := starlarkgo.UnpackArgs(
		"clone_without_text",
		args,
		kwargs,
		"node",
		&value,
		"css",
		&css,
	); err != nil {
		return nil, err
	}

	node, ok := value.(nodeValue)
	if !ok {
		return nil, fmt.Errorf("clone_without_text: want node, got %s", value.Type())
	}

	clone := node.selection.Clone()
	clone.Find(css).Each(func(_ int, match *goquery.Selection) {
		match.Remove()
	})

	return starlarkgo.String(clone.Text()), nil
}

// nodeNameBuiltin returns the HTML node name for a node.
func nodeNameBuiltin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	var value starlarkgo.Value
	if err := starlarkgo.UnpackArgs("node_name", args, kwargs, "node", &value); err != nil {
		return nil, err
	}

	node, ok := value.(nodeValue)
	if !ok {
		return nil, fmt.Errorf("node_name: want node, got %s", value.Type())
	}

	return starlarkgo.String(goquery.NodeName(node.selection)), nil
}
