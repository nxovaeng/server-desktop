import type { GlobalThemeOverrides } from 'naive-ui'

export type ThemeMode = 'dark' | 'light'

export const THEME_STORAGE_KEY = 'xray-panel-theme'

/**
 * 主题色板：一套克制的冷色（青/蓝系），深色默认、浅色可选。
 * 主色 sky 系：深色用偏亮的 sky-400 保证深底对比，浅色用 sky-600 保证白底对比；
 * 语义色只在需要与主色区分时微调，其余沿用 Naive 默认。
 */
const fontFamily =
  "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'PingFang SC', 'Microsoft YaHei', sans-serif"
const fontFamilyMono = 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace'

const darkOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#38bdf8',
    primaryColorHover: '#7dd3fc',
    primaryColorPressed: '#0284c7',
    primaryColorSuppl: '#0ea5e9',
    infoColor: '#38bdf8',
    infoColorHover: '#7dd3fc',
    infoColorPressed: '#0284c7',
    infoColorSuppl: '#0ea5e9',
    successColor: '#34d399',
    successColorHover: '#6ee7b7',
    successColorPressed: '#059669',
    successColorSuppl: '#10b981',
    warningColor: '#fbbf24',
    warningColorHover: '#fcd34d',
    warningColorPressed: '#d97706',
    warningColorSuppl: '#f59e0b',
    errorColor: '#f87171',
    errorColorHover: '#fca5a5',
    errorColorPressed: '#dc2626',
    errorColorSuppl: '#ef4444',
    baseColor: '#17171d',
    bodyColor: '#101014',
    cardColor: '#17171d',
    modalColor: '#1d1d24',
    popoverColor: '#1d1d24',
    tableColor: '#17171d',
    tableHeaderColor: '#1a1a21',
    tableColorHover: '#1f1f28',
    tableColorStriped: '#181820',
    inputColor: '#1e1e26',
    inputColorDisabled: '#1a1a21',
    codeColor: '#14141a',
    tabColor: '#17171d',
    actionColor: '#1e1e26',
    tagColor: '#232330',
    avatarColor: '#232330',
    hoverColor: '#1f1f28',
    pressedColor: '#23232e',
    textColorBase: '#e9e9ee',
    textColor1: '#e9e9ee',
    textColor2: '#a2a2ad',
    textColor3: '#6f6f7a',
    textColorDisabled: '#52525e',
    placeholderColor: '#6f6f7a',
    iconColor: '#a2a2ad',
    borderColor: '#26262f',
    dividerColor: '#232329',
    closeIconColor: '#a2a2ad',
    scrollbarColor: 'rgba(255, 255, 255, 0.16)',
    scrollbarColorHover: 'rgba(255, 255, 255, 0.28)',
    fontFamily,
    fontFamilyMono,
    borderRadius: '8px',
    borderRadiusSmall: '6px'
  },
  Layout: {
    color: '#101014',
    colorEmbedded: '#141419',
    headerColor: '#141419',
    siderColor: '#141419',
    headerBorderColor: '#232329',
    siderBorderColor: '#232329',
    textColor: '#e9e9ee'
  },
  Menu: {
    itemColorHover: '#1f1f28',
    itemColorActive: 'rgba(56, 189, 248, 0.12)',
    itemColorActiveHover: 'rgba(56, 189, 248, 0.16)',
    itemTextColor: '#a2a2ad',
    itemTextColorHover: '#e9e9ee',
    itemTextColorActive: '#7dd3fc',
    itemTextColorActiveHover: '#7dd3fc',
    itemIconColor: '#71717e',
    itemIconColorHover: '#e9e9ee',
    itemIconColorActive: '#38bdf8',
    itemIconColorActiveHover: '#38bdf8'
  },
  Card: {
    color: '#17171d',
    colorEmbedded: '#1a1a21',
    borderColor: '#26262f',
    titleTextColor: '#e9e9ee'
  }
}

const lightOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#0284c7',
    primaryColorHover: '#0ea5e9',
    primaryColorPressed: '#0369a1',
    primaryColorSuppl: '#38bdf8',
    infoColor: '#0284c7',
    infoColorHover: '#0ea5e9',
    infoColorPressed: '#0369a1',
    infoColorSuppl: '#38bdf8',
    successColor: '#16a34a',
    successColorHover: '#22c55e',
    successColorPressed: '#15803d',
    successColorSuppl: '#4ade80',
    warningColor: '#d97706',
    warningColorHover: '#f59e0b',
    warningColorPressed: '#b45309',
    warningColorSuppl: '#fbbf24',
    errorColor: '#dc2626',
    errorColorHover: '#ef4444',
    errorColorPressed: '#b91c1c',
    errorColorSuppl: '#f87171',
    baseColor: '#ffffff',
    bodyColor: '#f3f4f6',
    cardColor: '#ffffff',
    modalColor: '#ffffff',
    popoverColor: '#ffffff',
    tableColor: '#ffffff',
    tableHeaderColor: '#f9fafb',
    tableColorHover: '#f3f4f6',
    tableColorStriped: '#fafafa',
    inputColor: '#ffffff',
    inputColorDisabled: '#f3f4f6',
    codeColor: '#f3f4f6',
    tabColor: '#ffffff',
    actionColor: '#f9fafb',
    tagColor: '#f3f4f6',
    avatarColor: '#e5e7eb',
    hoverColor: '#f3f4f6',
    pressedColor: '#e5e7eb',
    textColorBase: '#1c2530',
    textColor1: '#1c2530',
    textColor2: '#4b5563',
    textColor3: '#8a94a3',
    textColorDisabled: '#c3c9d1',
    placeholderColor: '#8a94a3',
    iconColor: '#6b7280',
    borderColor: '#e4e6eb',
    dividerColor: '#eceef1',
    closeIconColor: '#6b7280',
    scrollbarColor: 'rgba(0, 0, 0, 0.16)',
    scrollbarColorHover: 'rgba(0, 0, 0, 0.28)',
    fontFamily,
    fontFamilyMono,
    borderRadius: '8px',
    borderRadiusSmall: '6px'
  },
  Layout: {
    color: '#f3f4f6',
    colorEmbedded: '#ffffff',
    headerColor: '#ffffff',
    siderColor: '#ffffff',
    headerBorderColor: '#e4e6eb',
    siderBorderColor: '#e4e6eb',
    textColor: '#1c2530'
  },
  Menu: {
    itemColorHover: '#f3f4f6',
    itemColorActive: 'rgba(2, 132, 199, 0.08)',
    itemColorActiveHover: 'rgba(2, 132, 199, 0.12)',
    itemTextColor: '#4b5563',
    itemTextColorHover: '#1c2530',
    itemTextColorActive: '#0284c7',
    itemTextColorActiveHover: '#0284c7',
    itemIconColor: '#8a94a3',
    itemIconColorHover: '#1c2530',
    itemIconColorActive: '#0284c7',
    itemIconColorActiveHover: '#0284c7'
  },
  Card: {
    color: '#ffffff',
    colorEmbedded: '#f9fafb',
    borderColor: '#e4e6eb',
    titleTextColor: '#1c2530'
  }
}

export function getThemeOverrides(mode: ThemeMode): GlobalThemeOverrides {
  return mode === 'light' ? lightOverrides : darkOverrides
}

/** 把主题模式同步到 <html data-theme>，CSS 变量与 color-scheme 跟着走。 */
export function applyThemeMode(mode: ThemeMode): void {
  document.documentElement.dataset.theme = mode
}
