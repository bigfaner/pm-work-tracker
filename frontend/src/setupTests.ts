import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterEach, vi } from 'vitest'

// Mock localStorage (required by zustand persist middleware in jsdom)
const localStorageMock = {
  getItem: vi.fn().mockReturnValue(null),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
}
Object.defineProperty(window, 'localStorage', { value: localStorageMock, writable: true })

// Mock frappe-gantt (vanilla JS library with scss imports that break in jsdom)
vi.mock('frappe-gantt', () => ({
  default: class MockGantt {
    change_view_mode = vi.fn()
    refresh = vi.fn()
  },
}))

// Mock pointer capture APIs (not implemented in jsdom, needed by @radix-ui/react-select)
HTMLElement.prototype.hasPointerCapture = vi.fn().mockReturnValue(false)
HTMLElement.prototype.setPointerCapture = vi.fn()
HTMLElement.prototype.releasePointerCapture = vi.fn()

// Mock scrollIntoView (not implemented in jsdom, needed by Radix UI Select)
Element.prototype.scrollIntoView = vi.fn()

// Mock window.matchMedia (required by various Radix UI components)
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})

afterEach(() => {
  cleanup()
})
