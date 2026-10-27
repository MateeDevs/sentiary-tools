"""Build and validate the shared documentation for one deployed environment."""

import argparse
from pathlib import Path
import re
import subprocess
import sys

ENVIRONMENTS = {
    "production": ("https://docs.sentiary.com/", "main", "api.sentiary.com"),
    "staging": (
        "https://docs.staging.sentiary.coolify.matee.cz/",
        "staging",
        "api.staging.sentiary.coolify.matee.cz",
    ),
}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--environment", choices=ENVIRONMENTS, default="production")
    args = parser.parse_args()
    root = Path(__file__).resolve().parents[1]
    site_url, branch, api_host = ENVIRONMENTS[args.environment]
    config = (root / "docs/zensical.toml").read_text(encoding="utf-8")
    for key, value in {
        "site_url": site_url,
        "edit_uri": f"edit/{branch}/docs/content/",
    }.items():
        config, count = re.subn(
            rf"^{key} = .*?$", f'{key} = "{value}"', config, flags=re.MULTILINE
        )
        if count != 1:
            raise SystemExit(f"Expected exactly one {key} setting")
    generated = root / "docs/.zensical-build.toml"
    generated.write_text(config, encoding="utf-8")
    nginx = (root / "docs/nginx.conf").read_text(encoding="utf-8")
    (root / "docs/.nginx-build.conf").write_text(
        nginx.replace("api.sentiary.com", api_host), encoding="utf-8"
    )
    subprocess.run(
        ["zensical", "build", "--strict", "--config-file", str(generated)],
        cwd=root,
        check=True,
    )
    subprocess.run(
        [sys.executable, str(root / "scripts/check-docs-exports.py"),
         "--config-file", str(generated)],
        cwd=root,
        check=True,
    )


if __name__ == "__main__":
    main()
