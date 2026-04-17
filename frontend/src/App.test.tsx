import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { useAuthStore } from '@/store/auth'
import type { User } from '@/types'
import App from './App'

// Mock the teams API (called by AppLayout on mount)
vi.mock('@/api/teams', () => ({
  listTeamsApi: vi.fn().mockResolvedValue([]),
}))

const mockUser: User = {
  id: 1,
  username: 'testuser',
  display_name: 'Test User',
  is_super_admin: false,
  can_create_team: false,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const superAdminUser: User = {
  ...mockUser,
  id: 2,
  username: 'admin',
  is_super_admin: true,
}

function renderApp(initialPath = '/') {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <App />
    </MemoryRouter>,
  )
}

describe('App Routing', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth()
  })

  describe('Public routes', () => {
    it('renders LoginPage at /login', () => {
      renderApp('/login')
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
    })
  })

  describe('Root redirect', () => {
    it('redirects / to /items when authenticated', () => {
      useAuthStore.getState().setAuth('token', mockUser)
      renderApp('/')
      expect(screen.getByTestId('item-view-page')).toBeInTheDocument()
    })

    it('redirects / to /login when not authenticated', () => {
      renderApp('/')
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
    })
  })

  describe('Protected routes - redirect to login when unauthenticated', () => {
    it('redirects /items to /login', () => {
      renderApp('/items')
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
    })

    it('redirects /items/123 to /login', () => {
      renderApp('/items/123')
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
    })

    it('redirects /items/123/sub/456 to /login', () => {
      renderApp('/items/123/sub/456')
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
    })

    it('redirects /weekly to /login', () => {
      renderApp('/weekly')
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
    })

    it('redirects /gantt to /login', () => {
      renderApp('/gantt')
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
    })

    it('redirects /table to /login', () => {
      renderApp('/table')
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
    })

    it('redirects /item-pool to /login', () => {
      renderApp('/item-pool')
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
    })

    it('redirects /report to /login', () => {
      renderApp('/report')
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
    })
  })

  describe('Protected routes - render pages when authenticated', () => {
    beforeEach(() => {
      useAuthStore.getState().setAuth('token', mockUser)
    })

    it('renders ItemViewPage at /items', () => {
      renderApp('/items')
      expect(screen.getByTestId('item-view-page')).toBeInTheDocument()
    })

    it('renders MainItemDetailPage at /items/:mainItemId', () => {
      renderApp('/items/123')
      expect(screen.getByTestId('main-item-detail-page')).toBeInTheDocument()
    })

    it('renders SubItemDetailPage at /items/:mainItemId/sub/:subItemId', () => {
      renderApp('/items/123/sub/456')
      expect(screen.getByTestId('sub-item-detail-page')).toBeInTheDocument()
    })

    it('renders WeeklyViewPage at /weekly', () => {
      renderApp('/weekly')
      expect(screen.getByTestId('weekly-view-page')).toBeInTheDocument()
    })

    it('renders GanttViewPage at /gantt', () => {
      renderApp('/gantt')
      expect(screen.getByTestId('gantt-view-page')).toBeInTheDocument()
    })

    it('renders TableViewPage at /table', () => {
      renderApp('/table')
      expect(screen.getByTestId('table-view-page')).toBeInTheDocument()
    })

    it('renders ItemPoolPage at /item-pool', () => {
      renderApp('/item-pool')
      expect(screen.getByTestId('item-pool-page')).toBeInTheDocument()
    })

    it('renders ReportPage at /report', () => {
      renderApp('/report')
      expect(screen.getByTestId('report-page')).toBeInTheDocument()
    })
  })

  describe('Admin route', () => {
    it('redirects /admin to /items for non-admin authenticated user', () => {
      useAuthStore.getState().setAuth('token', mockUser)
      renderApp('/admin')
      expect(screen.getByTestId('item-view-page')).toBeInTheDocument()
    })

    it('renders AdminPage at /admin for super admin', () => {
      useAuthStore.getState().setAuth('token', superAdminUser)
      renderApp('/admin')
      expect(screen.getByTestId('admin-page')).toBeInTheDocument()
    })

    it('redirects /admin to /login for unauthenticated user', () => {
      renderApp('/admin')
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
    })
  })
})
