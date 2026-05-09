package starlark

import (
	"crypto/sha1"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	starlarkgo "go.starlark.net/starlark"
)

// trimBuiltin exposes strings.TrimSpace to Starlark scripts.
func trimBuiltin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	var value string
	if err := starlarkgo.UnpackArgs("trim", args, kwargs, "s", &value); err != nil {
		return nil, err
	}

	return starlarkgo.String(strings.TrimSpace(value)), nil
}

// collapseSpaceBuiltin normalizes all whitespace runs to single spaces.
func collapseSpaceBuiltin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	var value string
	if err := starlarkgo.UnpackArgs("collapse_space", args, kwargs, "s", &value); err != nil {
		return nil, err
	}

	return starlarkgo.String(strings.Join(strings.Fields(value), " ")), nil
}

// urlJoinBuiltin resolves a reference URL against a base URL.
func urlJoinBuiltin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	var (
		baseValue string
		refValue  string
	)

	if err := starlarkgo.UnpackArgs(
		"url_join",
		args,
		kwargs,
		"base",
		&baseValue,
		"ref",
		&refValue,
	); err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(baseValue)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}

	refURL, err := url.Parse(refValue)
	if err != nil {
		return nil, fmt.Errorf("parse ref url: %w", err)
	}

	return starlarkgo.String(baseURL.ResolveReference(refURL).String()), nil
}

// urlWithoutAnchorBuiltin removes the fragment from a URL string.
func urlWithoutAnchorBuiltin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	var value string
	if err := starlarkgo.UnpackArgs("url_without_anchor", args, kwargs, "url", &value); err != nil {
		return nil, err
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return nil, err
	}

	parsed.Fragment = ""

	return starlarkgo.String(parsed.String()), nil
}

// pathBuiltin extracts the path component from a URL string.
func pathBuiltin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	var value string
	if err := starlarkgo.UnpackArgs("path", args, kwargs, "url", &value); err != nil {
		return nil, err
	}

	return starlarkgo.String(pathFromURL(value)), nil
}

// pathFromURL extracts URL path and returns empty string for invalid URLs.
func pathFromURL(value string) string {
	parsed, err := url.Parse(value)
	if err != nil {
		return ""
	}

	return parsed.Path
}

// sha1Builtin returns a hex SHA-1 digest for stable record IDs.
func sha1Builtin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	var value string
	if err := starlarkgo.UnpackArgs("sha1", args, kwargs, "s", &value); err != nil {
		return nil, err
	}

	return starlarkgo.String(fmt.Sprintf("%x", sha1.Sum([]byte(value)))), nil
}

// regexMatchBuiltin reports whether a regular expression matches a string.
func regexMatchBuiltin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	var (
		pattern string
		value   string
	)

	if err := starlarkgo.UnpackArgs(
		"regex_match",
		args,
		kwargs,
		"pattern",
		&pattern,
		"s",
		&value,
	); err != nil {
		return nil, err
	}

	matched, err := regexp.MatchString(pattern, value)
	if err != nil {
		return nil, err
	}

	return starlarkgo.Bool(matched), nil
}

// regexReplaceBuiltin applies a regular-expression replacement.
func regexReplaceBuiltin(
	_ *starlarkgo.Thread,
	_ *starlarkgo.Builtin,
	args starlarkgo.Tuple,
	kwargs []starlarkgo.Tuple,
) (starlarkgo.Value, error) {
	var (
		pattern     string
		replacement string
		value       string
	)

	if err := starlarkgo.UnpackArgs(
		"regex_replace",
		args,
		kwargs,
		"pattern",
		&pattern,
		"repl",
		&replacement,
		"s",
		&value,
	); err != nil {
		return nil, err
	}

	expression, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return starlarkgo.String(expression.ReplaceAllString(value, replacement)), nil
}
