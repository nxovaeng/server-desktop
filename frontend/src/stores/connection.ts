import { defineStore } from 'pinia'
import {
  desktopApi,
  formatError,
  type ConnectionStatus,
  type ServerProfile
} from '@/api/client'
import type { GeoDataInfo } from '@/api/types'

export interface TofuChallenge {
  profileId: string
  profileName: string
  fingerprint: string
}

export const useConnectionStore = defineStore('connection', {
  state: () => ({
    profiles: [] as ServerProfile[],
    activeProfileId: '' as string,
    connecting: false,
    status: {
      connected: false,
      profile_id: '',
      profile_name: '',
      mode: '',
      latency_ms: 0
    } as ConnectionStatus,
    tofuChallenge: null as TofuChallenge | null,
    geoDataCache: new Map<string, GeoDataInfo>()
  }),
  getters: {
    activeProfile(state): ServerProfile | undefined {
      return state.profiles.find((p) => p.id === state.activeProfileId)
    },
    cachedGeoData(state): GeoDataInfo | undefined {
      if (!state.status.profile_id) return undefined
      return state.geoDataCache.get(state.status.profile_id)
    }
  },
  actions: {
    async loadProfiles() {
      const res = await desktopApi.listProfiles()
      this.profiles = res.profiles || []
      if (res.active_profile_id) {
        this.activeProfileId = res.active_profile_id
      } else if (this.profiles.length > 0 && !this.activeProfileId) {
        this.activeProfileId = this.profiles[0].id
      }
    },

    async syncStatus() {
      try {
        this.status = await desktopApi.getConnectionStatus()
        if (this.status.connected && this.status.profile_id) {
          this.activeProfileId = this.status.profile_id
        }
      } catch {
        // Ignore when bridge unavailable in browser preview
      }
    },

    async connect(profileId: string, trustTOFU = false): Promise<ConnectionStatus | null> {
      this.connecting = true
      try {
        const st = await desktopApi.connectProfile(profileId, trustTOFU)
        this.status = st
        this.activeProfileId = st.profile_id
        this.tofuChallenge = null
        await this.loadProfiles()
        return st
      } catch (err) {
        const msg = formatError(err)
        const match = msg.match(/unknown SSH host key for .* \(fingerprint: (SHA256:[^)]+)\)/)
        if (match && !trustTOFU) {
          const prof = this.profiles.find((p) => p.id === profileId)
          this.tofuChallenge = {
            profileId,
            profileName: prof?.name || profileId,
            fingerprint: match[1]
          }
          return null
        }
        throw err
      } finally {
        this.connecting = false
      }
    },

    async confirmTOFU() {
      if (!this.tofuChallenge) return null
      const id = this.tofuChallenge.profileId
      return this.connect(id, true)
    },

    cancelTOFU() {
      this.tofuChallenge = null
    },

    async disconnect() {
      await desktopApi.disconnect()
      this.status = {
        connected: false,
        profile_id: '',
        profile_name: '',
        mode: '',
        latency_ms: 0
      }
    },

    async testLatency(): Promise<number> {
      const rtt = await desktopApi.testLatency()
      this.status.latency_ms = rtt
      return rtt
    },

    async restartPanelAndReconnect() {
      this.connecting = true
      try {
        const st = await desktopApi.restartPanelAndReconnect()
        this.status = st
        return st
      } finally {
        this.connecting = false
      }
    },

    async getGeoData(force = false): Promise<GeoDataInfo> {
      const pid = this.status.profile_id
      if (!force && pid && this.geoDataCache.has(pid)) {
        return this.geoDataCache.get(pid)!
      }
      const data = await desktopApi.getGeoData()
      if (pid) {
        this.geoDataCache.set(pid, data)
      }
      return data
    }
  }
})
