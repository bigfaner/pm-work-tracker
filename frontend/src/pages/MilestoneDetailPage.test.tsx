import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { TooltipProvider } from '@/components/ui/tooltip'
import MilestoneDetailPage from './MilestoneDetailPage'

// Mock API modules
vi.mock('@/api/milestones', () => ({
  getMilestoneMapApi: vi.fn(),
  listMilestonesByMapApi: vi.fn(),
  getMilestoneMapTransitionsApi: vi.fn(() => Promise.resolve([])),
  updateMilestoneMapApi: vi.fn(),
  updateMilestoneApi: vi.fn(),
  createMilestoneApi: vi.fn(),
  deleteMilestoneMapApi: vi.fn(),
}))

vi.mock('@/api/teams', () => ({
  listMembersApi: vi.fn(),
}))

vi.mock('@/api/mainItems', () => ({
  listMainItemsApi: vi.fn(),
  updateMainItemApi: vi.fn(),
}))

vi.mock('@/store/team', () => ({
  useTeamStore: vi.fn(
    (selector) => selector({ currentTeamId: 'team-1' }),
  ),
}))

vi.mock('@/components/ui/toast', () => ({
  useToast: vi.fn(() => ({ addToast: vi.fn() })),
}))

vi.mock('@/hooks/usePermission', () => ({
  usePermission: vi.fn(() => true),
}))

import {
  getMilestoneMapApi,
  listMilestonesByMapApi,
  updateMilestoneMapApi,
} from '@/api/milestones'
import { listMembersApi } from '@/api/teams'

const mockMap = {
  bizKey: 'map-1',
  teamKey: 'team-1',
  creatorKey: 'user-1',
  creatorName: 'Test User',
  assigneeKey: 'user-1',
  assigneeName: 'Test User',
  mapName: 'Product MVP',
  mapDesc: 'MVP description',
  mapStatus: 'executing',
  statusName: '实施中',
  planStartDate: '2026-05-01',
  expectedEndDate: '2026-12-31',
  milestoneCount: 2,
  itemCount: 2,
  overallProgress: 40,
  milestoneSummary: [
    { bizKey: 'ms-1', name: 'Phase 1', status: 'completed', progress: 100 },
    { bizKey: 'ms-2', name: 'Phase 2', status: 'not_started', progress: 0 },
  ],
  createTime: '2026-01-01',
  dbUpdateTime: '2026-01-01',
}

const mockMilestones = [
  {
    bizKey: 'ms-1',
    teamKey: 'team-1',
    milestoneMapKey: 'map-1',
    milestoneName: 'Alpha Release',
    milestoneDesc: '',
    expectedEndDate: '2026-06-30',
    milestoneStatus: 'in_progress',
    statusName: '进行中',
    completion: 80,
    relatedMICount: 2,
    createTime: '2026-01-01',
    dbUpdateTime: '2026-01-01',
  },
  {
    bizKey: 'ms-2',
    teamKey: 'team-1',
    milestoneMapKey: 'map-1',
    milestoneName: 'Beta Release',
    milestoneDesc: '',
    expectedEndDate: '2026-09-30',
    milestoneStatus: 'not_started',
    statusName: '未开始',
    completion: 0,
    relatedMICount: 0,
    createTime: '2026-01-01',
    dbUpdateTime: '2026-01-01',
  },
]

const mockMembers = [
  {
    bizKey: '1',
    teamKey: 'team-1',
    userKey: 'user-1',
    displayName: 'Test User',
    username: 'testuser',
    role: 'pm',
    roleKey: '0',
    roleName: 'pm',
    joinedAt: '2024-01-01',
  },
]

function createQueryClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
}

function renderPage(mapId = 'map-1') {
  const qc = createQueryClient()
  return render(
    <QueryClientProvider client={qc}>
      <TooltipProvider>
        <MemoryRouter initialEntries={['/milestones/' + mapId]}>
          <Routes>
            <Route path="/milestones/:mapId" element={<MilestoneDetailPage />} />
          </Routes>
        </MemoryRouter>
      </TooltipProvider>
    </QueryClientProvider>,
  )
}

