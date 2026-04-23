import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useAuthStore } from '@/store/auth'
import type { User } from '@/types'
import { ToastProvider } from '@/components/ui/toast'
import App from './App'

// Mock the teams API (called by AppLayout on mount)
vi.mock('@/api/teams', () => ({
  listTeamsApi: vi.fn().mockResolvedValue([
    { id: 1, name: 'Team 1', description: '', pmId: 1, createdAt: '', updatedAt: '' },
  ]),
  listMembersApi: vi.fn().mockResolvedValue([]),
}))

// Mock the mainItems API (called by ItemViewPage)
vi.mock('@/api/mainItems', () => ({
  listMainItemsApi: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, size: 20 }),
  getMainItemApi: vi.fn().mockResolvedValue({
    id: 123, teamId: 1, code: 'MI-0123', title: 'Test', priority: 'P2',
    proposerId: 1, assigneeId: null, startDate: null, expectedEndDate: null,
    actualEndDate: null, status: 'progressing', completion: 0,
    createdAt: '', updatedAt: '', subItems: [],
  }),
  createMainItemApi: vi.fn(),
  updateMainItemApi: vi.fn(),
}))

// Mock the subItems API
vi.mock('@/api/subItems', () => ({
  listSubItemsApi: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 20 }),
}))

const mockUser: User = {
  id: 1,
  username: 'testuser',
  displayName: 'Test User',
  isSuperAdmin: false,
}

const superAdminUser: User = {
  ...mockUser,
  id: 2,
  username: 'admin',
  isSuperAdmin: true,
}

function renderApp(initialPath = '/') {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <ToastProvider>
        <MemoryRouter initialEntries={[initialPath]}>
          <App />
        </MemoryRouter>
      </ToastProvider>
    </QueryClientProvider>,
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
    it('redirects /users to /items for non-admin authenticated user', () => {
      useAuthStore.getState().setAuth('token', mockUser)
      renderApp('/users')
      expect(screen.getByTestId('item-view-page')).toBeInTheDocument()
    })

    it('renders UserManagementPage at /users for super admin', () => {
      useAuthStore.getState().setAuth('token', superAdminUser)
      useAuthStore.getState().setPermissions({
        isSuperAdmin: true,
        teamPermissions: {},
      })
      renderApp('/users')
      expect(screen.getByTestId('user-management-page')).toBeInTheDocument()
    })

    it('redirects /users to /login for unauthenticated user', () => {
      renderApp('/users')
      expect(screen.getByTestId('login-page')).toBeInTheDocument()
    })
  })
})
