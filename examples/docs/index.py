import datetime
import json
import re
import urllib.request
from pathlib import Path

from pysitegen import (
    a,
    asset,
    canvas_background,
    div,
    footer,
    h1,
    h2,
    header,
    page,
    section,
    span,
    tag,
)
from pysitegen.components import RawHtml
from pysitegen.visual import visual_canvas, visual_scene


def fetch_github_releases():
    url = "https://api.github.com/repos/ujjwalvivek/loom/releases"
    req = urllib.request.Request(url, headers={"User-Agent": "PySiteGen-Builder/1.0"})
    releases = []
    try:
        with urllib.request.urlopen(req, timeout=5) as response:
            releases = json.loads(response.read().decode())
    except Exception as e:
        print(f"Error fetching releases: {e}")

    if not releases:
        releases = [
            {
                "tag_name": "v0.1.0",
                "published_at": "2026-06-12T00:00:00Z",
                "assets": [
                    {
                        "name": "loom-mario-term_linux_amd64.tar.gz",
                        "browser_download_url": "https://github.com/ujjwalvivek/loom/releases/download/v0.1.0/loom-mario-term_linux_amd64.tar.gz",
                    },
                    {
                        "name": "loom-mario_linux_amd64.tar.gz",
                        "browser_download_url": "https://github.com/ujjwalvivek/loom/releases/download/v0.1.0/loom-mario_linux_amd64.tar.gz",
                    },
                    {
                        "name": "loom-mario-term_darwin_amd64.tar.gz",
                        "browser_download_url": "https://github.com/ujjwalvivek/loom/releases/download/v0.1.0/loom-mario-term_darwin_amd64.tar.gz",
                    },
                    {
                        "name": "loom-mario-term_darwin_arm64.tar.gz",
                        "browser_download_url": "https://github.com/ujjwalvivek/loom/releases/download/v0.1.0/loom-mario-term_darwin_arm64.tar.gz",
                    },
                    {
                        "name": "loom-mario_darwin_amd64.tar.gz",
                        "browser_download_url": "https://github.com/ujjwalvivek/loom/releases/download/v0.1.0/loom-mario_darwin_amd64.tar.gz",
                    },
                    {
                        "name": "loom-mario_darwin_arm64.tar.gz",
                        "browser_download_url": "https://github.com/ujjwalvivek/loom/releases/download/v0.1.0/loom-mario_darwin_arm64.tar.gz",
                    },
                    {
                        "name": "loom-mario-term_windows_amd64.tar.gz",
                        "browser_download_url": "https://github.com/ujjwalvivek/loom/releases/download/v0.1.0/loom-mario-term_windows_amd64.tar.gz",
                    },
                    {
                        "name": "loom-mario_windows_amd64.tar.gz",
                        "browser_download_url": "https://github.com/ujjwalvivek/loom/releases/download/v0.1.0/loom-mario_windows_amd64.tar.gz",
                    },
                ],
            }
        ]

    return releases


def format_date(date_str):
    try:
        dt = datetime.datetime.strptime(date_str.split("T")[0], "%Y-%m-%d")
        return dt.strftime("%b %d, %Y")
    except Exception:
        return date_str

DEFAULT_CHECKSUMS = """7612392d626aa325253d4fdcafd1fa3835e3ad7695dc95ba82e63ea2e9b8517e  loom-mario-term_darwin_amd64.tar.gz
cb83a78c63c49d6163dba74a3da52d8534352b0d1186e7e2916315977792ea22  loom-mario-term_darwin_arm64.tar.gz
4494250f1864061d05070ae4b795b897a866b3d614b1bf635b4894f7b2a21e56  loom-mario-term_linux_amd64.tar.gz
6d4e1cfb77721a10e033de521bba94a08e387678c05c9fe6f479add2ab49adfa  loom-mario-term_windows_amd64.tar.gz
4887cacf14c30a9b8b4278878d340234bd3b6b60a344186a4c408cd20fe3ec3a  loom-mario_darwin_amd64.tar.gz
5c5621b8c5073130e7d1e8f436a5735144bde63101501738a3bad709e9097809  loom-mario_darwin_arm64.tar.gz
cc8f9cf2dccad45e3f0586e1b455c0f9511354e170120e23c632ae017fc51686  loom-mario_linux_amd64.tar.gz
c8924cedfc05e2879b85b2f395e13b9920bd66be4da2446f9f2c691372857eef  loom-mario_windows_amd64.tar.gz"""


