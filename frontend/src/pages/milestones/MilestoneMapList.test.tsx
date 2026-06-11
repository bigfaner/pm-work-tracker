import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { useTeamStore } from '@/store/team'
import { useAuthStore } from '@/store/auth'
import { ToastProvider } from '@/components/ui/toast'
import MilestoneMapList from './MilestoneMapList'
import type { MilestoneMap, TeamMemberResp } from '@/types'

// --- Mocks ---

vi.mock('@/api/milestones', () => ({
  listMilestoneMapsApi: vi.fn(),
  createMilestoneMapApi: vi.fn(),
}))

vi.mock('@/api/teams', () => ({
  listMembersApi: vi.fn(),
}))

import { listMilestoneMapsApi } from '@/api/milestones'
import { listMembersApi } from '@/api/teams'

// --- Helpers ---

function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0 },
    },
  })
}

const seedMembers: TeamMemberResp[] = [
  {
    bizKey: '1',
    teamKey: 'team-1',
    userKey: 'U001',
    displayName: '张三',
    username: 'zhangsan',
    role: 'pm',
    roleKey: '0',
    roleName: 'pm',
    joinedAt: '2024-01-01',
  },
  {
    bizKey: '2',
    teamKey: 'team-1',
    userKey: 'U002',
    displayName: '李四',
    username: 'lisi',
    role: 'dev',
    roleKey: '1',
    roleName: 'dev',
    joinedAt: '2024-01-01',
  },
]

function makeMap(overrides: Partial<MilestoneMap> = {}): MilestoneMap {
  return {
    bizKey: 'map-1',
    teamKey: 'team-1',
    creatorKey: 'U001',
    creatorName: '张三',
    assigneeKey: 'U001',
    assigneeName: '张三',
    mapName: '产品 MVP',
    mapDesc: '',
    mapStatus: 'executing',
    statusName: '实施中',
    planStartDate: '2026-05-01',
    expectedEndDate: '2026-12-31',
    milestoneCount: 4,
    itemCount: 12,
    overallProgress: 60,
    milestoneSummary: [
      { bizKey: 'ms-1', name: 'M1 需求确认', status: 'completed', progress: 100 },
      { bizKey: 'ms-2', name: 'M2 开发', status: 'in_progress', progress: 60 },
      { bizKey: 'ms-3', name: 'M3 测试', status: 'not_started', progress: 0 },
      { bizKey: 'ms-4', name: 'M4 上线', status: 'not_started', progress: 0 },
    ],
    createTime: '2026-01-01T00:00:00Z',
    dbUpdateTime: '2026-01-01T00:00:00Z',
    ...overrides,
  }
}

function setAuth(permissions: string[] = ['milestone:create']) {
  useAuthStore.setState({
    token: 'test-token',
    user: {
      bizKey: '1',
      username: 'testuser',
      displayName: 'Test User',
      createTime: '',
    },
    isAuthenticated: true,
    _hasHydrated: true,
    permissions: { teamPermissions: { 'team-1': permissions } },
    permissionsLoadedAt: Date.now(),
  })
}

function setupMocks(mapsResponse: { items: MilestoneMap[], total: number }) {
  vi.mocked(listMembersApi).mockResolvedValue(seedMembers)
  vi.mocked(listMilestoneMapsApi).mockResolvedValue({
    ...mapsResponse,
    page: 1,
    size: 20,
  })
}

