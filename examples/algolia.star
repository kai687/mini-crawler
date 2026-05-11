# Recor extraction for the Algolia docs

# Only grab headings with `id` attribute
HEADING_SELECTOR = "h2[id], h3[id], h4[id], h5[id], h6[id]"
# Grab `<ParameterField />` components
PARAM_FIELD_SELECTOR = "div.param-head[id]"
# Mintlify doesn't use `<p>` for some reason, but `<span[data-as=p]>`
PROSE_SELECTOR = "span[data-as=p], li"
# What counts as "content"
CONTENT_SELECTOR = HEADING_SELECTOR + ", " + PROSE_SELECTOR + ", " + PARAM_FIELD_SELECTOR
# Only index content inside these root elements (to cut noise)
CONTENT_ROOT_SELECTORS = ["#content", "#content-area"]

# When deriving breadcrumbs from the URL, we turn the `lowercase` segments into `sentence case`,
# But we need to respect these special capitalization rules
SPECIAL_BREADCRUMB_CASING = {
    "ai": "AI",
    "algolia": "Algolia",
    "api": "API",
    "apis": "APIs",
    "autocomplete": "Autocomplete",
    "cli": "CLI",
    "cpu": "CPU",
    "css": "CSS",
    "csv": "CSV",
    "dns": "DNS",
    "faq": "FAQ",
    "graphql": "GraphQL",
    "html": "HTML",
    "http": "HTTP",
    "https": "HTTPS",
    "id": "ID",
    "ids": "IDs",
    "ios": "iOS",
    "ip": "IP",
    "ipv4": "IPv4",
    "ipv6": "IPv6",
    "javascript": "JavaScript",
    "json": "JSON",
    "jwt": "JWT",
    "ml": "ML",
    "ocr": "OCR",
    "pdf": "PDF",
    "php": "PHP",
    "ram": "RAM",
    "rest": "REST",
    "rpc": "RPC",
    "sdk": "SDK",
    "sql": "SQL",
    "ssh": "SSH",
    "ssl": "SSL",
    "sso": "SSO",
    "tcp": "TCP",
    "tls": "TLS",
    "typescript": "TypeScript",
    "ui": "UI",
    "uid": "UID",
    "url": "URL",
    "urls": "URLs",
    "utc": "UTC",
    "ux": "UX",
    "xml": "XML",
}

# Extra casing exceptions for SDK/API method names.
SPECIAL_METHOD_CASING = {
    "ab": "AB",
}

def handle_api(pattern, doc, ctx):
    """Handle REST API pages."""
    page = page_record(doc, ctx)
    page["contentType"] = "api"

    # We need to derive method names from the URL, as it's not otherwise exposed in the HTML
    # This lets users search for the exact method names.
    method_name = method_name_from_url(ctx["url"])
    if method_name != "":
        page["methodName"] = method_name
    rest_path = breadcrumb_path(ctx["url"])

    # There are some legacy methods in the Ingestion API, which we can recognize by their URL.
    # Otherwise, this is also not exposed in HTML
    if rest_path.startswith("/rest-api/ingestion/") and rest_path.split("/")[-1].endswith("v1"):
        page["variant"] = "legacy"

    return [page] + content_records(doc, ctx, page)


def handle_api_parameters(pattern, doc, ctx):
    """Handle API parameter reference pages."""
    page = page_record(doc, ctx)
    page["contentType"] = "api"
    return [page] + content_records(doc, ctx, page)


def handle_sdk(pattern, doc, ctx):
    """Handle SDK pages."""
    page = page_record(doc, ctx)
    page["contentType"] = "sdk"

    sdk_path = breadcrumb_path(ctx["url"])

    # Build the method name from URL
    if sdk_path.startswith("/libraries/sdk/methods/"):
        method_name = method_name_from_url(ctx["url"])
        if method_name != "":
            page["methodName"] = method_name

    # Handle legacy pages
    if sdk_path.startswith("/libraries/sdk/v1"):
        page["variant"] = "legacy"

    return [page] + content_records(doc, ctx, page)


def handle_web_frameworks(pattern, doc, ctx):
    """Handle web-framework pages."""
    page = page_record(doc, ctx)
    page["contentType"] = "sdk"

    return [page] + content_records(doc, ctx, page)


