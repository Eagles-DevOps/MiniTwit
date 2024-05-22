#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

markdown_files=$(ls *.md | grep -v 'template.md' | sort -V) # Sort the files numerically and ignore template.md file


command="pandoc --columns=3 --wrap=auto $markdown_files -s -o MSc_group_h.pdf"

# Infinite loop
while true
do
    # Execute the command
    $command
    
    # Optional: Add a sleep interval to avoid overwhelming the system
    # sleep 1
done