def fetch_checksums(assets):
    checksum_asset = None
    for asset_item in assets:
        name = asset_item.get("name", "")
        if "checksums" in name.lower():
            checksum_asset = asset_item
            break

    if not checksum_asset:
        return DEFAULT_CHECKSUMS

    url = checksum_asset.get("browser_download_url", "")
    if not url:
        return DEFAULT_CHECKSUMS

    try:
        req = urllib.request.Request(url, headers={"User-Agent": "PySiteGen-Builder/1.0"})
        with urllib.request.urlopen(req, timeout=5) as response:
            return response.read().decode("utf-8").strip()
    except Exception as e:
        print(f"Error fetching checksums: {e}")
        return DEFAULT_CHECKSUMS


def group_assets(assets):
    os_map = {"linux": "LINUX", "darwin": "MAC", "windows": "WIN"}
    known = {"loom-mario-term": "TERM", "loom-mario": "GUI"}
    groups = {"MAC": [], "LINUX": [], "WIN": []}

    for asset_item in assets:
        name = asset_item.get("name", "")
        if not name.endswith(".tar.gz"):
            continue
        m = re.match(r"^(.+?)_(linux|darwin|windows)_(amd64|arm64)\.tar\.gz$", name)
        if m:
            bin_name = known.get(m.group(1), m.group(1).upper())
            os_name = os_map.get(m.group(2), m.group(2).upper())
            arch_name = m.group(3)
            url = asset_item.get("browser_download_url", "#")
            if os_name in groups:
                groups[os_name].append({"bin": bin_name, "arch": arch_name, "url": url})
    return groups


def build_downloads_html():
    releases = fetch_github_releases()
    if not releases:
        return None, div("NO DOWNLOADS FOUND", class_="no-assets")

    rel = releases[0]
    tag_name = rel.get("tag_name", "v0.0.0")
    pub_date = format_date(rel.get("published_at", ""))
    groups = group_assets(rel.get("assets", []))
    checksums_content = fetch_checksums(rel.get("assets", []))

    highlighted_lines = []
    for line in checksums_content.splitlines():
        parts = line.strip().split(None, 1)
        if len(parts) == 2:
            hash_val, filename = parts
            padded_hash = hash_val.ljust(64)
            highlighted_lines.append(f'<span class="sha-hash">{padded_hash}</span>  {filename}')
        else:
            highlighted_lines.append(line)
    checksums_content_html = "\n".join(highlighted_lines)

    meta = div(
        span(f"PUBLISHED: {pub_date}", class_="downloads-pub-text"),
        span(f"[{tag_name}]", class_="downloads-badge"),
        class_="downloads-header-meta",
    )

    grid_cards = []
    for os_name in ["MAC", "LINUX", "WIN"]:
        list_assets = groups.get(os_name, [])
        rows = []
        if not list_assets:
            rows.append(div("NO BINARIES AVAILABLE", class_="no-assets"))
        else:
            bin_groups = {}
            for item in list_assets:
                bin_groups.setdefault(item["bin"], []).append(item)
            
            for bin_name in sorted(bin_groups.keys(), key=lambda x: (0 if x == "TERM" else 1)):
                items = bin_groups[bin_name]
                items.sort(key=lambda x: (0 if x["arch"] == "amd64" else 1))
                
                buttons = []
                for item in items:
                    buttons.append(
                        a(
                            f"[{item['arch'].upper()}]",
                            href=item["url"],
                            class_="dl-link",
                        )
                    )
                
                rows.append(
                    div(
                        span(bin_name, class_=f"dl-bin bin-{bin_name}"),
                        div(
                            *buttons,
                            class_="dl-buttons",
                        ),
                        class_="dl-row",
                    )
                )

        grid_cards.append(
            div(
                div(os_name, class_="platform-header"),
                div(*rows, class_="platform-body"),
                class_="platform-card",
            )
        )

    grid = div(*grid_cards, class_="downloads-grid")

    checksums_box = div(
        div(
            span("checksums"),
            tag(
                "button",
                "[COPY]",
                class_="pixel-btn copy-btn",
                data_cmd=checksums_content,
            ),
            class_="checksums-header",
        ),
        RawHtml(f'<pre class="checksums-pre">{checksums_content_html}</pre>'),
        class_="checksums-box",
    )

    return meta, [grid, checksums_box]


