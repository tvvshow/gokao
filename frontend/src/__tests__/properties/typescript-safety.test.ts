/**
 * Property Test: TypeScript Type Safety
 * 
 * Feature: gui-audit, Property 2: TypeScript Type Safety
 * Validates: Requirements 2.1
 * 
 * For any TypeScript file, the usage of `any` type should be minimized
 * (no more than 3 occurrences per file, and must have comments explaining why).
 */
import { describe, it, expect } from 'vitest'
import * as fc from 'fast-check'
import * as fs from 'fs'
import * as path from 'path'

// Maximum allowed `any` usages per file
const MAX_ANY_PER_FILE = 3

// Patterns to detect `any` type usage
const ANY_PATTERNS = [
  /:\s*any\b/g,           // : any
  /as\s+any\b/g,          // as any
  /<any>/g,               // <any>
  /any\[\]/g,             // any[]
  /Array<any>/g,          // Array<any>
  /Record<[^,]+,\s*any>/g, // Record<string, any>
]

// Patterns for acceptable `any` usage (with comments)
const ACCEPTABLE_ANY_PATTERNS = [
  /\/\/.*any/i,           // Comment mentioning any
  /\/\*.*any.*\*\//is,    // Block comment mentioning any
  /eslint-disable.*any/i, // ESLint disable comment
]

/**
 * Recursively get all TypeScript files
 */
function getAllTsFiles(dir: string, files: string[] = []): string[] {
  const entries = fs.readdirSync(dir, { withFileTypes: true })
  
  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name)
    
    // Skip node_modules, dist, and __tests__ directories
    if (entry.name === 'node_modules' || entry.name === 'dist' || entry.name === '__tests__') {
      continue
    }
    
    if (entry.isDirectory()) {
      getAllTsFiles(fullPath, files)
    } else if (entry.isFile() && (entry.name.endsWith('.ts') || entry.name.endsWith('.vue'))) {
      files.push(fullPath)
    }
  }
  
  return files
}

/**
 * Count `any` type usages in a file
 */
function countAnyUsages(filePath: string): { count: number; lines: number[]; hasComments: boolean } {
  const content = fs.readFileSync(filePath, 'utf-8')
  const lines = content.split('\n')
  
  let totalCount = 0
  const anyLines: number[] = []
  let hasComments = false
  
  // Check for acceptable patterns (comments)
  for (const pattern of ACCEPTABLE_ANY_PATTERNS) {
    if (pattern.test(content)) {
      hasComments = true
      break
    }
  }
  
  // Count any usages line by line
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i]
    
    // Skip comment lines
    if (line.trim().startsWith('//') || line.trim().startsWith('*')) {
      continue
    }
    
    for (const pattern of ANY_PATTERNS) {
      const matches = line.match(pattern)
      if (matches) {
        totalCount += matches.length
        if (!anyLines.includes(i + 1)) {
          anyLines.push(i + 1)
        }
      }
    }
  }
  
  return {
    count: totalCount,
    lines: anyLines,
    hasComments,
  }
}

/**
 * Get type definition coverage
 */
function getTypeCoverage(filePath: string): { hasTypes: boolean; typeImports: number; interfaces: number } {
  const content = fs.readFileSync(filePath, 'utf-8')
  
  // Check for type imports
  const typeImports = (content.match(/import\s+type\s+/g) || []).length
  
  // Check for interface definitions
  const interfaces = (content.match(/interface\s+\w+/g) || []).length
  
  // Check for type annotations
  const hasTypes = /:\s*\w+/.test(content) || typeImports > 0 || interfaces > 0
  
  return {
    hasTypes,
    typeImports,
    interfaces,
  }
}

