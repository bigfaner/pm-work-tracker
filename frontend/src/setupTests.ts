import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterEach, vi } from 'vitest'

// Mock frappe-gantt (vanilla JS library with scss imports that break in jsdom)
vi.mock('frappe-gantt', () => ({
  default: class MockGantt {
    change_view_mode = vi.fn()
    refresh = vi.fn()
  },
}))

// Mock window.matchMedia for antd responsive components
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
