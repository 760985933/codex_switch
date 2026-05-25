import { onBeforeUnmount, onMounted } from 'vue'
import { EventsOff, EventsOn } from '../../wailsjs/runtime/runtime'
import type { BridgeStatusPayload, LogEntry } from '../types'

interface BridgeEventHandlers {
  onStatus?: (payload: BridgeStatusPayload) => void
  onLog?: (entry: LogEntry) => void
}

export function useBridgeEvents(handlers: BridgeEventHandlers): void {
  onMounted(() => {
    if (handlers.onStatus) {
      void EventsOn('bridge:status', handlers.onStatus)
    }
    if (handlers.onLog) {
      void EventsOn('log:entry', handlers.onLog)
    }
  })

  onBeforeUnmount(() => {
    void EventsOff('bridge:status')
    void EventsOff('log:entry')
  })
}