def handle_guides(pattern, doc, ctx):
    """Handle prose pages."""
    page = page_record(doc, ctx)
    page["contentType"] = "guide"

    return [page] + content_records(doc, ctx, page)


def handle_integrations(pattern, doc, ctx):
    """Handle integration pages."""
    page = page_record(doc, ctx)
    page["contentType"] = "integration"

    return [page] + content_records(doc, ctx, page)


def handle_other(pattern, doc, ctx):
    """Handle other pages."""
    page = page_record(doc, ctx)
    return [page] + content_records(doc, ctx, page)

# Register extractor functions. Order matters
extract("^/doc/rest-api/", handle_api)
extract("^/doc/api-reference/api-parameters/", handle_api_parameters)
extract("^/doc/libraries/sdk/", handle_sdk)
extract("^/doc/framework-integration/", handle_web_frameworks)
extract("^/doc/guides/", handle_guides)
extract("^/doc/integration/", handle_integrations)
extract("^/doc/", handle_other)

# Walk docs content nodes in DOM order.
def content_records(doc, ctx, page):
    out = []
    current_hierarchy = clone_hierarchy(page["hierarchy"])
    current_url = page["url"]

    # Missing content root is valid: keep page record, skip child records.
    root = content_root(doc)
    if root == None:
        return out

    position = 1
    for node in root.select(CONTENT_SELECTOR):
        tag = node_name(node)
        if is_heading_tag(tag):
            # Headings update current hierarchy and current anchor URL.
            record = heading_record(page, current_hierarchy, ctx["url"], node, tag, position)
            if record != None:
                out.append(record)
                current_hierarchy = clone_hierarchy(record["hierarchy"])
                current_url = record["url"]
                position += 1
        elif tag == "span":
            # Prose inherits nearest heading URL/hierarchy.
            record = prose_record(page, current_hierarchy, current_url, node, position)
            if record != None:
                out.append(record)
                position += 1
        elif tag == "li":
            # List items become content records unless they are links-only noise.
            record = list_item_record(page, current_hierarchy, current_url, node, position)
            if record != None:
                out.append(record)
                position += 1
        elif tag == "div":
            # API parameter rows act like lvl3 entries, with field metadata as content.
            record = field_record(page, current_hierarchy, ctx["url"], node, position)
            if record != None:
                out.append(record)
                current_hierarchy = clone_hierarchy(record["hierarchy"])
                current_url = record["url"]
                position += 1

    return out


def content_root(doc):
    for selector in CONTENT_ROOT_SELECTORS:
        root = doc.select_first(selector)
        if root != None:
            return root
    return None


def page_record(doc, ctx):
    """Base record that only depends on page-level properties, like the `h1`, or the URL structure."""
    page_url = ctx["url"]
    title = node_text(doc.select_first("h1#page-title"))
    description = node_attr(doc.select_first("meta[name=description]"), "content")
    # We need this later for deduplication with `distinct`
    base_url = url_without_anchor(page_url)
    segments = breadcrumb_segments(base_url)
    record = {
        "url": base_url,
        "urlWithoutAnchor": base_url,
        "recordType": "lvl1",
        "hierarchy": {},
        "position": 0,
        "objectID": object_id(base_url, "lvl1", 0),
    }
    if len(segments) > 0:
        record["breadcrumbSegments"] = segments
        record["breadcrumbHierarchy"] = breadcrumb_hierarchy(segments)
    if description != "":
        record["content"] = escape_html(description)
    if title != "":
        record["hierarchy"]["lvl1"] = title
    return record


def heading_record(page, current_hierarchy, page_url, node, tag, position):
    """A heading record."""
    value = node_text(node)
    if value == "":
        return None
    level = int(tag[1:])
    anchor = node_attr(node, "id")
    record_url = url_with_anchor(page_url, anchor)
    hierarchy = clone_hierarchy(current_hierarchy)
    set_hierarchy(hierarchy, level, value)
    clear_below(hierarchy, level)
    record = clone_record_base(page)
    record["recordType"] = "lvl" + str(level)
    record["hierarchy"] = hierarchy
    record["position"] = position
    record["url"] = record_url
    record["objectID"] = object_id(record_url, record["recordType"], position)
    return record


