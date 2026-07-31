import { create } from "zustand"

type MapState = {
  selectedEventID: string | null
  selectEvent: (eventID: string) => void
}

export const useMapStore = create<MapState>((set) => ({
  selectedEventID: null,
  selectEvent: (selectedEventID) => set({ selectedEventID }),
}))
