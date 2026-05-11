def extract_page(pattern, doc, ctx):
    title = escape_html(collapse_space(first_text(doc, "h1")))
    page_url = ctx["url"]

    return [{
        "url": page_url,
        "urlWithoutAnchor": url_without_anchor(page_url),
        "objectID": sha1(page_url),
        "title": title,
    }]


extract(".*", extract_page)
