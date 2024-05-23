#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

markdown_files=$(ls *.md | grep -v 'template.md' | sort -V) # Sort the files numerically and ignore template.md file
pandoc $markdown_files -s -o MSc_group_h.pdf
