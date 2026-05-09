# Record extraction for the Algolia docs

# DOM contract for algolia.com/doc pages.
# Keep selectors narrow: crawler emits records only for visible docs content.
HEADING_SELECTOR = "h2[id], h3[id], h4[id], h5[id], h6[id]"
PARAM_FIELD_SELECTOR = "div.param-head[id]"
PROSE_SELECTOR = "span[data-as=p], li"
CONTENT_SELECTOR = HEADING_SELECTOR + ", " + PROSE_SELECTOR + ", " + PARAM_FIELD_SELECTOR
CONTENT_ROOT_SELECTORS = ["#content", "#content-area"]

# Slug -> display-name exceptions used when building breadcrumb labels.
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


# Extractor functions. Register order matters: first matching URL path wins.
def extract_rest_api(pattern, doc, ctx):
    attrs = {"contentType": "api"}
    set_if_not_empty(attrs, "methodName", method_name_from_url(ctx["url"]))
    rest_path = breadcrumb_path(ctx["url"])
    if rest_path.startswith("/rest-api/ingestion/") and rest_path.split("/")[-1].endswith("v1"):
        attrs["variant"] = "legacy"
    page = docs_page_record(doc, ctx, attrs)
    return [page] + content_records(doc, ctx, page, CONTENT_ROOT_SELECTORS, CONTENT_SELECTOR)


def extract_api_parameters(pattern, doc, ctx):
    page = docs_page_record(doc, ctx, {"contentType": "api"})
    return [page] + content_records(doc, ctx, page, CONTENT_ROOT_SELECTORS, CONTENT_SELECTOR)


def extract_sdk(pattern, doc, ctx):
    attrs = {"contentType": "sdk"}
    sdk_path = breadcrumb_path(ctx["url"])
    if sdk_path.startswith("/libraries/sdk/methods/"):
        set_if_not_empty(attrs, "methodName", method_name_from_url(ctx["url"]))
    if sdk_path.startswith("/libraries/sdk/v1"):
        attrs["variant"] = "legacy"
    page = docs_page_record(doc, ctx, attrs)
    return [page] + content_records(doc, ctx, page, CONTENT_ROOT_SELECTORS, CONTENT_SELECTOR)


def extract_framework_integration(pattern, doc, ctx):
    page = docs_page_record(doc, ctx, {"contentType": "sdk"})
    return [page] + content_records(doc, ctx, page, CONTENT_ROOT_SELECTORS, CONTENT_SELECTOR)


def extract_guides(pattern, doc, ctx):
    page = docs_page_record(doc, ctx, {"contentType": "guide"})
    return [page] + content_records(doc, ctx, page, CONTENT_ROOT_SELECTORS, CONTENT_SELECTOR)


def extract_integration(pattern, doc, ctx):
    page = docs_page_record(doc, ctx, {"contentType": "integration"})
    return [page] + content_records(doc, ctx, page, CONTENT_ROOT_SELECTORS, CONTENT_SELECTOR)


def extract_docs(pattern, doc, ctx):
    page = docs_page_record(doc, ctx, {})
    return [page] + content_records(doc, ctx, page, CONTENT_ROOT_SELECTORS, CONTENT_SELECTOR)

# Register extractor functions. Order matters
extract("^/doc/rest-api/", extract_rest_api)
extract("^/doc/api-reference/api-parameters/", extract_api_parameters)
extract("^/doc/libraries/sdk/", extract_sdk)
extract("^/doc/framework-integration/", extract_framework_integration)
extract("^/doc/guides/", extract_guides)
extract("^/doc/integration/", extract_integration)
extract("^/doc/", extract_docs)


# Page-level metadata from stable docs selectors.
def docs_page_record(doc, ctx, attrs):
    meta = {
        "title": node_text(doc.select_first("h1#page-title")),
        "description": node_attr(doc.select_first("meta[name=description]"), "content"),
    }
    return page_record(ctx["url"], meta, attrs)


# Walk docs content nodes in DOM order.
def content_records(doc, ctx, page, root_selectors, content_selector):
    out = []
    current_hierarchy = clone_hierarchy(page["hierarchy"])
    current_url = page["url"]

    # Missing content root is valid: keep page record, skip child records.
    root = content_root(doc, root_selectors)
    if root == None:
        return out

    position = 1
    for node in root.select(content_selector):
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


