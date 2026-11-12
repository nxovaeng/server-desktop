import { defineStore } from 'pinia'
import { applyThemeMode, THEME_STORAGE_KEY, type ThemeMode } from '@/theme'

function readStoredMode(): ThemeMode {
  try {
    return window.localStorage.getItem(THEME_STORAGE_KEY) === 'light' ? 'light' : 'dark'
  } catch {
    return 'dark'
  }
}

export const useThemeStore = defineStore('theme', {
  state: () => ({
    // 默认深色；偏好持久化到 localStorage
    mode: readStoredMode() as ThemeMode
  }),
  getters: {
    isDark(state): boolean {
      return state.mode === 'dark'
    }
  },
  actions: {
    setMode(mode: ThemeMode) {
      this.mode = mode
      try {
        window.localStorage.setItem(THEME_STORAGE_KEY, mode)
      } catch {
        // 隐私模式等写不进 localStorage 时只影响本次会话
      }
      applyThemeMode(mode)
    },
    toggle() {
      this.setMode(this.mode === 'dark' ? 'light' : 'dark')
    },
    /** 启动时把已持久化的模式同步到 <html>，避免首屏闪烁后才换色。 */
    applyStored() {
      applyThemeMode(this.mode)
    }
  }
})
