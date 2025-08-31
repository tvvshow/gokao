import { ref, computed, onMounted, onUnmounted, watch, type Ref } from 'vue'

// 响应式断点定义
export const breakpoints = {
  xs: 0,
  sm: 576,
  md: 768,
  lg: 992,
  xl: 1200,
  xxl: 1400
} as const

export type Breakpoint = keyof typeof breakpoints

// 响应式组合式函数
export function useResponsive() {
  const windowWidth = ref(window.innerWidth)
  const windowHeight = ref(window.innerHeight)

  // 更新窗口尺寸
  const updateSize = () => {
    windowWidth.value = window.innerWidth
    windowHeight.value = window.innerHeight
  }

  // 当前断点
  const currentBreakpoint = computed<Breakpoint>(() => {
    const width = windowWidth.value
    if (width >= breakpoints.xxl) return 'xxl'
    if (width >= breakpoints.xl) return 'xl'
    if (width >= breakpoints.lg) return 'lg'
    if (width >= breakpoints.md) return 'md'
    if (width >= breakpoints.sm) return 'sm'
    return 'xs'
  })

  // 断点检查函数
  const isXs = computed(() => currentBreakpoint.value === 'xs')
  const isSm = computed(() => currentBreakpoint.value === 'sm')
  const isMd = computed(() => currentBreakpoint.value === 'md')
  const isLg = computed(() => currentBreakpoint.value === 'lg')
  const isXl = computed(() => currentBreakpoint.value === 'xl')
  const isXxl = computed(() => currentBreakpoint.value === 'xxl')

  // 范围检查函数
  const isSmAndUp = computed(() => windowWidth.value >= breakpoints.sm)
  const isMdAndUp = computed(() => windowWidth.value >= breakpoints.md)
  const isLgAndUp = computed(() => windowWidth.value >= breakpoints.lg)
  const isXlAndUp = computed(() => windowWidth.value >= breakpoints.xl)

  const isSmAndDown = computed(() => windowWidth.value < breakpoints.md)
  const isMdAndDown = computed(() => windowWidth.value < breakpoints.lg)
  const isLgAndDown = computed(() => windowWidth.value < breakpoints.xl)

  // 移动端检查
  const isMobile = computed(() => windowWidth.value < breakpoints.md)
  const isTablet = computed(() => 
    windowWidth.value >= breakpoints.md && windowWidth.value < breakpoints.lg
  )
  const isDesktop = computed(() => windowWidth.value >= breakpoints.lg)

  // 设备方向
  const isLandscape = computed(() => windowWidth.value > windowHeight.value)
  const isPortrait = computed(() => windowWidth.value <= windowHeight.value)

  // 触控设备检查
  const isTouchDevice = computed(() => {
    return 'ontouchstart' in window || navigator.maxTouchPoints > 0
  })

  // 响应式列数计算
  const getResponsiveColumns = (config: Partial<Record<Breakpoint, number>>) => {
    return computed(() => {
      const bp = currentBreakpoint.value
      return config[bp] || config.xs || 1
    })
  }

  // 响应式间距计算
  const getResponsiveSpacing = (config: Partial<Record<Breakpoint, number>>) => {
    return computed(() => {
      const bp = currentBreakpoint.value
      return config[bp] || config.xs || 16
    })
  }

  // 响应式字体大小
  const getResponsiveFontSize = (config: Partial<Record<Breakpoint, string>>) => {
    return computed(() => {
      const bp = currentBreakpoint.value
      return config[bp] || config.xs || '14px'
    })
  }

  // 生命周期管理
  onMounted(() => {
    window.addEventListener('resize', updateSize)
    window.addEventListener('orientationchange', updateSize)
  })

  onUnmounted(() => {
    window.removeEventListener('resize', updateSize)
    window.removeEventListener('orientationchange', updateSize)
  })

  return {
    // 窗口尺寸
    windowWidth: readonly(windowWidth),
    windowHeight: readonly(windowHeight),
    
    // 当前断点
    currentBreakpoint,
    
    // 断点检查
    isXs,
    isSm,
    isMd,
    isLg,
    isXl,
    isXxl,
    
    // 范围检查
    isSmAndUp,
    isMdAndUp,
    isLgAndUp,
    isXlAndUp,
    isSmAndDown,
    isMdAndDown,
    isLgAndDown,
    
    // 设备类型
    isMobile,
    isTablet,
    isDesktop,
    
    // 设备方向
    isLandscape,
    isPortrait,
    
    // 触控设备
    isTouchDevice,
    
    // 工具函数
    getResponsiveColumns,
    getResponsiveSpacing,
    getResponsiveFontSize
  }
}

