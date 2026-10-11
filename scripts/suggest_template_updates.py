#!/usr/bin/env python3
"""
Scans templates/ folder and the existing backend sqlite DB and prints SQL UPDATE statements
to align templates.source_path, templates.demo_path, and templates.thumbnail_path with actual
HTML files found. This is a safe, non-destructive helper: it only prints SQL statements for
manual review and execution.

Usage:
  python3 scripts/suggest_template_updates.py

Note: The script assumes the DB file is at backend/data/sukamoto.db and templates are under
./templates. Adjust paths in the script if your layout differs.
"""
import os
import sqlite3
import sys

ROOT = os.path.dirname(os.path.dirname(__file__))
DB_PATH = os.path.join(ROOT, 'backend', 'data', 'sukamoto.db')
TEMPLATES_DIR = os.path.join(ROOT, 'templates')

if not os.path.exists(DB_PATH):
    print(f"Database not found at {DB_PATH}")
    sys.exit(1)

if not os.path.isdir(TEMPLATES_DIR):
    print(f"Templates folder not found at {TEMPLATES_DIR}")
    sys.exit(1)

# Collect HTML files under templates/
html_files = []
for dirpath, dirs, files in os.walk(TEMPLATES_DIR):
    for f in files:
        if f.lower().endswith('.html'):
            rel = os.path.relpath(os.path.join(dirpath, f), ROOT)
            html_files.append(rel.replace('\\', '/'))

# Connect to DB and read existing templates
conn = sqlite3.connect(DB_PATH)
c = conn.cursor()
c.execute('SELECT id, slug, package_slug, source_path, demo_path, thumbnail_path FROM templates')
rows = c.fetchall()

print('-- Found HTML files under templates/:')
for p in html_files:
    print(f"--   {p}")
print('\n-- Suggested SQL updates (review before applying):\n')

def find_best_match(slug, package):
    # Heuristics: prefer files that contain slug token, or package dir, or nearest size
    candidates = []
    for f in html_files:
        name = os.path.basename(f).lower()
        path = f.lower()
        score = 0
        if slug.lower() in name or slug.lower() in path:
            score += 100
        if package and package.lower() in path:
            score += 20
        # prefer shorter filenames (likely direct template file)
        score += max(0, 10 - len(name))
        candidates.append((score, f))
    candidates.sort(reverse=True)
    return candidates[0][1] if candidates and candidates[0][0] > 0 else None

for id_, slug, pkg, source, demo, thumb in rows:
    match = find_best_match(slug, pkg)
    if match:
        # Build recommended paths
        new_source = match
        new_demo = f"/demo/{slug}"
        # keep thumbnail path if it exists; otherwise try to guess under assets
        new_thumb = thumb if thumb and thumb.strip() else f"/assets/templates/{slug}/thumb.jpg"
        print(f"-- Template id={id_} slug={slug}")
        print("-- Current:")
        print(f"--   source_path={source}")
        print(f"--   demo_path={demo}")
        print(f"--   thumbnail_path={thumb}\n")
        print("-- Recommended UPDATE:")
        print("BEGIN TRANSACTION;")
        safe_source = new_source.replace("'","''")
        safe_demo = new_demo.replace("'","''")
        safe_thumb = new_thumb.replace("'","''")
        print(f"UPDATE templates SET source_path = '{safe_source}', demo_path = '{safe_demo}', thumbnail_path = '{safe_thumb}', updated_at = datetime('now') WHERE id = {id_};")
        print("COMMIT;\n")
    else:
        print(f"-- No good match found for template id={id_} slug={slug}; manual action required.\n")

conn.close()
print('-- End of suggestions')