function setupMocks() {
  vi.mocked(getMilestoneMapApi).mockResolvedValue(mockMap)
  vi.mocked(listMilestonesByMapApi).mockResolvedValue({
    items: mockMilestones,
    total: mockMilestones.length,
  })
  vi.mocked(listMembersApi).mockResolvedValue(mockMembers)
}

describe('MilestoneDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setupMocks()
  })

  it('renders breadcrumb with map name', async () => {
    renderPage()
    await waitFor(() => {
      // "Product MVP" appears in breadcrumb and h1
      expect(screen.getAllByText('Product MVP').length).toBeGreaterThanOrEqual(1)
    })
    const link = screen.getByText('里程碑图').closest('a')
    expect(link).toHaveAttribute('href', '/milestones')
  })

  it('renders milestone timeline with nodes', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Release')).toBeInTheDocument()
      expect(screen.getByText('Beta Release')).toBeInTheDocument()
    })
  })

  it('renders info card with map metadata', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getAllByText('Test User').length).toBeGreaterThanOrEqual(1)
      expect(screen.getAllByText('2026-05-01').length).toBeGreaterThanOrEqual(1)
      expect(screen.getAllByText('2026-12-31').length).toBeGreaterThanOrEqual(1)
    })
  })

  it('opens edit map dialog when edit button is clicked', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Release')).toBeInTheDocument()
    })

    const editButtons = screen.getAllByRole('button', { name: /编辑/ })
    await user.click(editButtons[0])

    await waitFor(() => {
      expect(screen.getByText('编辑里程碑图')).toBeInTheDocument()
    })
  })

  it('populates edit map dialog form with current data', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Release')).toBeInTheDocument()
    })

    const editButtons = screen.getAllByRole('button', { name: /编辑/ })
    await user.click(editButtons[0])

    await waitFor(() => {
      const nameInput = screen.getByDisplayValue('Product MVP')
      expect(nameInput).toBeInTheDocument()
    })
  })

  it('calls updateMilestoneMapApi on map edit submit', async () => {
    const user = userEvent.setup()
    vi.mocked(updateMilestoneMapApi).mockResolvedValue({
      ...mockMap,
      mapName: 'Product MVP Updated',
    })
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Release')).toBeInTheDocument()
    })

    const editButtons = screen.getAllByRole('button', { name: /编辑/ })
    await user.click(editButtons[0])

    await waitFor(() => {
      expect(screen.getByText('编辑里程碑图')).toBeInTheDocument()
    })

    const nameInput = screen.getByDisplayValue('Product MVP')
    await user.clear(nameInput)
    await user.type(nameInput, 'Product MVP Updated')

    const submitBtn = screen.getByRole('button', { name: '确认' })
    await user.click(submitBtn)

    await waitFor(() => {
      expect(updateMilestoneMapApi).toHaveBeenCalledWith(
        'team-1',
        'map-1',
        expect.objectContaining({ mapName: 'Product MVP Updated' }),
      )
    })
  })

  it('renders filter bar with search input', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Release')).toBeInTheDocument()
    })
    expect(screen.getByPlaceholderText('搜索里程碑...')).toBeInTheDocument()
  })

  it('renders zoom controls', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Release')).toBeInTheDocument()
    })
    expect(screen.getByTestId('zoom-week')).toBeInTheDocument()
    expect(screen.getByTestId('zoom-month')).toBeInTheDocument()
    expect(screen.getByTestId('zoom-quarter')).toBeInTheDocument()
  })

  it('shows loading state initially', () => {
    vi.mocked(getMilestoneMapApi).mockReturnValue(new Promise(() => {}))
    vi.mocked(listMilestonesByMapApi).mockReturnValue(new Promise(() => {}))

    renderPage()
    expect(document.querySelector('.animate-pulse')).toBeTruthy()
  })
})
