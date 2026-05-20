def extract_page(pattern, doc, ctx):
    page = page_record(doc, ctx)
    return [page] + heading_records(doc, page)


extract(".*", extract_page)


def page_record(doc, ctx):
    page_url = url_without_anchor(ctx["url"])
    title_value = title(doc)
    desc = description(doc)
    record = {
        "url": page_url,
        "urlWithoutAnchor": page_url,
        "recordType": "lvl1",
        "hierarchy": {},
        "position": 0,
        "objectID": object_id(page_url, "lvl1", 0),
    }
    if title_value != "":
        record["hierarchy"]["lvl1"] = title_value
    if desc != "":
        record["content"] = desc
    return record


def heading_records(doc, page):
    out = []
    current_hierarchy = clone_hierarchy(page["hierarchy"])
    position = 1

    for heading in headings(doc):
        level = heading["level"]
        if level <= 1:
            continue

        text_value = heading["text"]
        hierarchy = clone_hierarchy(current_hierarchy)
        set_hierarchy(hierarchy, level, text_value)
        clear_below(hierarchy, level)

        record = clone_record_base(page)
        record["recordType"] = "lvl" + str(level)
        record["hierarchy"] = hierarchy
        record["position"] = position
        record["objectID"] = object_id(page["url"], record["recordType"], position)

        out.append(record)
        current_hierarchy = clone_hierarchy(hierarchy)
        position += 1

    return out


def clone_record_base(page):
    return {
        "url": page["url"],
        "urlWithoutAnchor": page["urlWithoutAnchor"],
        "recordType": page["recordType"],
        "hierarchy": clone_hierarchy(page["hierarchy"]),
        "position": page["position"],
        "objectID": page["objectID"],
    }


def clone_hierarchy(hierarchy):
    out = {}
    for key in ["lvl0", "lvl1", "lvl2", "lvl3", "lvl4", "lvl5", "lvl6"]:
        if key in hierarchy:
            out[key] = hierarchy[key]
    return out


def set_hierarchy(hierarchy, level, value):
    hierarchy["lvl" + str(level)] = value


def clear_below(hierarchy, level):
    for item in [2, 3, 4, 5, 6]:
        key = "lvl" + str(item)
        if level < item and key in hierarchy:
            hierarchy.pop(key)


def object_id(record_url, record_type, position):
    base = path(record_url).strip("/").replace("/", "-")
    if base == "":
        base = sha1(record_url)
    if record_type == "lvl1":
        return base
    return base + "--" + record_type + "-" + str(position)
