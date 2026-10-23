#!/usr/bin/env python3
"""Validate .github/ISSUE_TEMPLATE/*.yml: valid YAML and required keys.

config.yml must hold contact_links entries with name, url and about.
Every other file is an issue form and needs name, description and a
non-empty body whose items have a known type and the attributes it needs.
"""
import glob
import os
import sys

import yaml

FORM_TYPES = {"markdown", "textarea", "input", "dropdown", "checkboxes"}


def nonempty_str(v):
    return isinstance(v, str) and v.strip() != ""


def check_config(doc):
    errs = []
    if not isinstance(doc, dict):
        return ["top level must be a mapping"]
    if "blank_issues_enabled" in doc and not isinstance(doc["blank_issues_enabled"], bool):
        errs.append("blank_issues_enabled must be true or false")
    links = doc.get("contact_links", [])
    if not isinstance(links, list):
        return errs + ["contact_links must be a list"]
    for i, link in enumerate(links):
        if not isinstance(link, dict):
            errs.append(f"contact_links[{i}] must be a mapping")
            continue
        for key in ("name", "url", "about"):
            if not nonempty_str(link.get(key)):
                errs.append(f"contact_links[{i}] is missing {key}")
    return errs


def check_form(doc):
    errs = []
    if not isinstance(doc, dict):
        return ["top level must be a mapping"]
    for key in ("name", "description"):
        if not nonempty_str(doc.get(key)):
            errs.append(f"missing {key}")
    body = doc.get("body")
    if not isinstance(body, list) or not body:
        return errs + ["body must be a non-empty list"]
    ids = set()
    for i, item in enumerate(body):
        where = f"body[{i}]"
        if not isinstance(item, dict):
            errs.append(f"{where} must be a mapping")
            continue
        kind = item.get("type")
        if kind not in FORM_TYPES:
            errs.append(f"{where} has unknown type {kind!r}")
            continue
        attrs = item.get("attributes")
        if not isinstance(attrs, dict):
            errs.append(f"{where} is missing attributes")
            continue
        if kind == "markdown":
            if not nonempty_str(attrs.get("value")):
                errs.append(f"{where} markdown is missing attributes.value")
            continue
        if not nonempty_str(attrs.get("label")):
            errs.append(f"{where} is missing attributes.label")
        item_id = item.get("id")
        if item_id is not None:
            if item_id in ids:
                errs.append(f"{where} duplicates id {item_id!r}")
            ids.add(item_id)
        if kind in ("dropdown", "checkboxes"):
            opts = attrs.get("options")
            if not isinstance(opts, list) or not opts:
                errs.append(f"{where} needs a non-empty attributes.options list")
            elif kind == "checkboxes":
                for j, opt in enumerate(opts):
                    if not isinstance(opt, dict) or not nonempty_str(opt.get("label")):
                        errs.append(f"{where} options[{j}] is missing label")
    return errs


def main():
    root = sys.argv[1] if len(sys.argv) > 1 else ".github/ISSUE_TEMPLATE"
    files = sorted(glob.glob(os.path.join(root, "*.yml")) + glob.glob(os.path.join(root, "*.yaml")))
    if not files:
        print(f"::error::no issue templates found under {root}")
        return 1
    failed = 0
    for path in files:
        try:
            with open(path, encoding="utf-8") as fh:
                doc = yaml.safe_load(fh)
        except yaml.YAMLError as exc:
            mark = getattr(exc, "problem_mark", None)
            line = f",line={mark.line + 1}" if mark else ""
            print(f"::error file={path}{line}::invalid YAML: {getattr(exc, 'problem', exc)}")
            failed += 1
            continue
        name = os.path.basename(path)
        errs = check_config(doc) if name in ("config.yml", "config.yaml") else check_form(doc)
        for err in errs:
            print(f"::error file={path}::{err}")
        if errs:
            failed += 1
        else:
            print(f"ok  {path}")
    if failed:
        print(f"{failed} of {len(files)} issue template(s) failed validation")
        return 1
    print(f"all {len(files)} issue template(s) are valid")
    return 0


if __name__ == "__main__":
    sys.exit(main())
