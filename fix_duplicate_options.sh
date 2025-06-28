#!/bin/bash

# Script to fix files with duplicate options declarations
# This normalizes the test files to have imports first, then options

for js_file in $(grep -c "export const options" tests/**/*.js | grep -v ":1$" | cut -d ":" -f 1); do
  echo "Fixing $js_file"
  
  # Extract all import statements
  imports=$(grep "^import" "$js_file" | sort | uniq)
  
  # Get the options definition
  options=$(grep -A4 "export const options" "$js_file" | head -5 | sort | uniq | grep -v "^--")
  
  # Get the rest of the file after all imports and options
  code=$(sed -n '/export default function/,$p' "$js_file")
  
  # Create the fixed file
  echo "$imports" > "${js_file}.fixed"
  echo "" >> "${js_file}.fixed"
  echo "$options" >> "${js_file}.fixed"
  echo "" >> "${js_file}.fixed"
  echo "$code" >> "${js_file}.fixed"
  
  # Replace the original file
  mv "${js_file}.fixed" "$js_file"
done

echo "Done fixing duplicate options!"