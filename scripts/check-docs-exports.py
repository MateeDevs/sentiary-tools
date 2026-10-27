"""Check the built llmstxt outputs without contacting external services."""

from html.parser import HTMLParser
from pathlib import Path
import re
import tomllib
from urllib.parse import unquote, urljoin, urlsplit

import markdown


class Content(HTMLParser):
    def __init__(self, html):
        super().__init__(convert_charrefs=True)
        self.text = []
        self.links = []
        self.code = []
        self.cells = []
        self.details = []
        self._captures = []
        self.feed(html)

    def handle_starttag(self, tag, attrs):
        attrs = dict(attrs)
        if tag == "a" and "href" in attrs:
            self.links.append(attrs["href"])
        if tag in ("pre", "td", "th", "details"):
            self._captures.append((tag, []))

    def handle_data(self, data):
        self.text.append(data)
        for _, parts in self._captures:
            parts.append(data)

    def handle_endtag(self, tag):
        if self._captures and self._captures[-1][0] == tag:
            _, parts = self._captures.pop()
            text = "".join(parts)
            if tag == "pre":
                self.code.append(text.strip())
            elif tag == "details":
                self.details.append(text)
            else:
                self.cells.append(" ".join(text.split()))


def require(condition, message):
    if not condition:
        raise SystemExit(message)


def read_output(path):
    require(path.is_file(), f"Missing output: {path}")
    text = path.read_text(encoding="utf-8")
    require(text.strip(), f"Empty output: {path}")
    return text


def export_path(source):
    path = Path(source)
    return path if path.name == "index.md" else path.with_suffix("") / "index.md"


def main():
    project = tomllib.loads(Path("zensical.toml").read_text())["project"]
    plugin = project["plugins"]["llmstxt"]
    site = Path(project["site_dir"])
    base = project["site_url"].rstrip("/") + "/"
    require("base_url" not in plugin, "Export URLs must follow site_url")
    index = read_output(site / "llms.txt")
    full = read_output(site / plugin["full_output"])
    pages = [
        export_path(source)
        for entries in plugin["sections"].values()
        for entry in entries
        for source in (entry if isinstance(entry, dict) else [entry])
    ]
    expected_urls = [urljoin(base, page.as_posix()) for page in pages]
    index_links = Content(markdown.markdown(index)).links
    require(
        [link for link in index_links if urlsplit(link).path.endswith(".md")]
        == expected_urls,
        "llms.txt must link to all selected pages in configured order using site_url",
    )
    require(
        set(site.rglob("*.md")) == {site / page for page in pages},
        "Markdown exports must contain only the explicitly selected pages",
    )

    def check_links(text, url):
        for link in Content(markdown.markdown(text)).links:
            target = urlsplit(urljoin(url, link))
            origin = urlsplit(base)
            if (target.scheme, target.netloc) != (origin.scheme, origin.netloc):
                continue
            require(target.path.startswith(origin.path), f"Link outside site_url: {link}")
            relative = unquote(target.path[len(origin.path):])
            # nginx proxies the authoritative schema; no static copy belongs here.
            if relative == "api/openapi.yaml":
                continue
            destination = site / relative
            if destination.is_dir():
                destination /= "index.html"
            require(destination.is_file(), f"Broken link in {url}: {link}")

    check_links(index, urljoin(base, "llms.txt"))
    check_links(full, urljoin(base, plugin["full_output"]))
    for page, url in zip(pages, expected_urls):
        text = read_output(site / page)
        require(text.strip() in full, f"Full output omits {page}")
        html = read_output((site / page).with_suffix(".html"))
        require('data-md-copy-url="' in html, f"Missing Copy as Markdown action: {page}")
        article = re.search(r"<article\b[^>]*>(.*?)</article>", html, re.S)
        require(article is not None, f"Missing rendered article: {page}")
        rendered = Content(article.group(1))
        exported = Content(markdown.markdown(text, extensions=["tables", "pymdownx.superfences"]))
        require(rendered.code == exported.code, f"Code blocks changed or omitted: {page}")
        require(rendered.cells == exported.cells, f"Table cells changed or omitted: {page}")
        plain = " ".join(" ".join(exported.text).split())
        for detail in rendered.details:
            require(" ".join(detail.split()) in plain, f"Collapsed content omitted: {page}")
        check_links(text, url)
    print(f"Verified llms.txt, {plugin['full_output']}, and {len(pages)} Markdown exports")


if __name__ == "__main__":
    main()