def prose_record(page, current_hierarchy, current_url, node, position):
    """
    Records for paragraphs. Paragraphs inside list items are ignored.
    Otherwise, you'd get duplicated records.
    """
    if has_parent(node, "li"):
        return None
    value = node_text(node)
    if value == "":
        return None
    return content_record(page, current_hierarchy, current_url, value, position)


def list_item_record(page, current_hierarchy, current_url, node, position):
    """Record for list items. Remove links to exclude lists that are only links."""
    value = node_text(node)
    without_links = collapse_space(clone_without_text(node, "a"))
    if value == "" or without_links == "":
        return None
    return content_record(page, current_hierarchy, current_url, value, position)


def content_record(page, current_hierarchy, current_url, value, position):
    """A record for matching paragraphs or list items attached to the current section."""
    record = clone_record_base(page)
    record["recordType"] = "content"
    record["content"] = value
    record["hierarchy"] = clone_hierarchy(current_hierarchy)
    record["position"] = position
    record["url"] = current_url
    record["objectID"] = object_id(current_url, "content", position)
    return record


def field_record(page, current_hierarchy, page_url, node, position):
    """A record for <ParameterField> components. The field name is set as the lvl3 hierarchy value."""
    anchor = node_attr(node, "id")
    if anchor == "":
        return None
    name = node_text(node.select_first("[data-component-part=field-name]"))
    if name == "":
        return None
    hierarchy = clone_hierarchy(current_hierarchy)
    set_hierarchy(hierarchy, 3, name)
    clear_below(hierarchy, 3)
    record_url = url_with_anchor(page_url, anchor)
    record = clone_record_base(page)
    record["recordType"] = "field"
    desc = field_description(node)
    if desc != "":
        record["content"] = desc
    record["hierarchy"] = hierarchy
    record["position"] = position
    record["url"] = record_url
    record["objectID"] = object_id(record_url, "field", position)
    return record


# Build concise field summary from type/required pills and first prose line.
def field_description(node):
    """Get required/info from rendered <ParameterField> components."""
    parts = []
    pill = node_text(node.select_first("[data-component-part=field-info-pill]"))
    if pill != "":
        parts.append(pill)
    required = node_text(node.select_first("[data-component-part=field-required-pill]"))
    if required != "":
        parts.append("required")
    desc_block = node.next("div.mt-4")
    if desc_block != None:
        prose = desc_block.select_first("div")
        if prose != None:
            first_p = prose.select_first("p")
            text_value = node_text(first_p)
            if text_value != "":
                parts.append(text_value)
    return ". ".join(parts)


def clone_record_base(page):
    """Copy shared fields from the parent record."""
    record = {
        "url": page["url"],
        "urlWithoutAnchor": page["urlWithoutAnchor"],
        "recordType": page["recordType"],
        "hierarchy": clone_hierarchy(page["hierarchy"]),
        "position": page["position"],
        "objectID": page["objectID"],
    }
    copy_if_present(page, record, "breadcrumbSegments")
    copy_if_present(page, record, "breadcrumbHierarchy")
    copy_if_present(page, record, "contentType")
    copy_if_present(page, record, "variant")
    copy_if_present(page, record, "methodName")
    return record


def copy_if_present(src, dst, key):
    if key in src:
        dst[key] = src[key]


# Defensive copy: Starlark dicts are mutable and hierarchy evolves while walking DOM.
def clone_hierarchy(hierarchy):
    out = {}
    for key in ["lvl0", "lvl1", "lvl2", "lvl3", "lvl4", "lvl5", "lvl6"]:
        if key in hierarchy:
            out[key] = hierarchy[key]
    return out


def set_hierarchy(hierarchy, level, value):
    hierarchy["lvl" + str(level)] = value


# When entering a higher-level heading, remove lower-level ancestors.
def clear_below(hierarchy, level):
    for item in [3, 4, 5, 6]:
        key = "lvl" + str(item)
        if level < item and key in hierarchy:
            hierarchy.pop(key)


# Safe DOM text read for record attributes: missing node -> "", whitespace collapsed and HTML escaped.
def node_text(node):
    if node == None:
        return ""
    return escape_html(collapse_space(text(node)))


