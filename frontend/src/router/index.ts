import { createRouter, createWebHashHistory } from 'vue-router'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', redirect: '/dashboard' },
    { path: '/profiles', name: 'profiles', component: () => import('@/views/ProfilesView.vue') },
    { path: '/dashboard', name: 'dashboard', component: () => import('@/views/DashboardView.vue') },
    { path: '/users', name: 'users', component: () => import('@/views/UsersView.vue') },
    { path: '/inbounds', name: 'inbounds', component: () => import('@/views/InboundsView.vue') },
    { path: '/outbounds', name: 'outbounds', component: () => import('@/views/OutboundsView.vue') },
    { path: '/routing', name: 'routing', component: () => import('@/views/RoutingView.vue') },
    { path: '/domains', name: 'domains', component: () => import('@/views/DomainsView.vue') },
    { path: '/wireguard', name: 'wireguard', component: () => import('@/views/WireGuardView.vue') },
    { path: '/settings', name: 'settings', component: () => import('@/views/SettingsView.vue') },
    { path: '/logs', name: 'logs', component: () => import('@/views/LogsView.vue') },
    { path: '/tokens', name: 'tokens', component: () => import('@/views/TokensView.vue') }
  ]
})

export default router
