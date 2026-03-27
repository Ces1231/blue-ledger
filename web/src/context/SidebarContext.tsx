import { createContext, useContext } from 'react'

interface SidebarContextValue {
  toggleSidebar: () => void
}

export const SidebarContext = createContext<SidebarContextValue>({
  toggleSidebar: () => {},
})

/** Hook to access the sidebar toggle from any component */
export function useSidebarToggle() {
  return useContext(SidebarContext)
}