# Safe DOM attr read: missing node/attr -> "", whitespace collapsed.
def node_attr(node, name):
    if node == None:
        return ""
    value = attr(node, name)
    if value == None:
        return ""
    return collapse_space(value)


def is_heading_tag(tag):
    return tag in ["h2", "h3", "h4", "h5", "h6"]


def url_with_anchor(page_url, anchor):
    """Resolve URLs with anchors."""
    if anchor == "":
        return page_url
    return url_without_anchor(page_url) + "#" + anchor


def object_id(record_url, record_type, position):
    """
    Construct a `objectID` from URLs, using the position in the DOM
    and record_type (heading, content) to make them unique.
    """
    value = record_url
    parsed_path = path(value)
    if regex_match("^https?://", value):
        # Drop scheme/host so objectIDs stay stable across domains.
        value = parsed_path
        if "#" in record_url:
            value = value + "#" + record_url.split("#", 1)[1]
        if "?" in record_url:
            value = value + "?" + record_url.split("?", 1)[1].split("#", 1)[0]
    base = value.strip("/").replace("/", "-")
    if record_type == "lvl1":
        return base
    return base + "--" + record_type + "-" + str(position)


# URL path used for breadcrumbs/content type. Strip /doc prefix.
def breadcrumb_path(page_url):
    """Remove the `/doc prefix from URLs`."""
    value = path(url_without_anchor(page_url))
    if value == "" or value == "/":
        return ""
    if value.startswith("/doc"):
        value = value[len("/doc"):]
        if value == "" or value == "/":
            return ""
    return value


# Breadcrumbs come from URL folders, excluding current page slug.
def breadcrumb_segments(page_url):
    """
    Create a list of breadcrumb segments from the current URL.
    So, `/doc/guides/do/something` becomes `["guides", "do"]`
    """
    value = breadcrumb_path(page_url).strip("/")
    if value == "":
        return []
    parts = value.split("/")
    if len(parts) <= 1:
        return []
    out = []
    for part in parts[:-1]:
        if part != "":
            out.append(humanize_slug(part))
    return out


# Build cumulative Algolia-style breadcrumb hierarchy levels.
def breadcrumb_hierarchy(segments):
    """
    Build a cumulative hierarchy from the breadcrumb segments.
    This can then be used in a `HierarchicalFacets` component from InstantSearch for hierarchical navigation.
    Eg. `/doc/guides/do/something` becomes `{"lvl1": "Guides", "lvl2": "Guides > Do"}`
    """
    hierarchy = {}
    parts = []
    for i, segment in enumerate(segments):
        parts.append(segment)
        if i <= 5:
            hierarchy["lvl" + str(i)] = " > ".join(parts)
    return hierarchy


def humanize_slug(value):
    """Convert a URL slug to a human-readable label, preserving special acronyms/product names."""
    parts = value.replace("-", " ").split()
    for i, part in enumerate(parts):
        normalized = part.lower()
        if normalized in SPECIAL_BREADCRUMB_CASING:
            parts[i] = SPECIAL_BREADCRUMB_CASING[normalized]
        elif i == 0:
            parts[i] = sentence_case(normalized)
        else:
            parts[i] = normalized
    return " ".join(parts)


def sentence_case(value):
    """Convert a string to sentence case."""
    if value == "":
        return value
    return value[0].upper() + value[1:]


# Derive camelCase methodName from the current page slug.
def method_name_from_url(page_url):
    """
    Derive a camelCase method name from the page_url which is in kebab case.
    """
    value = breadcrumb_path(page_url).strip("/")
    parts = value.split("/")
    if len(parts) == 0:
        return ""
    last = parts[-1]
    if last == "":
        return ""
    # This handles situations like `--` better
    tokens = last.replace("-", " ").split()
    for i, token in enumerate(tokens):
        normalized = token.lower()
        if normalized in SPECIAL_METHOD_CASING:
            cased = SPECIAL_METHOD_CASING[normalized]
            if i == 0:
                tokens[i] = cased.lower()
            else:
                tokens[i] = cased
        elif i == 0:
            tokens[i] = normalized
        else:
            tokens[i] = sentence_case(normalized)
    return "".join(tokens)


