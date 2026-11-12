<template>
  <n-config-provider :theme="naiveTheme" :theme-overrides="themeOverrides">
    <n-message-provider>
      <n-dialog-provider>
        <MainLayout />
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { darkTheme } from 'naive-ui'
import MainLayout from './layouts/MainLayout.vue'
import { useThemeStore } from './stores/theme'
import { getThemeOverrides } from './theme'

const theme = useThemeStore()

// 深色用 Naive darkTheme + overrides；浅色用默认（浅色）主题 + overrides
const naiveTheme = computed(() => (theme.isDark ? darkTheme : undefined))
const themeOverrides = computed(() => getThemeOverrides(theme.mode))

onMounted(() => {
  theme.applyStored()
})
</script>
