<template>
  <div
    ref="containerRef"
    class="virtual-list-container"
    :style="{ height: containerHeight }"
    @scroll="handleScroll"
    role="list"
    :aria-label="ariaLabel"
  >
    <div class="virtual-list-spacer" :style="{ height: `${totalHeight}px` }">
      <div
        class="virtual-list-content"
        :style="{ transform: `translateY(${offsetY}px)` }"
      >
        <div
          v-for="(item, index) in visibleItems"
          :key="getItemKey(item, index)"
          class="virtual-list-item"
          role="listitem"
        >
          <slot :item="item" :index="startIndex + index" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts" generic="T">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue';

defineSlots<{
  default(props: { item: T; index: number }): unknown;
}>();

const props = withDefaults(
  defineProps<{
    items: T[];
    itemHeight: number;
    containerHeight?: string;
    overscan?: number;
    keyField?: string;
    ariaLabel?: string;
  }>(),
  {
    containerHeight: '600px',
    overscan: 3,
    keyField: 'id',
    ariaLabel: '虚拟滚动列表',
  }
);

const containerRef = ref<HTMLElement | null>(null);
const scrollTop = ref(0);

// Calculate total height of all items
const totalHeight = computed(() => props.items.length * props.itemHeight);

// Calculate visible range
const startIndex = computed(() => {
  const start = Math.floor(scrollTop.value / props.itemHeight);
  return Math.max(0, start - props.overscan);
});

const endIndex = computed(() => {
  if (!containerRef.value) return props.overscan * 2;
  const containerHeightPx = containerRef.value.clientHeight;
  const visibleCount = Math.ceil(containerHeightPx / props.itemHeight);
  const end = Math.floor(scrollTop.value / props.itemHeight) + visibleCount;
  return Math.min(props.items.length, end + props.overscan);
});

// Get visible items
const visibleItems = computed(() => {
  return props.items.slice(startIndex.value, endIndex.value);
});

// Calculate offset for positioning
const offsetY = computed(() => startIndex.value * props.itemHeight);

// Get unique key for each item
const getItemKey = (item: T, index: number): string | number => {
  if (typeof item === 'object' && item !== null && props.keyField in item) {
    return (item as Record<string, unknown>)[props.keyField] as string | number;
  }
  return startIndex.value + index;
};

// Handle scroll event
const handleScroll = (event: Event) => {
  const target = event.target as HTMLElement;
  scrollTop.value = target.scrollTop;
};

// Reset scroll position when items change
watch(
  () => props.items,
  () => {
    if (containerRef.value) {
      containerRef.value.scrollTop = 0;
      scrollTop.value = 0;
    }
  }
);

// Expose scroll methods
const scrollToIndex = (index: number) => {
  if (containerRef.value) {
    containerRef.value.scrollTop = index * props.itemHeight;
  }
};

const scrollToTop = () => {
  if (containerRef.value) {
    containerRef.value.scrollTop = 0;
  }
};

defineExpose({
  scrollToIndex,
  scrollToTop,
});

onMounted(() => {
  // Initial setup if needed
});

onUnmounted(() => {
  // Cleanup if needed
});
</script>

<style scoped>
.virtual-list-container {
  overflow-y: auto;
  position: relative;
}

.virtual-list-spacer {
  position: relative;
}

.virtual-list-content {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
}

.virtual-list-item {
  box-sizing: border-box;
}

/* Scrollbar styling */
.virtual-list-container::-webkit-scrollbar {
  width: 8px;
}

.virtual-list-container::-webkit-scrollbar-track {
  background: var(--gray-100, #f3f4f6);
  border-radius: 4px;
}

.virtual-list-container::-webkit-scrollbar-thumb {
  background: var(--gray-300, #d1d5db);
  border-radius: 4px;
}

.virtual-list-container::-webkit-scrollbar-thumb:hover {
  background: var(--gray-400, #9ca3af);
}

/* Dark mode scrollbar */
.dark .virtual-list-container::-webkit-scrollbar-track {
  background: var(--gray-800, #1f2937);
}

.dark .virtual-list-container::-webkit-scrollbar-thumb {
  background: var(--gray-600, #4b5563);
}

.dark .virtual-list-container::-webkit-scrollbar-thumb:hover {
  background: var(--gray-500, #6b7280);
}
</style>
