const fs = require('fs');
const path = require('path');
const testDir = path.join(__dirname, 'tests');

// Function to process files in a directory recursively
function processDirectory(dirPath) {
  const files = fs.readdirSync(dirPath);
  
  for (const file of files) {
    const filePath = path.join(dirPath, file);
    const stats = fs.statSync(filePath);
    
    if (stats.isDirectory()) {
      processDirectory(filePath);
    } else if (stats.isFile() && file.endsWith('.js')) {
      fixOptionsInFile(filePath);
    }
  }
}

function fixOptionsInFile(filePath) {
  const content = fs.readFileSync(filePath, 'utf8');
  const optionsCount = (content.match(/export const options/g) || []).length;
  
  if (optionsCount <= 1) {
    return; // No duplication, skip this file
  }
  
  console.log(`Fixing ${path.relative(__dirname, filePath)}`);
  
  // Extract imports
  const importRegex = /^import.*?;$/gm;
  const imports = [];
  let match;
  while ((match = importRegex.exec(content)) !== null) {
    imports.push(match[0]);
  }
  
  // Extract unique imports
  const uniqueImports = [...new Set(imports)];
  
  // Extract options object 
  const optionsRegex = /export const options = \{\s+thresholds: \{\s+\/\/ Require 100% of checks to pass\s+'checks': \['rate==1\.0'\],\s+\},\s+\};/;
  const optionsMatch = content.match(optionsRegex);
  const options = optionsMatch ? optionsMatch[0] : '';
  
  // Extract rest of the file (after last import)
  const lastImportEnd = content.lastIndexOf(';', content.lastIndexOf('import')) + 1;
  const lastOptionsEnd = content.lastIndexOf(';', content.lastIndexOf('options')) + 1;
  const startOfCode = Math.max(lastImportEnd, lastOptionsEnd);
  
  let mainCode = content.substring(startOfCode);
  // Remove any remaining options declarations
  mainCode = mainCode.replace(/export const options[\s\S]*?};/g, '').trim();
  
  // Build the fixed content
  const fixedContent = uniqueImports.join('\n') + '\n\n' + options + '\n\n' + mainCode;
  
  // Write the fixed content back
  fs.writeFileSync(filePath, fixedContent);
}

// Start processing from the tests directory
processDirectory(testDir);
console.log('All files fixed!');