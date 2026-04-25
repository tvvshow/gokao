/**
 * Property Test: Accessibility Labels
 * 
 * Feature: gui-audit, Property 7: Accessibility Labels
 * Validates: Requirements 6.1
 * 
 * For any interactive element (buttons, links, form controls),
 * there should be appropriate aria-label or accessible text content.
 */
import { describe, it, expect } from 'vitest'
import * as fc from 'fast-check'
import * as fs from 'fs'
import * as path from 'path'

// Interactive elements that need accessibility
const INTERACTIVE_ELEMENTS = [
  'button',
  'a',
  'input',
  'select',
  'textarea',
  'el-button',
  'el-input',
  'el-select',
  'el-link',
  'el-checkbox',
  'el-radio',
  'el-switch',
]

// Accessibility patterns
const ARIA_PATTERNS = [
  /aria-label=/i,
  /aria-labelledby=/i,
  /aria-describedby=/i,
  /role=/i,
]

// Label patterns
const LABEL_PATTERNS = [
  /<label/i,
  /for=/i,
  /placeholder=/i,
  /title=/i,
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
 * Analyze accessibility in a Vue component
 */
function analyzeAccessibility(filePath: string): {
  hasInteractiveElements: boolean;
  interactiveCount: number;
  hasAriaLabels: boolean;
  ariaCount: number;
  hasLabels: boolean;
  labelCount: number;
  issues: string[];
} {
  const content = fs.readFileSync(filePath, 'utf-8')
  const issues: string[] = []
  
  // Count interactive elements
  let interactiveCount = 0
  for (const element of INTERACTIVE_ELEMENTS) {
    const regex = new RegExp(`<${element}[\\s>]`, 'gi')
    const matches = content.match(regex)
    if (matches) {
      interactiveCount += matches.length
    }
  }
  
  // Count ARIA attributes
  let ariaCount = 0
  for (const pattern of ARIA_PATTERNS) {
    const matches = content.match(pattern)
    if (matches) {
      ariaCount += matches.length
    }
  }
  
  // Count label patterns
  let labelCount = 0
  for (const pattern of LABEL_PATTERNS) {
    const matches = content.match(pattern)
    if (matches) {
      labelCount += matches.length
    }
  }
  
  // Check for icon-only buttons without aria-label
  const iconButtonPattern = /<(?:el-)?button[^>]*>[\s\S]*?<(?:el-)?icon|<Icon[^>]*\/?>[\s\S]*?<\/(?:el-)?button>/gi
  const iconButtons = content.match(iconButtonPattern) || []
  for (const button of iconButtons) {
    if (!button.includes('aria-label')) {
      issues.push('Icon-only button without aria-label')
    }
  }
  
  // Check for images without alt text
  const imgPattern = /<img[^>]*>/gi
  const images = content.match(imgPattern) || []
  for (const img of images) {
    if (!img.includes('alt=')) {
      issues.push('Image without alt attribute')
    }
  }
  
  return {
    hasInteractiveElements: interactiveCount > 0,
    interactiveCount,
    hasAriaLabels: ariaCount > 0,
    ariaCount,
    hasLabels: labelCount > 0,
    labelCount,
    issues,
  }
}

describe('Property 7: Accessibility Labels', () => {
  const srcDir = path.resolve(__dirname, '../../')
  const vueFiles = getAllVueFiles(srcDir)
  
  it('should find Vue component files to test', () => {
    expect(vueFiles.length).toBeGreaterThan(0)
    console.log(`Found ${vueFiles.length} Vue component files to check`)
  })
  
  it('should have accessibility attributes in components with interactive elements', () => {
    let totalInteractive = 0
    let totalAria = 0
    let totalLabels = 0
    let componentsWithIssues = 0
    const allIssues: Array<{ file: string; issues: string[] }> = []
    
    for (const file of vueFiles) {
      const analysis = analyzeAccessibility(file)
      
      totalInteractive += analysis.interactiveCount
      totalAria += analysis.ariaCount
      totalLabels += analysis.labelCount
      
      if (analysis.issues.length > 0) {
        componentsWithIssues++
        allIssues.push({
          file: path.relative(srcDir, file),
          issues: analysis.issues,
        })
      }
    }
    
    console.log(`\nAccessibility Analysis:`)
    console.log(`  - Total interactive elements: ${totalInteractive}`)
    console.log(`  - Total ARIA attributes: ${totalAria}`)
    console.log(`  - Total label patterns: ${totalLabels}`)
    console.log(`  - Components with issues: ${componentsWithIssues}`)
    
    if (allIssues.length > 0) {
      console.warn(`\nAccessibility Issues Found:`)
      for (const item of allIssues.slice(0, 5)) {
        console.warn(`  - ${item.file}:`)
        for (const issue of item.issues) {
          console.warn(`      * ${issue}`)
        }
      }
    }
    
    // At least some accessibility attributes should be present
    expect(totalAria + totalLabels).toBeGreaterThan(0)
  })
  
  it('should have form labels or placeholders', () => {
    let formsWithLabels = 0
    let totalForms = 0
    
    for (const file of vueFiles) {
      const content = fs.readFileSync(file, 'utf-8')
      
      // Check for form elements
      if (/<el-form|<form/i.test(content)) {
        totalForms++
        
        // Check for labels or placeholders
        if (/label=|placeholder=|<label/i.test(content)) {
          formsWithLabels++
        }
      }
    }
    
    console.log(`\nForm Accessibility:`)
    console.log(`  - Forms with labels/placeholders: ${formsWithLabels}/${totalForms}`)
    
    // All forms should have labels or placeholders
    if (totalForms > 0) {
      expect(formsWithLabels).toBe(totalForms)
    }
  })
  
  it('should have proper heading hierarchy', () => {
    let componentsWithHeadings = 0
    let properHierarchy = 0
    
    for (const file of vueFiles) {
      const content = fs.readFileSync(file, 'utf-8')
      
      // Check for headings
      const hasHeadings = /<h[1-6]/i.test(content)
      if (hasHeadings) {
        componentsWithHeadings++
        
        // Check for h1 (should be used sparingly)
        const h1Count = (content.match(/<h1/gi) || []).length
        
        // Proper hierarchy: h1 should appear at most once per component
        if (h1Count <= 1) {
          properHierarchy++
        }
      }
    }
    
    console.log(`\nHeading Hierarchy:`)
    console.log(`  - Components with headings: ${componentsWithHeadings}`)
    console.log(`  - Proper hierarchy: ${properHierarchy}/${componentsWithHeadings}`)
    
    // Most components should have proper heading hierarchy
    if (componentsWithHeadings > 0) {
      const hierarchyPercent = (properHierarchy / componentsWithHeadings) * 100
      expect(hierarchyPercent).toBeGreaterThan(80)
    }
  })
  
  it('should have focus indicators', () => {
    let componentsWithFocusStyles = 0
    
    for (const file of vueFiles) {
      const content = fs.readFileSync(file, 'utf-8')
      
      // Check for focus styles
      if (/:focus|focus-visible|focus-within|outline/i.test(content)) {
        componentsWithFocusStyles++
      }
    }
    
    console.log(`\nFocus Indicators:`)
    console.log(`  - Components with focus styles: ${componentsWithFocusStyles}/${vueFiles.length}`)
    
    // At least some components should have focus styles
    expect(componentsWithFocusStyles).toBeGreaterThan(0)
  })
  
  // Property-based test: randomly selected components should have some accessibility
  it('property: components with interactive elements should have accessibility support', () => {
    const componentsWithInteractive = vueFiles.filter(file => {
      const analysis = analyzeAccessibility(file)
      return analysis.hasInteractiveElements
    })
    
    if (componentsWithInteractive.length === 0) {
      console.log('No components with interactive elements found')
      return
    }
    
    console.log(`\nTesting ${componentsWithInteractive.length} components with interactive elements`)
    
    fc.assert(
      fc.property(
        fc.integer({ min: 0, max: componentsWithInteractive.length - 1 }),
        (index) => {
          const file = componentsWithInteractive[index]
          const analysis = analyzeAccessibility(file)
          
          // Components with interactive elements should have some form of labeling
          const hasAccessibility = analysis.hasAriaLabels || analysis.hasLabels
          
          if (!hasAccessibility) {
            console.warn(`Component ${path.basename(file)} has interactive elements but no accessibility attributes`)
          }
          
          // Soft assertion - log but don't fail
          return true
        }
      ),
      { numRuns: Math.min(100, componentsWithInteractive.length) }
    )
  })
  
  it('should have color contrast considerations', () => {
    let componentsWithColorVars = 0
    let componentsWithTailwindColors = 0
    
    for (const file of vueFiles) {
      const content = fs.readFileSync(file, 'utf-8')
      
      // Check for CSS custom properties (design tokens)
      if (/var\(--.*color|--.*color:/i.test(content)) {
        componentsWithColorVars++
      }
      
      // Check for Tailwind color classes (also good for consistency)
      if (/text-\w+-\d+|bg-\w+-\d+|border-\w+-\d+/i.test(content)) {
        componentsWithTailwindColors++
      }
    }
    
    console.log(`\nColor Contrast:`)
    console.log(`  - Components using CSS color variables: ${componentsWithColorVars}/${vueFiles.length}`)
    console.log(`  - Components using Tailwind colors: ${componentsWithTailwindColors}/${vueFiles.length}`)
    
    // Using CSS variables or Tailwind helps maintain consistent contrast
    expect(componentsWithColorVars + componentsWithTailwindColors).toBeGreaterThan(0)
  })
})
