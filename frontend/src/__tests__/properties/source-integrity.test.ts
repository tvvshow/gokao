/**
 * Property Test: Source Code Integrity
 * 
 * Feature: gui-audit, Property 1: Source Code Integrity
 * Validates: Requirements 2.3
 * 
 * For any source code file (.ts, .vue, .js), it should not contain
 * Git merge conflict markers
 */
import { describe, it, expect } from 'vitest'
import * as fc from 'fast-check'
import * as fs from 'fs'
import * as path from 'path'

// Git conflict markers to detect (match real marker lines only)
const CONFLICT_PATTERNS = [
  /^<<<<<<<\s.+$/m,
  /^=======$/m,
  /^>>>>>>>\s.+$/m,
]

// Source file extensions to check
const SOURCE_EXTENSIONS = ['.ts', '.vue', '.js', '.tsx', '.jsx']

/**
 * Recursively get all source files in a directory
 */
function getAllSourceFiles(dir: string, files: string[] = []): string[] {
  const entries = fs.readdirSync(dir, { withFileTypes: true })
  
  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name)
    
    // Skip node_modules, dist, and __tests__ directories
    if (entry.name === 'node_modules' || entry.name === 'dist' || entry.name === '__tests__') {
      continue
    }
    
    if (entry.isDirectory()) {
      getAllSourceFiles(fullPath, files)
    } else if (entry.isFile()) {
      const ext = path.extname(entry.name)
      if (SOURCE_EXTENSIONS.includes(ext)) {
        files.push(fullPath)
      }
    }
  }
  
  return files
}

/**
 * Check if a file contains Git conflict markers
 */
function hasGitConflictMarkers(filePath: string): { hasConflict: boolean; markers: string[] } {
  const content = fs.readFileSync(filePath, 'utf-8')
  const foundMarkers: string[] = []
  
  for (const pattern of CONFLICT_PATTERNS) {
    const match = content.match(pattern)
    if (match) {
      foundMarkers.push(match[0])
    }
  }
  
  return {
    hasConflict: foundMarkers.length > 0,
    markers: foundMarkers,
  }
}

describe('Property 1: Source Code Integrity', () => {
  const srcDir = path.resolve(__dirname, '../../')
  const sourceFiles = getAllSourceFiles(srcDir)
  
  it('should find source files to test', () => {
    expect(sourceFiles.length).toBeGreaterThan(0)
    console.log(`Found ${sourceFiles.length} source files to check`)
  })
  
  it('should not contain Git merge conflict markers in any source file', () => {
    const filesWithConflicts: Array<{ file: string; markers: string[] }> = []
    
    for (const file of sourceFiles) {
      const result = hasGitConflictMarkers(file)
      if (result.hasConflict) {
        filesWithConflicts.push({
          file: path.relative(srcDir, file),
          markers: result.markers,
        })
      }
    }
    
    if (filesWithConflicts.length > 0) {
      const errorMessage = filesWithConflicts
        .map(f => `  - ${f.file}: found markers [${f.markers.join(', ')}]`)
        .join('\n')
      
      expect.fail(
        `Found Git conflict markers in ${filesWithConflicts.length} file(s):\n${errorMessage}`
      )
    }
    
    expect(filesWithConflicts).toHaveLength(0)
  })
  
  // Property-based test: For any randomly selected source file,
  // it should not contain conflict markers
  it('property: randomly selected source files should not contain conflict markers', () => {
    if (sourceFiles.length === 0) {
      console.log('No source files found, skipping property test')
      return
    }
    
    fc.assert(
      fc.property(
        fc.integer({ min: 0, max: sourceFiles.length - 1 }),
        (index) => {
          const file = sourceFiles[index]
          const result = hasGitConflictMarkers(file)
          
          if (result.hasConflict) {
            return false
          }
          return true
        }
      ),
      { numRuns: Math.min(100, sourceFiles.length) }
    )
  })
})
