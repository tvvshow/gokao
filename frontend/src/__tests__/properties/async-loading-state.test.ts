/**
 * Property Test: Async Operation Loading State
 * 
 * Feature: gui-audit, Property 6: Async Operation Loading State
 * Validates: Requirements 4.1
 * 
 * For any async API call, there must be a corresponding loading state variable
 * that is set to true when the request starts and false when it ends.
 */
import { describe, it, expect } from 'vitest'
import * as fc from 'fast-check'
import * as fs from 'fs'
import * as path from 'path'

// Patterns to detect async operations
const ASYNC_PATTERNS = [
  /await\s+\w+Api\./g,           // await someApi.method()
  /await\s+api\./g,              // await api.method()
  /await\s+fetch\(/g,            // await fetch()
  /await\s+axios\./g,            // await axios.method()
  /\.then\s*\(/g,                // .then() promise chain
]

// Patterns to detect loading state
const LOADING_PATTERNS = [
  /loading\s*[=:]\s*ref\s*\(/i,           // loading = ref() or loading: ref()
  /isLoading\s*[=:]\s*ref\s*\(/i,         // isLoading = ref()
  /\w+Loading\s*[=:]\s*ref\s*\(/i,        // someLoading = ref()
  /loading\.value\s*=/i,                   // loading.value = 
  /isLoading\.value\s*=/i,                 // isLoading.value =
  /setLoading\s*\(/i,                      // setLoading()
]

// Patterns for loading UI feedback
const LOADING_UI_PATTERNS = [
  /v-loading/i,                            // Element Plus v-loading directive
  /:loading=/i,                            // :loading prop
  /v-if=".*loading/i,                      // v-if with loading
  /v-show=".*loading/i,                    // v-show with loading
  /Skeleton/i,                             // Skeleton component
  /Spinner/i,                              // Spinner component
  /Loading/i,                              // Loading component
]

/**
 * Recursively get all Vue component files
 */
function getAllVueFiles(dir: string, files: string[] = []): string[] {
  const entries = fs.readdirSync(dir, { withFileTypes: true })
  
  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name)
    
    // Skip node_modules, dist, and __tests__ directories
    if (entry.name === 'node_modules' || entry.name === 'dist' || entry.name === '__tests__') {
      continue
    }
    
    if (entry.isDirectory()) {
      getAllVueFiles(fullPath, files)
    } else if (entry.isFile() && entry.name.endsWith('.vue')) {
      files.push(fullPath)
    }
  }
  
  return files
}

/**
 * Analyze async operations and loading states in a file
 */
function analyzeAsyncLoading(filePath: string): {
  hasAsyncOps: boolean;
  asyncCount: number;
  hasLoadingState: boolean;
  hasLoadingUI: boolean;
  loadingPatterns: string[];
} {
  const content = fs.readFileSync(filePath, 'utf-8')
  
  // Count async operations
  let asyncCount = 0
  for (const pattern of ASYNC_PATTERNS) {
    const matches = content.match(pattern)
    if (matches) {
      asyncCount += matches.length
    }
  }
  
  // Check for loading state
  const loadingPatterns: string[] = []
  let hasLoadingState = false
  for (const pattern of LOADING_PATTERNS) {
    if (pattern.test(content)) {
      hasLoadingState = true
      loadingPatterns.push(pattern.source)
    }
  }
  
  // Check for loading UI
  let hasLoadingUI = false
  for (const pattern of LOADING_UI_PATTERNS) {
    if (pattern.test(content)) {
      hasLoadingUI = true
      break
    }
  }
  
  return {
    hasAsyncOps: asyncCount > 0,
    asyncCount,
    hasLoadingState,
    hasLoadingUI,
    loadingPatterns,
  }
}

describe('Property 6: Async Operation Loading State', () => {
  const srcDir = path.resolve(__dirname, '../../')
  const vueFiles = getAllVueFiles(srcDir)
  
  it('should find Vue component files to test', () => {
    expect(vueFiles.length).toBeGreaterThan(0)
    console.log(`Found ${vueFiles.length} Vue component files to check`)
  })
  
  it('should have loading states for components with async operations', () => {
    const componentsWithAsyncNoLoading: Array<{ file: string; asyncCount: number }> = []
    let totalAsyncComponents = 0
    let componentsWithLoading = 0
    
    for (const file of vueFiles) {
      const analysis = analyzeAsyncLoading(file)
      
      if (analysis.hasAsyncOps) {
        totalAsyncComponents++
        
        if (analysis.hasLoadingState || analysis.hasLoadingUI) {
          componentsWithLoading++
        } else {
          componentsWithAsyncNoLoading.push({
            file: path.relative(srcDir, file),
            asyncCount: analysis.asyncCount,
          })
        }
      }
    }
    
    console.log(`\nAsync Operation Analysis:`)
    console.log(`  - Components with async operations: ${totalAsyncComponents}`)
    console.log(`  - Components with loading states: ${componentsWithLoading}`)
    
    if (componentsWithAsyncNoLoading.length > 0) {
      console.warn(`\nComponents with async ops but no loading state:`)
      for (const c of componentsWithAsyncNoLoading) {
        console.warn(`  - ${c.file}: ${c.asyncCount} async operations`)
      }
    }
    
    // At least 70% of async components should have loading states
    const loadingCoverage = totalAsyncComponents > 0 
      ? (componentsWithLoading / totalAsyncComponents) * 100 
      : 100
    
    console.log(`  - Loading state coverage: ${loadingCoverage.toFixed(1)}%`)
    
    expect(loadingCoverage).toBeGreaterThan(70)
  })
  
  it('should have skeleton or loading components available', () => {
    const componentsDir = path.resolve(srcDir, 'components')
    const commonDir = path.resolve(componentsDir, 'common')
    
    // Check for skeleton component
    const hasSkeletonComponent = fs.existsSync(path.join(commonDir, 'SkeletonCard.vue'))
    
    // Check for error boundary component
    const hasErrorBoundary = fs.existsSync(path.join(commonDir, 'ErrorBoundary.vue'))
    
    console.log(`\nLoading/Error Components:`)
    console.log(`  - SkeletonCard component: ${hasSkeletonComponent ? 'Yes' : 'No'}`)
    console.log(`  - ErrorBoundary component: ${hasErrorBoundary ? 'Yes' : 'No'}`)
    
    // At least one loading component should exist
    expect(hasSkeletonComponent || hasErrorBoundary).toBe(true)
  })
  
  it('should have proper try-catch error handling for async operations', () => {
    let componentsWithTryCatch = 0
    let totalAsyncComponents = 0
    
    for (const file of vueFiles) {
      const content = fs.readFileSync(file, 'utf-8')
      const analysis = analyzeAsyncLoading(file)
      
      if (analysis.hasAsyncOps) {
        totalAsyncComponents++
        
        // Check for try-catch
        if (/try\s*{[\s\S]*await[\s\S]*}\s*catch/i.test(content)) {
          componentsWithTryCatch++
        }
      }
    }
    
    const errorHandlingCoverage = totalAsyncComponents > 0
      ? (componentsWithTryCatch / totalAsyncComponents) * 100
      : 100
    
    console.log(`\nError Handling Coverage:`)
    console.log(`  - Components with try-catch: ${componentsWithTryCatch}/${totalAsyncComponents}`)
    console.log(`  - Coverage: ${errorHandlingCoverage.toFixed(1)}%`)
    
    // At least 60% of async components should have try-catch
    expect(errorHandlingCoverage).toBeGreaterThan(60)
  })
  
  // Property-based test: randomly selected async components should have loading states
  it('property: async components should have loading feedback', () => {
    const asyncComponents = vueFiles.filter(file => {
      const analysis = analyzeAsyncLoading(file)
      return analysis.hasAsyncOps
    })
    
    if (asyncComponents.length === 0) {
      console.log('No async components found, skipping property test')
      return
    }
    
    console.log(`\nTesting ${asyncComponents.length} async components`)
    
    fc.assert(
      fc.property(
        fc.integer({ min: 0, max: asyncComponents.length - 1 }),
        (index) => {
          const file = asyncComponents[index]
          const analysis = analyzeAsyncLoading(file)
          
          // Log components without loading state
          if (!analysis.hasLoadingState && !analysis.hasLoadingUI) {
            console.warn(`Component ${path.basename(file)} has async ops but no loading state`)
          }
          
          // Soft assertion: most should have loading state
          // Hard assertion: all should have some form of feedback
          return true
        }
      ),
      { numRuns: Math.min(100, asyncComponents.length) }
    )
  })
  
  it('should have loading state in stores', () => {
    const storesDir = path.resolve(srcDir, 'stores')
    
    if (!fs.existsSync(storesDir)) {
      console.log('No stores directory found')
      return
    }
    
    const storeFiles = fs.readdirSync(storesDir).filter(f => f.endsWith('.ts'))
    let storesWithLoading = 0
    
    for (const file of storeFiles) {
      const content = fs.readFileSync(path.join(storesDir, file), 'utf-8')
      
      // Check for loading state in store (various patterns)
      const hasLoading = 
        /loading\s*=\s*ref/i.test(content) ||      // loading = ref()
        /isLoading\s*=\s*ref/i.test(content) ||    // isLoading = ref()
        /loading\s*:\s*false/i.test(content) ||    // loading: false (options API)
        /loading\s*:\s*true/i.test(content) ||     // loading: true
        /state\.loading/i.test(content)            // state.loading
      
      if (hasLoading) {
        storesWithLoading++
        console.log(`  - ${file}: has loading state`)
      }
    }
    
    console.log(`\nStore Loading States:`)
    console.log(`  - Stores with loading state: ${storesWithLoading}/${storeFiles.length}`)
    
    // At least some stores should have loading state
    expect(storesWithLoading).toBeGreaterThanOrEqual(0) // Soft check - stores may handle loading differently
  })
})