function renderList() {
  const qc = createQueryClient()
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <ToastProvider>
          <MilestoneMapList />
        </ToastProvider>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

// --- Tests ---

describe('MilestoneMapList', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    useTeamStore.setState({ currentTeamId: 'team-1' })
    setAuth()
  })

  // AC-1: Card grid renders with correct layout
  it('renders card grid with maps', async () => {
    setupMocks({ items: [makeMap()], total: 1 })

    renderList()
    await waitFor(() => {
      expect(
        screen.getByTestId('milestone-map-card-map-1'),
      ).toBeInTheDocument()
    })
    expect(screen.getByText('产品 MVP')).toBeInTheDocument()
  })

  // AC-2: Name search (debounce, client-side fuzzy)
  it('filters maps by name search', async () => {
    setupMocks({
      items: [
        makeMap({ bizKey: 'map-1', mapName: '产品 MVP' }),
        makeMap({
          bizKey: 'map-2',
          mapName: '二期迭代',
          assigneeKey: 'U002',
          assigneeName: '李四',
        }),
      ],
      total: 2,
    })

    renderList()
    await waitFor(() => {
      expect(screen.getByText('产品 MVP')).toBeInTheDocument()
      expect(screen.getByText('二期迭代')).toBeInTheDocument()
    })

    const user = userEvent.setup()
    await user.type(screen.getByTestId('search-input'), '产品')

    // Wait for debounce (300ms) + render
    await waitFor(
      () => {
        expect(screen.queryByText('二期迭代')).not.toBeInTheDocument()
      },
      { timeout: 2000 },
    )
    expect(screen.getByText('产品 MVP')).toBeInTheDocument()
  })

  // AC-2: Assignee filter
  it('filters maps by assignee', async () => {
    setupMocks({
      items: [
        makeMap({
          bizKey: 'map-1',
          mapName: 'Map A',
          assigneeKey: 'U001',
          assigneeName: '张三',
        }),
        makeMap({
          bizKey: 'map-2',
          mapName: 'Map B',
          assigneeKey: 'U002',
          assigneeName: '李四',
        }),
      ],
      total: 2,
    })

    renderList()
    await waitFor(() => {
      expect(screen.getByText('Map A')).toBeInTheDocument()
      expect(screen.getByText('Map B')).toBeInTheDocument()
    })

    const user = userEvent.setup()
    await user.click(screen.getByTestId('assignee-filter'))
    const listbox = await screen.findByRole('listbox')
    await user.click(within(listbox).getByText('张三'))

    await waitFor(() => {
      expect(screen.getByText('Map A')).toBeInTheDocument()
      expect(screen.queryByText('Map B')).not.toBeInTheDocument()
    })
  })

  // AC-2: Status filter
  it('filters maps by status', async () => {
    setupMocks({
      items: [
        makeMap({
          bizKey: 'map-1',
          mapName: 'Map A',
          mapStatus: 'executing',
          statusName: '实施中',
        }),
        makeMap({
          bizKey: 'map-2',
          mapName: 'Map B',
          mapStatus: 'planning',
          statusName: '规划中',
        }),
      ],
      total: 2,
    })

    renderList()
    await waitFor(() => {
      expect(screen.getByText('Map A')).toBeInTheDocument()
      expect(screen.getByText('Map B')).toBeInTheDocument()
    })

    const user = userEvent.setup()
    await user.click(screen.getByTestId('status-filter-executing'))

    await waitFor(() => {
      expect(screen.getByText('Map A')).toBeInTheDocument()
      expect(screen.queryByText('Map B')).not.toBeInTheDocument()
    })
  })

  // AC-2: Reset button clears all filters
  it('resets all filters', async () => {
    setupMocks({ items: [makeMap()], total: 1 })

    renderList()
    await waitFor(() => {
      expect(screen.getByText('产品 MVP')).toBeInTheDocument()
    })

    const user = userEvent.setup()
    await user.type(screen.getByTestId('search-input'), 'xyz')
    await user.click(screen.getByText('重置'))
    expect(screen.getByTestId('search-input')).toHaveValue('')
  })

  // AC-3: Empty state
  it('shows empty state when no maps exist', async () => {
    setupMocks({ items: [], total: 0 })

    renderList()
    await waitFor(() => {
      expect(screen.getByTestId('empty-state')).toBeInTheDocument()
    })
    expect(screen.getByText('暂无里程碑图')).toBeInTheDocument()
    expect(screen.getByTestId('create-map-btn')).toBeInTheDocument()
  })

  // AC-3: Loading state shows skeleton cards
  it('shows loading skeleton cards', () => {
    // Never resolve to keep loading state
    vi.mocked(listMembersApi).mockReturnValue(new Promise(() => {}))
    vi.mocked(listMilestoneMapsApi).mockReturnValue(new Promise(() => {}))

    renderList()
    expect(screen.getByTestId('loading-skeleton')).toBeInTheDocument()
  })

  // AC-3: Error state shows retry
  it('shows error state with retry button', async () => {
    vi.mocked(listMembersApi).mockResolvedValue(seedMembers)
    vi.mocked(listMilestoneMapsApi).mockRejectedValue(
      new Error('server error'),
    )

    renderList()
    await waitFor(() => {
      expect(screen.getByTestId('error-state')).toBeInTheDocument()
    })
    expect(screen.getByText('加载失败，请重试')).toBeInTheDocument()
    expect(screen.getByText('重试')).toBeInTheDocument()
  })

  // AC-4: Dashed create card triggers dialog
  it('shows dashed create card when user has permission', async () => {
    setupMocks({ items: [makeMap()], total: 1 })

    renderList()
    await waitFor(() => {
      expect(screen.getByTestId('create-map-card')).toBeInTheDocument()
    })

    const user = userEvent.setup()
    await user.click(screen.getByTestId('create-map-card'))
    expect(
      screen.getByRole('heading', { name: '创建里程碑图' }),
    ).toBeInTheDocument()
  })

  // AC-4: Card links to detail page
  it('renders cards as links to /milestones/:mapId', async () => {
    setupMocks({ items: [makeMap()], total: 1 })

    renderList()
    await waitFor(() => {
      expect(screen.getByTestId('milestone-map-card-map-1')).toHaveAttribute(
        'href',
        '/milestones/map-1',
      )
    })
  })

  // Story 14: Read-only hides create button and dashed card
  it('hides create button and dashed card when user lacks permission', async () => {
    setAuth([])
    setupMocks({ items: [makeMap()], total: 1 })

    renderList()
    await waitFor(() => {
      expect(
        screen.getByTestId('milestone-map-card-map-1'),
      ).toBeInTheDocument()
    })
    expect(screen.queryByTestId('create-map-card')).not.toBeInTheDocument()
    expect(screen.queryByTestId('create-map-btn')).not.toBeInTheDocument()
  })
})
