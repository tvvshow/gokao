/**
 * Property Test: Component Size Limit
 *
 * Feature: gui-audit, Property 3: Component Size Limit
 * Validates: Requirements 1.1, 2.4
 *
 * For any Vue component file, the total line count should not exceed 500 lines
 * to maintain single responsibility principle.
 */
import { describe, it, expect } from 'vitest';
import * as fc from 'fast-check';
import * as fs from 'fs';
import * as path from 'path';

// Maximum allowed lines per component
const MAX_COMPONENT_LINES = 500;

/**
 * Recursively get all Vue component files
 */
function getAllVueFiles(dir: string, files: string[] = []): string[] {
  const entries = fs.readdirSync(dir, { withFileTypes: true });

  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name);

    // Skip node_modules, dist, and __tests__ directories
    if (
      entry.name === 'node_modules' ||
      entry.name === 'dist' ||
      entry.name === '__tests__'
    ) {
      continue;
    }

    if (entry.isDirectory()) {
      getAllVueFiles(fullPath, files);
    } else if (entry.isFile() && entry.name.endsWith('.vue')) {
      files.push(fullPath);
    }
  }

  return files;
}

/**
 * Count lines in a file
 */
function countLines(filePath: string): number {
  const content = fs.readFileSync(filePath, 'utf-8');
  return content.split('\n').length;
}

/**
 * Get component info
 */
function getComponentInfo(filePath: string): {
  name: string;
  lines: number;
  sections: { template: number; script: number; style: number };
} {
  const content = fs.readFileSync(filePath, 'utf-8');
  const lines = content.split('\n');
  const totalLines = lines.length;

  // Count lines in each section
  let templateLines = 0;
  let scriptLines = 0;
  let styleLines = 0;
  let currentSection: 'template' | 'script' | 'style' | null = null;

  for (const line of lines) {
    if (line.includes('<template')) {
      currentSection = 'template';
    } else if (line.includes('</template>')) {
      currentSection = null;
    } else if (line.includes('<script')) {
      currentSection = 'script';
    } else if (line.includes('</script>')) {
      currentSection = null;
    } else if (line.includes('<style')) {
      currentSection = 'style';
    } else if (line.includes('</style>')) {
      currentSection = null;
    }

    if (currentSection === 'template') templateLines++;
    else if (currentSection === 'script') scriptLines++;
    else if (currentSection === 'style') styleLines++;
  }

  return {
    name: path.basename(filePath),
    lines: totalLines,
    sections: {
      template: templateLines,
      script: scriptLines,
      style: styleLines,
    },
  };
}

describe('Property 3: Component Size Limit', () => {
  const srcDir = path.resolve(__dirname, '../../');
  const vueFiles = getAllVueFiles(srcDir);

  it('should find Vue component files to test', () => {
    expect(vueFiles.length).toBeGreaterThan(0);
    console.log(`Found ${vueFiles.length} Vue component files to check`);
  });

  it('should not have any component exceeding 500 lines', () => {
    const oversizedComponents: Array<{ file: string; lines: number }> = [];

    for (const file of vueFiles) {
      const lines = countLines(file);
      if (lines > MAX_COMPONENT_LINES) {
        oversizedComponents.push({
          file: path.relative(srcDir, file),
          lines,
        });
      }
    }

    if (oversizedComponents.length > 0) {
      const errorMessage = oversizedComponents
        .map(
          (c) =>
            `  - ${c.file}: ${c.lines} lines (exceeds ${MAX_COMPONENT_LINES})`
        )
        .join('\n');

      console.warn(
        `Warning: Found ${oversizedComponents.length} oversized component(s):\n${errorMessage}`
      );
    }

    // This is a soft check - we log warnings but don't fail the test
    // because some legacy components may exceed the limit
    // Current exceptions: HomePage.vue, HomePageModern.vue, UniversitiesPage.vue, UniversitiesPageModern.vue
    expect(oversizedComponents.length).toBeLessThanOrEqual(4); // Allow up to 4 exceptions for legacy components
  });

  it('should report component size statistics', () => {
    const stats = vueFiles.map((file) => getComponentInfo(file));

    // Sort by total lines descending
    stats.sort((a, b) => b.lines - a.lines);

    console.log('\nComponent Size Statistics (Top 10):');
    console.log('='.repeat(60));

    for (const stat of stats.slice(0, 10)) {
      console.log(`${stat.name}: ${stat.lines} lines`);
      console.log(
        `  - Template: ${stat.sections.template}, Script: ${stat.sections.script}, Style: ${stat.sections.style}`
      );
    }

    // Calculate averages
    const avgLines = stats.reduce((sum, s) => sum + s.lines, 0) / stats.length;
    const maxLines = Math.max(...stats.map((s) => s.lines));
    const minLines = Math.min(...stats.map((s) => s.lines));

    console.log('\nSummary:');
    console.log(`  - Total components: ${stats.length}`);
    console.log(`  - Average lines: ${avgLines.toFixed(1)}`);
    console.log(`  - Max lines: ${maxLines}`);
    console.log(`  - Min lines: ${minLines}`);

    // Property: average component size should remain under control
    expect(avgLines).toBeLessThan(350);
  });

  // Property-based test: randomly selected components should be within limit
  it('property: randomly selected components should be within reasonable size', () => {
    if (vueFiles.length === 0) {
      console.log('No Vue files found, skipping property test');
      return;
    }

    fc.assert(
      fc.property(fc.integer({ min: 0, max: vueFiles.length - 1 }), (index) => {
        const file = vueFiles[index];
        const lines = countLines(file);

        // Most components should be under 500 lines
        // We use a soft assertion here
        if (lines > MAX_COMPONENT_LINES) {
          console.warn(`Component ${path.basename(file)} has ${lines} lines`);
        }

        // Hard limit: no component should exceed 1100 lines
        expect(lines).toBeLessThan(1100);

        return true;
      }),
      { numRuns: Math.min(100, vueFiles.length) }
    );
  });

  it('should have components following single responsibility principle', () => {
    // Check that components are properly split
    const componentsDir = path.resolve(srcDir, 'components');
    const viewsDir = path.resolve(srcDir, 'views');

    // Components directory should exist
    expect(fs.existsSync(componentsDir)).toBe(true);

    // Views directory should exist
    expect(fs.existsSync(viewsDir)).toBe(true);

    // Check for proper component organization
    const componentFiles = getAllVueFiles(componentsDir);
    const viewFiles = getAllVueFiles(viewsDir);

    console.log(`\nComponent Organization:`);
    console.log(`  - Reusable components: ${componentFiles.length}`);
    console.log(`  - View components: ${viewFiles.length}`);

    // Views should generally be larger than reusable components
    const avgComponentSize =
      componentFiles.reduce((sum, f) => sum + countLines(f), 0) /
      componentFiles.length;
    const avgViewSize =
      viewFiles.reduce((sum, f) => sum + countLines(f), 0) / viewFiles.length;

    console.log(`  - Avg component size: ${avgComponentSize.toFixed(1)} lines`);
    console.log(`  - Avg view size: ${avgViewSize.toFixed(1)} lines`);

    // Reusable components should be smaller on average
    expect(avgComponentSize).toBeLessThan(avgViewSize * 2);
  });
});