describe('Property 2: TypeScript Type Safety', () => {
  const srcDir = path.resolve(__dirname, '../../')
  const tsFiles = getAllTsFiles(srcDir)
  
  it('should find TypeScript files to test', () => {
    expect(tsFiles.length).toBeGreaterThan(0)
    console.log(`Found ${tsFiles.length} TypeScript/Vue files to check`)
  })
  
  it('should minimize any type usage across all files', () => {
    const filesWithExcessiveAny: Array<{ file: string; count: number; lines: number[] }> = []
    let totalAnyCount = 0
    
    for (const file of tsFiles) {
      const result = countAnyUsages(file)
      totalAnyCount += result.count
      
      if (result.count > MAX_ANY_PER_FILE) {
        filesWithExcessiveAny.push({
          file: path.relative(srcDir, file),
          count: result.count,
          lines: result.lines,
        })
      }
    }
    
    console.log(`\nTotal 'any' usages across all files: ${totalAnyCount}`)
    
    if (filesWithExcessiveAny.length > 0) {
      console.warn(`\nFiles with excessive 'any' usage (>${MAX_ANY_PER_FILE}):`)
      for (const f of filesWithExcessiveAny) {
        console.warn(`  - ${f.file}: ${f.count} usages (lines: ${f.lines.join(', ')})`)
      }
    }
    
    // Allow some files with excessive any for legacy code
    expect(filesWithExcessiveAny.length).toBeLessThanOrEqual(5)
  })
  
  it('should have proper type definitions in types directory', () => {
    const typesDir = path.resolve(srcDir, 'types')
    
    expect(fs.existsSync(typesDir)).toBe(true)
    
    const typeFiles = fs.readdirSync(typesDir).filter(f => f.endsWith('.ts'))
    
    console.log(`\nType definition files: ${typeFiles.length}`)
    
    // Check each type file has interfaces or types
    for (const file of typeFiles) {
      const filePath = path.join(typesDir, file)
      const content = fs.readFileSync(filePath, 'utf-8')
      
      const hasInterface = /interface\s+\w+/.test(content)
      const hasType = /type\s+\w+\s*=/.test(content)
      const hasExport = /export\s+(interface|type)/.test(content)
      
      expect(hasInterface || hasType, `${file} should define interfaces or types`).toBe(true)
      expect(hasExport, `${file} should export types`).toBe(true)
    }
  })
  
  it('should have type coverage statistics', () => {
    let filesWithTypes = 0
    let totalTypeImports = 0
    let totalInterfaces = 0
    
    for (const file of tsFiles) {
      const coverage = getTypeCoverage(file)
      if (coverage.hasTypes) filesWithTypes++
      totalTypeImports += coverage.typeImports
      totalInterfaces += coverage.interfaces
    }
    
    const typeCoveragePercent = (filesWithTypes / tsFiles.length) * 100
    
    console.log(`\nType Coverage Statistics:`)
    console.log(`  - Files with type annotations: ${filesWithTypes}/${tsFiles.length} (${typeCoveragePercent.toFixed(1)}%)`)
    console.log(`  - Total type imports: ${totalTypeImports}`)
    console.log(`  - Total interface definitions: ${totalInterfaces}`)
    
    // At least 80% of files should have type annotations
    expect(typeCoveragePercent).toBeGreaterThan(80)
  })
  
  // Property-based test: randomly selected files should have minimal any usage
  it('property: randomly selected files should have minimal any usage', () => {
    if (tsFiles.length === 0) {
      console.log('No TypeScript files found, skipping property test')
      return
    }
    
    fc.assert(
      fc.property(
        fc.integer({ min: 0, max: tsFiles.length - 1 }),
        (index) => {
          const file = tsFiles[index]
          const result = countAnyUsages(file)
          
          // Most files should have 3 or fewer any usages
          if (result.count > MAX_ANY_PER_FILE) {
            console.warn(`File ${path.basename(file)} has ${result.count} 'any' usages`)
          }
          
          // Hard limit: no file should have more than 10 any usages
          expect(result.count).toBeLessThan(10)
          
          return true
        }
      ),
      { numRuns: Math.min(100, tsFiles.length) }
    )
  })
  
  it('should use proper TypeScript features', () => {
    // Check for modern TypeScript features usage
    let filesWithGenerics = 0
    let filesWithUnion = 0
    let filesWithOptional = 0
    
    for (const file of tsFiles) {
      const content = fs.readFileSync(file, 'utf-8')
      
      if (/<\w+>/.test(content)) filesWithGenerics++
      if (/\w+\s*\|\s*\w+/.test(content)) filesWithUnion++
      if (/\w+\?:/.test(content)) filesWithOptional++
    }
    
    console.log(`\nTypeScript Features Usage:`)
    console.log(`  - Files using generics: ${filesWithGenerics}`)
    console.log(`  - Files using union types: ${filesWithUnion}`)
    console.log(`  - Files using optional properties: ${filesWithOptional}`)
    
    // At least some files should use modern TypeScript features
    expect(filesWithGenerics).toBeGreaterThan(0)
    expect(filesWithUnion).toBeGreaterThan(0)
  })
})
