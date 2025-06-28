#!/bin/bash

# Script to add threshold options to all k6 test files
# This ensures that any check failure will cause k6 to exit with code 1

for js_file in $(find tests -name "*.js" -not -path "tests/memory/*"); do
  echo "Processing $js_file"
  
  # Check if the file already has options defined
  if grep -q "export const options" "$js_file"; then
    echo "  Options already defined, skipping"
    continue
  fi
  
  # Add options block after the imports but before the first function
  sed -i.bak '
    /^import.*k6/a\
export const options = {\
    thresholds: {\
        // Require 100% of checks to pass\
        '\''checks'\'': ['\''rate==1.0'\''],\
    },\
};
  ' "$js_file"
  
  # Remove backup file
  rm -f "${js_file}.bak"
done

echo "Done!"