// 响应式网格系统
export function useResponsiveGrid() {
  const { currentBreakpoint, isMobile, isTablet } = useResponsive()

  // 默认网格配置
  const defaultGridConfig = {
    xs: { cols: 1, gap: 8 },
    sm: { cols: 2, gap: 12 },
    md: { cols: 3, gap: 16 },
    lg: { cols: 4, gap: 20 },
    xl: { cols: 5, gap: 24 },
    xxl: { cols: 6, gap: 24 }
  }

  const getGridConfig = (customConfig?: Partial<typeof defaultGridConfig>) => {
    const config = { ...defaultGridConfig, ...customConfig }
    return computed(() => config[currentBreakpoint.value])
  }

  // 计算网格样式
  const getGridStyle = (customConfig?: Partial<typeof defaultGridConfig>) => {
    const gridConfig = getGridConfig(customConfig)
    
    return computed(() => ({
      display: 'grid',
      gridTemplateColumns: `repeat(${gridConfig.value.cols}, 1fr)`,
      gap: `${gridConfig.value.gap}px`
    }))
  }

  // 响应式卡片宽度
  const getCardWidth = () => {
    return computed(() => {
      if (isMobile.value) return '100%'
      if (isTablet.value) return 'calc(50% - 8px)'
      return 'calc(33.333% - 12px)'
    })
  }

  return {
    getGridConfig,
    getGridStyle,
    getCardWidth
  }
}

// 响应式导航
export function useResponsiveNavigation() {
  const { isMobile, isTablet } = useResponsive()
  const isMenuCollapsed = ref(false)

  // 自动折叠菜单
  const autoCollapseMenu = () => {
    if (isMobile.value) {
      isMenuCollapsed.value = true
    }
  }

  // 切换菜单状态
  const toggleMenu = () => {
    isMenuCollapsed.value = !isMenuCollapsed.value
  }

  // 导航样式
  const navigationStyle = computed(() => ({
    position: isMobile.value ? 'fixed' : 'relative',
    zIndex: isMobile.value ? 1000 : 'auto',
    width: isMobile.value ? '100%' : 'auto',
    transform: isMobile.value && isMenuCollapsed.value ? 'translateX(-100%)' : 'translateX(0)',
    transition: 'transform 0.3s ease'
  }))

  // 监听断点变化
  watch([isMobile, isTablet], () => {
    autoCollapseMenu()
  })

  return {
    isMenuCollapsed,
    toggleMenu,
    navigationStyle,
    autoCollapseMenu
  }
}

// 响应式表格
export function useResponsiveTable() {
  const { isMobile, isTablet } = useResponsive()

  // 表格显示模式
  const tableMode = computed(() => {
    if (isMobile.value) return 'card'
    if (isTablet.value) return 'compact'
    return 'full'
  })

  // 表格样式
  const tableStyle = computed(() => ({
    overflowX: isMobile.value ? 'hidden' : 'auto',
    fontSize: isMobile.value ? '14px' : '16px'
  }))

  // 列配置
  const getResponsiveColumns = (columns: any[]) => {
    return computed(() => {
      if (isMobile.value) {
        // 移动端只显示关键列
        return columns.filter(col => col.mobile !== false)
      }
      if (isTablet.value) {
        // 平板端隐藏次要列
        return columns.filter(col => col.tablet !== false)
      }
      return columns
    })
  }

  return {
    tableMode,
    tableStyle,
    getResponsiveColumns
  }
}

// 导出只读的响应式状态
function readonly<T>(ref: Ref<T>) {
  return computed(() => ref.value)
}