ROOT = Path(__file__).resolve().parent
ASSETS = ROOT / "assets"

scene = visual_scene(
    "background",
    ("grid", {"opacity": 2.0, "spacing": 120, "warpAmount": 5, "speed": 0.05, "color": "primary",}),
    ("particles", {"opacity": 1, "density": 1.0, "speed": 0.03, "minRadius": 420, "maxRadius": 1050,}),
    ("glitch", {"opacity": 0.85, "threshold": 0.98, "speed": 0.15}),
    ("streams", {"opacity": 0.65, "density": 0.7, "speed": 0.15}),
    ("scanlines", {"opacity": 0.75, "spacing": 5, "speed": 0.015}),
    ("vignette", {"opacity": 0.8}),
)


def build():
    downloads_meta, downloads_grid = build_downloads_html()
    h2_args: list = ["DOWNLOADS"]
    if downloads_meta:
        h2_args.append(downloads_meta)

    return page(
        visual_canvas(scene, id="mario-bg", fps=30),
        div(
            "LOOM ENGINE",
            class_="top-logo",
        ),
        div(
            a(
                tag(
                    "img",
                    src="https://echopoint.ujjwalvivek.com/svg/badges/stars?bg=111c44&badgeColor=003566&textColor=f8fafc&border=0066aa&borderWidth=2&rx=0&px=6&py=4&repo=loom&logo=github",
                    alt="GitHub Stars",
                ),
                href="https://github.com/ujjwalvivek/loom",
                target="_blank",
                rel="noreferrer",
                class_="github-badge-link",
            ),
            tag(
                "button",
                "[LIGHT]",
                id="theme-toggle",
                class_="theme-btn",
                aria_label="Toggle theme",
            ),
            class_="top-controls",
        ),
        div(
            header(
                div(
                    h1(
                        span("LOOM", class_="title-line title-shadow"),
                        span("MARIO", class_="title-line title-main"),
                        class_="game-title",
                    ),
                    class_="pixel-border",
                ),
                class_="title-screen",
            ),
            section(
                h2(
                    "QUICK INSTALL",
                    div(
                        tag(
                            "button",
                            "[UNIX]",
                            class_="tab-btn active",
                            data_tab_group="install",
                            data_tab_target="unix",
                        ),
                        tag(
                            "button",
                            "[WIN]",
                            class_="tab-btn",
                            data_tab_group="install",
                            data_tab_target="windows",
                        ),
                        tag(
                            "button",
                            "[MANUAL]",
                            class_="tab-btn",
                            data_tab_group="install",
                            data_tab_target="manual",
                        ),
                        class_="tab-bar",
                    ),
                    class_="section-title",
                ),
                div(
                    div(
                        div(
                            div(
                                span("TERMINAL VERSION"),
                                div(
                                    span("runs anywhere", class_="badge"),
                                    tag(
                                        "button",
                                        "[COPY]",
                                        class_="pixel-btn copy-btn",
                                        data_cmd="curl -sS https://raw.githubusercontent.com/ujjwalvivek/loom/main/examples/scripts/install.sh | sh",
                                    ),
                                    class_="cli-actions",
                                ),
                                class_="cli-header",
                            ),
                            RawHtml(
                                '<pre class="language-bash has-highlight"><span class="prompt">$</span> curl -sS <a href="https://raw.githubusercontent.com/ujjwalvivek/loom/main/examples/scripts/install.sh" target="_blank" class="code-link">https://raw.githubusercontent.com/ujjwalvivek/loom/main/examples/scripts/install.sh</a> | sh\n<span class="prompt">$</span> loom-mario-term</pre>'
                            ),
                            class_="cli-box",
                        ),
                        div(
                            div(
                                span("GUI VERSION"),
                                div(
                                    span("needs OpenGL 3.3+", class_="badge badge-opengl"),
                                    a("needs GLFW", href="https://www.glfw.org/", target="_blank", class_="badge badge-glfw"),
                                    tag(
                                        "button",
                                        "[COPY]",
                                        class_="pixel-btn copy-btn",
                                        data_cmd="curl -sS https://raw.githubusercontent.com/ujjwalvivek/loom/main/examples/scripts/install.sh | sh -s -- loom-mario",
                                    ),
                                    class_="cli-actions",
                                ),
                                class_="cli-header",
                            ),
                            RawHtml(
                                '<pre class="language-bash has-highlight"><span class="prompt">$</span> curl -sS <a href="https://raw.githubusercontent.com/ujjwalvivek/loom/main/examples/scripts/install.sh" target="_blank" class="code-link">https://raw.githubusercontent.com/ujjwalvivek/loom/main/examples/scripts/install.sh</a> | sh -s -- loom-mario\n<span class="prompt">$</span> loom-mario</pre>'
                            ),
                            class_="cli-box",
                        ),
                        data_tab_group="install",
                        data_tab_id="unix",
                        class_="tab-pane active",
                    ),
                    div(
                        div(
                            div(
                                span("TERMINAL VERSION"),
                                div(
                                    span("runs anywhere", class_="badge"),
                                    tag(
                                        "button",
                                        "[COPY]",
                                        class_="pixel-btn copy-btn",
                                        data_cmd="irm https://raw.githubusercontent.com/ujjwalvivek/loom/main/examples/scripts/install.ps1 | iex",
                                    ),
                                    class_="cli-actions",
                                ),
                                class_="cli-header",
                            ),
                            RawHtml(
                                '<pre class="language-bash has-highlight"><span class="prompt">&gt;</span> irm <a href="https://raw.githubusercontent.com/ujjwalvivek/loom/main/examples/scripts/install.ps1" target="_blank" class="code-link">https://raw.githubusercontent.com/ujjwalvivek/loom/main/examples/scripts/install.ps1</a> | iex\n<span class="prompt">&gt;</span> loom-mario-term</pre>'
                            ),
                            class_="cli-box",
                        ),
                        div(
                            div(
                                span("GUI VERSION"),
                                div(
                                    span("needs OpenGL 3.3+", class_="badge badge-opengl"),
                                    a("needs GLFW", href="https://www.glfw.org/", target="_blank", class_="badge badge-glfw"),
                                    tag(
                                        "button",
                                        "[COPY]",
                                        class_="pixel-btn copy-btn",
                                        data_cmd='$Binary="loom-mario"; irm https://raw.githubusercontent.com/ujjwalvivek/loom/main/examples/scripts/install.ps1 | iex',
                                    ),
                                    class_="cli-actions",
                                ),
                                class_="cli-header",
                            ),
                            RawHtml(
                                '<pre class="language-bash has-highlight"><span class="prompt">&gt;</span> $Binary="loom-mario"; irm <a href="https://raw.githubusercontent.com/ujjwalvivek/loom/main/examples/scripts/install.ps1" target="_blank" class="code-link">https://raw.githubusercontent.com/ujjwalvivek/loom/main/examples/scripts/install.ps1</a> | iex\n<span class="prompt">&gt;</span> loom-mario</pre>'
                            ),
                            class_="cli-box",
                        ),
                        data_tab_group="install",
                        data_tab_id="windows",
                        class_="tab-pane",
                    ),
                    div(
                        div(
                            div(
                                span("GO INSTALL"),
                                div(
                                    tag(
                                        "button",
                                        "[COPY]",
                                        class_="pixel-btn copy-btn",
                                        data_cmd="go install github.com/ujjwalvivek/loom/examples/mario-term@latest\ngo install github.com/ujjwalvivek/loom/examples/mario@latest",
                                    ),
                                    class_="cli-actions",
                                ),
                                class_="cli-header",
                            ),
                            RawHtml(
                                '<pre class="language-bash has-highlight"><span class="prompt">$</span> go install <a href="https://github.com/ujjwalvivek/loom/tree/main/examples/mario-term" target="_blank" class="code-link">github.com/ujjwalvivek/loom/examples/mario-term</a>@latest\n<span class="prompt">$</span> go install <a href="https://github.com/ujjwalvivek/loom/tree/main/examples/mario" target="_blank" class="code-link">github.com/ujjwalvivek/loom/examples/mario</a>@latest</pre>'
                            ),
                            class_="cli-box",
                        ),
                        div(
                            div(
                                span("INSTALL FROM SOURCE"),
                                div(
                                    tag(
                                        "button",
                                        "[COPY]",
                                        class_="pixel-btn copy-btn",
                                        data_cmd="git clone https://github.com/ujjwalvivek/loom.git\ncd loom\ngo build -o loom-mario-term ./examples/mario-term\n./loom-mario-term",
                                    ),
                                    class_="cli-actions",
                                ),
                                class_="cli-header",
                            ),
                            RawHtml(
                                '<pre class="language-bash has-highlight"><span class="prompt">$</span> git clone <a href="https://github.com/ujjwalvivek/loom" target="_blank" class="code-link">https://github.com/ujjwalvivek/loom.git</a> && cd loom\n<span class="prompt">$</span> go build -o loom-mario-term <a href="https://github.com/ujjwalvivek/loom/tree/main/examples/mario-term" target="_blank" class="code-link">./examples/mario-term</a>\n<span class="prompt">$</span> ./loom-mario-term</pre>'
                            ),
                            class_="cli-box",
                        ),
                        data_tab_group="install",
                        data_tab_id="manual",
                        class_="tab-pane",
                    ),
                    class_="tab-content-wrapper",
                    data_tab_group="install",
                ),
                class_="section-card",
            ),
            section(
                h2(*h2_args, class_="section-title"),
                downloads_grid,
                class_="section-card",
            ),
            footer(
                div(
                    span("[MIT LICENSE]", class_="footer-badge"),
                    span("[BUILT WITH GO]", class_="footer-badge"),
                    span("[(c) 2026]", class_="footer-badge"),
                    class_="footer-badges",
                ),
                class_="footer",
            ),
            class_="container",
        ),
        title="Loom Mario - Download",
        description="Download the Loom Engine Mario demo.",
        theme=None,
        assets=[
            asset(ASSETS / "index.css", "assets/index.css"),
        ],
        stylesheets=["assets/index.css"],
        scripts=[],
        behaviors=[canvas_background()],
        head=[
            tag("link", rel="icon", type="image/png", href="favicon.png"),
            tag(
                "script",
                RawHtml(r"""
                    (function() {
                        function translateColor(color) {
                            if (typeof color !== "string") return color;
                            if (!document.documentElement.classList.contains("light-mode")) return color;
                            let lower = color.toLowerCase().trim();
                            if (lower.includes("#00aaff") || lower.includes("0, 170, 255")) return color.replace(/#00aaff/gi, "#0d6efd").replace(/0,\s*170,\s*255/g, "13, 110, 253");
                            if (lower.includes("#003566") || lower.includes("0, 53, 102")) return color.replace(/#003566/gi, "#a9d4ff").replace(/0,\s*53,\s*102/g, "169, 212, 255");
                            if (lower.includes("#00ffff") || lower.includes("0, 255, 255")) return color.replace(/#00ffff/gi, "#17a2b8").replace(/0,\s*255,\s*255/g, "23, 162, 184");
                            if (lower.includes("#0b132b") || lower.includes("11, 19, 43")) return color.replace(/#0b132b/gi, "#f0f8ff").replace(/11,\s*19,\s*43/g, "240, 248, 255");
                            if (lower.includes("#050914") || lower.includes("5, 9, 20")) return color.replace(/#050914/gi, "#e6f2ff").replace(/5,\s*9,\s*20/g, "230, 242, 255");
                            return color;
                        }
                        const descFill = Object.getOwnPropertyDescriptor(CanvasRenderingContext2D.prototype, "fillStyle");
                        const originalSetFill = descFill.set;
                        Object.defineProperty(CanvasRenderingContext2D.prototype, "fillStyle", {
                            ...descFill,
                            set(val) { originalSetFill.call(this, translateColor(val)); }
                        });
                        const descStroke = Object.getOwnPropertyDescriptor(CanvasRenderingContext2D.prototype, "strokeStyle");
                        const originalSetStroke = descStroke.set;
                        Object.defineProperty(CanvasRenderingContext2D.prototype, "strokeStyle", {
                            ...descStroke,
                            set(val) { originalSetStroke.call(this, translateColor(val)); }
                        });
                        const descShadow = Object.getOwnPropertyDescriptor(CanvasRenderingContext2D.prototype, "shadowColor");
                        const originalSetShadow = descShadow.set;
                        Object.defineProperty(CanvasRenderingContext2D.prototype, "shadowColor", {
                            ...descShadow,
                            set(val) { originalSetShadow.call(this, translateColor(val)); }
                        });
                        const originalAddColorStop = CanvasGradient.prototype.addColorStop;
                        CanvasGradient.prototype.addColorStop = function (offset, color) { return originalAddColorStop.call(this, offset, translateColor(color)); };
                    })();
                """),
            ),
            tag(
                "script",
                RawHtml(r"""
                    window.addEventListener("DOMContentLoaded", () => {
                        document.querySelectorAll("[data-tab-target]").forEach((btn) => {
                            btn.addEventListener("click", () => {
                                const group = btn.dataset.tabGroup;
                                const target = btn.dataset.tabTarget;
                                const activeBtn = document.querySelector(`[data-tab-group="${group}"][data-tab-target].active`);
                                if (activeBtn === btn) return;
                                document.querySelectorAll(`[data-tab-group="${group}"][data-tab-target]`).forEach((x) => x.classList.remove("active"));
                                document.querySelectorAll(`[data-tab-group="${group}"][data-tab-id]`).forEach((x) => x.classList.remove("active"));
                                btn.classList.add("active");
                                const targetPane = document.querySelector(`[data-tab-group="${group}"][data-tab-id="${target}"]`);
                                if (targetPane) { targetPane.classList.add("active"); }
                            });
                        });
                        document.querySelectorAll(".copy-btn").forEach((b) =>
                            b.addEventListener("click", () => {
                                navigator.clipboard.writeText(b.dataset.cmd);
                                let o = b.textContent;
                                b.textContent = "[COPIED!]";
                                setTimeout(() => (b.textContent = o), 1500);
                            })
                        );
                        (function () {
                            let btn = document.getElementById("theme-toggle"),
                                html = document.documentElement;
                            let badge = document.querySelector(".github-badge-link img");
                            if (!btn) return;
                            let dBadge = "https://echopoint.ujjwalvivek.com/svg/badges/stars?bg=111c44&badgeColor=003566&textColor=f8fafc&border=0066aa&borderWidth=2&rx=0&px=6&py=4&repo=loom&logo=github";
                            let lBadge = "https://echopoint.ujjwalvivek.com/svg/badges/stars?bg=d0e8ff&badgeColor=a9d4ff&textColor=0f172a&border=0a58ca&borderWidth=2&rx=0&px=6&py=4&repo=loom&logo=github";
                            btn.addEventListener("click", () => {
                                let isLight = html.classList.toggle("light-mode");
                                btn.textContent = isLight ? "[DARK]" : "[LIGHT]";
                                if (badge) badge.src = isLight ? lBadge : dBadge;
                            });
                        })();
                    });
                """),
            ),
        ],
    )