# New and older docs layouts use different content container IDs.
def content_root(doc, root_selectors):
    for selector in root_selectors:
        root = doc.select_first(selector)
        if root != None:
            return root
    return None


# Page record anchors distinct-per-page behavior and carries shared metadata.
def page_record(page_url, meta, attrs):
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
    copy_if_present(attrs, record, "contentType")
    copy_if_present(attrs, record, "variant")
    copy_if_present(attrs, record, "methodName")
    if meta.get("description") != None and meta["description"] != "":
        record["content"] = meta["description"]
    if meta.get("title") != None and meta["title"] != "":
        record["hierarchy"]["lvl1"] = meta["title"]
    return record


# Heading records map h2-h6 to hierarchy lvl2-lvl6.
def heading_record(page, current_hierarchy, page_url, node, tag, position):
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


# Prose selector can match spans inside list items; skip to avoid duplicates.
def prose_record(page, current_hierarchy, current_url, node, position):
    if has_parent(node, "li"):
        return None
    value = node_text(node)
    if value == "":
        return None
    return content_record(page, current_hierarchy, current_url, value, position)


# Ignore list items whose meaningful text disappears after removing links.
def list_item_record(page, current_hierarchy, current_url, node, position):
    value = node_text(node)
    without_links = collapse_space(clone_without_text(node, "a"))
    if value == "" or without_links == "":
        return None
    return content_record(page, current_hierarchy, current_url, value, position)


# Generic content record: paragraph/list item attached to current section.
def content_record(page, current_hierarchy, current_url, value, position):
    record = clone_record_base(page)
    record["recordType"] = "content"
    record["content"] = value
    record["hierarchy"] = clone_hierarchy(current_hierarchy)
    record["position"] = position
    record["url"] = current_url
    record["objectID"] = object_id(current_url, "content", position)
    return record


# API parameter field record. Force field name into lvl3 for legacy relevance.
def field_record(page, current_hierarchy, page_url, node, position):
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


# Copy shared page fields into child records before overriding specifics.
def clone_record_base(page):
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


def set_if_not_empty(record, key, value):
    if value != "":
        record[key] = value


# Defensive copy: Starlark dicts are mutable and hierarchy evolves while walking DOM.
def clone_hierarchy(hierarchy):
    out = {}
    for key in ["lvl0", "lvl1", "lvl2", "lvl3", "lvl4", "lvl5", "lvl6"]:
        if key in hierarchy:
            out[key] = hierarchy[key]
    return out


def set_hierarchy(hierarchy, level, value):
    hierarchy["lvl" + str(level)] = value


# When entering a higher-level heading, remove stale lower-level ancestors.
def clear_below(hierarchy, level):
    for item in [3, 4, 5, 6]:
        key = "lvl" + str(item)
        if level < item and key in hierarchy:
            hierarchy.pop(key)


# Safe DOM text read: missing node -> "", whitespace collapsed.
def node_text(node):
    if node == None:
        return ""
    return collapse_space(text(node))


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


# Resolve record URL to page URL or page URL + heading/field anchor.
def url_with_anchor(page_url, anchor):
    if anchor == "":
        return page_url
    return url_without_anchor(page_url) + "#" + anchor


# Build Algolia objectID for a record.
# Page records use normalized URL as-is.
# Section/content/field records add type + DOM position for uniqueness.
def object_id(record_url, record_type, position):
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
    hierarchy = {}
    parts = []
    for i, segment in enumerate(segments):
        parts.append(segment)
        if i <= 5:
            hierarchy["lvl" + str(i)] = " > ".join(parts)
    return hierarchy


# Convert URL slug to display label, preserving known acronyms/product names.
def humanize_slug(value):
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
    if value == "":
        return value
    return value[0].upper() + value[1:]


# Derive camelCase methodName from the current page slug.
def method_name_from_url(page_url):
    value = breadcrumb_path(page_url).strip("/")
    parts = value.split("/")
    if len(parts) == 0:
        return ""
    last = parts[-1]
    if last == "":
        return ""
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


