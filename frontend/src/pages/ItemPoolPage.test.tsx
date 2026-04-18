import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor, within, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import ItemPoolPage from './ItemPoolPage'
import { useAuthStore } from '@/store/auth'
import { useTeamStore } from '@/store/team'
import type { User, Team, ItemPool, PageResult, TeamMemberResp, MainItem } from '@/types'

// --- Mocks ---

const mockListItemPool = vi.fn()
const mockSubmitItemPool = vi.fn()
const mockAssignItemPool = vi.fn()
const mockRejectItemPool = vi.fn()
const mockListMembers = vi.fn()
const mockListMainItems = vi.fn()

vi.mock('@/api/itemPool', () => ({
  listItemPoolApi: (...args: unknown[]) => mockListItemPool(...args),
  submitItemPoolApi: (...args: unknown[]) => mockSubmitItemPool(...args),
  assignItemPoolApi: (...args: unknown[]) => mockAssignItemPool(...args),
  rejectItemPoolApi: (...args: unknown[]) => mockRejectItemPool(...args),
}))

vi.mock('@/api/teams', () => ({
  listMembersApi: (...args: unknown[]) => mockListMembers(...args),
}))

vi.mock('@/api/mainItems', () => ({
  listMainItemsApi: (...args: unknown[]) => mockListMainItems(...args),
}))

// --- Test Data ---

const mockUser: User = {
  id: 1,
  username: 'pmuser',
  display_name: 'PM User',
  is_super_admin: false,
  can_create_team: false,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const memberUser: User = {
  ...mockUser,
  id: 10,
  username: 'member',
  display_name: 'Member User',
}

const mockTeam: Team = {
  id: 1,
  name: 'Team Alpha',
  description: '',
  pm_id: 1,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const mockMembers: TeamMemberResp[] = [
  { userId: 1, displayName: 'PM User', username: 'pmuser', role: 'pm', joinedAt: '2024-01-01' },
  { userId: 10, displayName: 'Member User', username: 'member', role: 'member', joinedAt: '2024-01-01' },
]

function makeItemPool(overrides: Partial<ItemPool> = {}): ItemPool {
  return {
    id: 1,
    team_id: 1,
    title: 'Test Pool Item',
    background: 'Some background info',
    expected_output: 'Expected output text',
    submitter_id: 10,
    status: '待分配',
    assigned_main_id: null,
    assigned_sub_id: null,
    assignee_id: null,
    reject_reason: '',
    reviewed_at: null,
    reviewer_id: null,
    created_at: '2024-06-01T10:00:00Z',
    updated_at: '2024-06-01T10:00:00Z',
    ...overrides,
  }
}

const emptyPageResult: PageResult<ItemPool> = {
  items: [],
  total: 0,
  page: 1,
  pageSize: 20,
}

// --- Helpers ---

function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })
}

function renderPage(user: User = mockUser, teamId: number | null = 1) {
  useAuthStore.getState().setAuth('token', user)
  const qc = createQueryClient()
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={['/item-pool']}>
        <Routes>
          <Route path="/item-pool" element={<ItemPoolPage />} />
          <Route path="/items/:mainItemId" element={<div data-testid="detail-page">Detail</div>} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

async function openAndSelectOption(selectTestId: string, optionText: string) {
  const selectEl = screen.getByTestId(selectTestId)
  const selector = selectEl.querySelector('.ant-select-selector')!
  fireEvent.mouseDown(selector)
  await waitFor(() => {
    const options = document.querySelectorAll('.ant-select-item-option')
    const match = Array.from(options).find((el) => el.textContent === optionText)
    expect(match).toBeTruthy()
  })
  const options = document.querySelectorAll('.ant-select-item-option')
  const match = Array.from(options).find((el) => el.textContent === optionText)!
  fireEvent.click(match)
}

// --- Tests ---

describe('ItemPoolPage', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth()
    useAuthStore.getState().setAuth('token', mockUser)
    useTeamStore.getState().setTeams([mockTeam])
    useTeamStore.getState().setCurrentTeam(1)

    mockListItemPool.mockResolvedValue(emptyPageResult)
    mockListMembers.mockResolvedValue(mockMembers)
    mockListMainItems.mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 20 })
    mockSubmitItemPool.mockReset()
    mockAssignItemPool.mockReset()
    mockRejectItemPool.mockReset()
  })

  // --- Basic rendering ---

  it('renders page with data-testid', () => {
    renderPage()
    expect(screen.getByTestId('item-pool-page')).toBeInTheDocument()
  })

  it('renders page title 事项池', () => {
    renderPage()
    expect(screen.getByText('事项池')).toBeInTheDocument()
  })

  // --- Submit button visibility ---

  it('shows submit button for all team members', async () => {
    renderPage(memberUser)
    await waitFor(() => {
      expect(screen.getByTestId('submit-pool-btn')).toBeInTheDocument()
    })
  })

  // --- Filter bar ---

  it('renders status filter radio group with 4 options', () => {
    renderPage()
    expect(screen.getByTestId('status-filter')).toBeInTheDocument()
    expect(screen.getByText('全部')).toBeInTheDocument()
    expect(screen.getByText('待分配')).toBeInTheDocument()
    expect(screen.getByText('已分配')).toBeInTheDocument()
    expect(screen.getByText('已拒绝')).toBeInTheDocument()
  })

  it('calls API with status filter when radio option clicked', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(mockListItemPool).toHaveBeenCalled()
    })
    mockListItemPool.mockClear()

    await user.click(screen.getByText('待分配'))
    await waitFor(() => {
      expect(mockListItemPool).toHaveBeenCalledWith(1, expect.objectContaining({ status: '待分配' }))
    })
  })

  // --- Loading state ---

  it('shows skeleton loading state while fetching', () => {
    let resolvePromise!: (v: unknown) => void
    mockListItemPool.mockReturnValue(new Promise((resolve) => { resolvePromise = resolve }))
    renderPage()
    expect(screen.getByTestId('pool-skeleton')).toBeInTheDocument()
    resolvePromise(emptyPageResult)
  })

  // --- Empty states ---

  it('shows empty state when no items exist', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('pool-empty')).toBeInTheDocument()
    })
    expect(screen.getByText('事项池暂无内容')).toBeInTheDocument()
  })

  it('shows filter empty state when filter yields no results', async () => {
    const items = [makeItemPool({ status: '待分配' })]
    mockListItemPool.mockResolvedValue({ items, total: 1, page: 1, pageSize: 20 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('Test Pool Item')).toBeInTheDocument()
    })

    mockListItemPool.mockResolvedValue(emptyPageResult)
    await userEvent.setup().click(screen.getByText('已拒绝'))

    await waitFor(() => {
      expect(screen.getByTestId('pool-filter-empty')).toBeInTheDocument()
    })
    expect(screen.getByText('该状态下暂无事项')).toBeInTheDocument()
  })

  // --- Pool item cards ---

  it('renders pool item cards with title, status tag, submitter and time', async () => {
    const items = [makeItemPool({
      id: 1,
      title: 'Important Task',
      status: '待分配',
      submitter_id: 10,
      created_at: '2024-06-01T10:00:00Z',
    })]
    mockListItemPool.mockResolvedValue({ items, total: 1, page: 1, pageSize: 20 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('Important Task')).toBeInTheDocument()
    })
    expect(screen.getByTestId('pool-card-1')).toBeInTheDocument()
  })

  // --- 待分配 card: blue border + action buttons ---

  it('shows assign and reject buttons for 待分配 items (PM only)', async () => {
    const items = [makeItemPool({ id: 1, status: '待分配' })]
    mockListItemPool.mockResolvedValue({ items, total: 1, page: 1, pageSize: 20 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByTestId('assign-btn-1')).toBeInTheDocument()
      expect(screen.getByTestId('reject-btn-1')).toBeInTheDocument()
    })
  })

  it('hides assign/reject buttons for non-PM users', async () => {
    const items = [makeItemPool({ id: 1, status: '待分配' })]
    mockListItemPool.mockResolvedValue({ items, total: 1, page: 1, pageSize: 20 })
    renderPage(memberUser)

    await waitFor(() => {
      expect(screen.getByText('Test Pool Item')).toBeInTheDocument()
    })
    expect(screen.queryByTestId('assign-btn-1')).not.toBeInTheDocument()
    expect(screen.queryByTestId('reject-btn-1')).not.toBeInTheDocument()
  })

  // --- 已分配 card: grayed out, mounted info ---

  it('shows 已分配 card grayed out with mounted info', async () => {
    const items = [makeItemPool({
      id: 2,
      status: '已分配',
      assigned_main_id: 100,
      assignee_id: 10,
    })]
    mockListItemPool.mockResolvedValue({ items, total: 1, page: 1, pageSize: 20 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByTestId('pool-card-2')).toBeInTheDocument()
    })
    // Grayed out via opacity
    const card = screen.getByTestId('pool-card-2')
    expect(card).toHaveStyle({ opacity: '0.7' })
  })

  // --- 已拒绝 card: grayed out, reject reason ---

  it('shows 已拒绝 card grayed out with reject reason', async () => {
    const items = [makeItemPool({
      id: 3,
      status: '已拒绝',
      reject_reason: 'Not relevant',
    })]
    mockListItemPool.mockResolvedValue({ items, total: 1, page: 1, pageSize: 20 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByTestId('pool-card-3')).toBeInTheDocument()
    })
    const card = screen.getByTestId('pool-card-3')
    expect(card).toHaveStyle({ opacity: '0.7' })
    expect(screen.getByText(/Not relevant/)).toBeInTheDocument()
  })

  // --- Submit modal ---

  it('opens submit modal on button click', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('submit-pool-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('submit-pool-btn'))
    expect(screen.getByTestId('submit-pool-modal')).toBeInTheDocument()
  })

  it('submit modal has title, background, expectedOutput fields', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('submit-pool-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('submit-pool-btn'))
    expect(screen.getByTestId('form-pool-title')).toBeInTheDocument()
    expect(screen.getByTestId('form-pool-background')).toBeInTheDocument()
    expect(screen.getByTestId('form-pool-expected-output')).toBeInTheDocument()
  })

  it('submits pool item form and calls API', async () => {
    const user = userEvent.setup()
    mockSubmitItemPool.mockResolvedValue(makeItemPool({ id: 99 }))
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('submit-pool-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('submit-pool-btn'))

    await user.type(screen.getByTestId('form-pool-title'), 'New Pool Item')
    await user.click(screen.getByTestId('submit-pool-ok'))

    await waitFor(() => {
      expect(mockSubmitItemPool).toHaveBeenCalledWith(1, expect.objectContaining({
        title: 'New Pool Item',
      }))
    })
  })

  it('shows validation error for empty title on submit', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('submit-pool-btn')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('submit-pool-btn'))

    await user.click(screen.getByTestId('submit-pool-ok'))

    await waitFor(() => {
      expect(screen.getByText('请填写标题')).toBeInTheDocument()
    })
  })

  // --- Assign modal ---

  it('opens assign modal on assign button click', async () => {
    const items = [makeItemPool({ id: 1, status: '待分配' })]
    mockListItemPool.mockResolvedValue({ items, total: 1, page: 1, pageSize: 20 })
    mockListMainItems.mockResolvedValue({
      items: [{
        id: 100, team_id: 1, code: 'MI-0100', title: 'Main Item A',
        priority: 'P2', proposer_id: 1, assignee_id: 1,
        start_date: null, expected_end_date: null, actual_end_date: null,
        status: '进行中', completion: 50, is_key_item: false,
        delay_count: 0, archived_at: null,
        created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z',
      }],
      total: 1, page: 1, pageSize: 20,
    })

    const user = userEvent.setup()
    renderPage()

    await waitFor(() => {
      expect(screen.getByTestId('assign-btn-1')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('assign-btn-1'))

    await waitFor(() => {
      expect(screen.getByTestId('assign-modal')).toBeInTheDocument()
    })
    expect(screen.getByTestId('form-assign-main-item')).toBeInTheDocument()
    expect(screen.getByTestId('form-assign-assignee')).toBeInTheDocument()
  })

  it('submits assign form and calls API', async () => {
    const items = [makeItemPool({ id: 1, status: '待分配' })]
    mockListItemPool.mockResolvedValue({ items, total: 1, page: 1, pageSize: 20 })
    const mainItem: MainItem = {
      id: 100, team_id: 1, code: 'MI-0100', title: 'Main Item A',
      priority: 'P2', proposer_id: 1, assignee_id: 1,
      start_date: null, expected_end_date: null, actual_end_date: null,
      status: '进行中', completion: 50, is_key_item: false,
      delay_count: 0, archived_at: null,
      created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z',
    }
    mockListMainItems.mockResolvedValue({ items: [mainItem], total: 1, page: 1, pageSize: 20 })
    mockAssignItemPool.mockResolvedValue({ subItemId: 200 })

    const user = userEvent.setup()
    renderPage()

    await waitFor(() => {
      expect(screen.getByTestId('assign-btn-1')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('assign-btn-1'))

    await waitFor(() => {
      expect(screen.getByTestId('assign-modal')).toBeInTheDocument()
    })

    // Select main item
    await openAndSelectOption('form-assign-main-item', 'MI-0100 Main Item A')
    // Select assignee
    await openAndSelectOption('form-assign-assignee', 'PM User')

    await user.click(screen.getByTestId('assign-ok'))

    await waitFor(() => {
      expect(mockAssignItemPool).toHaveBeenCalledWith(1, 1, expect.objectContaining({
        mainItemId: 100,
        assigneeId: 1,
      }))
    })
  })

  // --- Reject modal ---

  it('opens reject modal on reject button click', async () => {
    const items = [makeItemPool({ id: 1, status: '待分配' })]
    mockListItemPool.mockResolvedValue({ items, total: 1, page: 1, pageSize: 20 })

    const user = userEvent.setup()
    renderPage()

    await waitFor(() => {
      expect(screen.getByTestId('reject-btn-1')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('reject-btn-1'))

    await waitFor(() => {
      expect(screen.getByTestId('reject-modal')).toBeInTheDocument()
    })
    expect(screen.getByTestId('form-reject-reason')).toBeInTheDocument()
  })

  it('submits reject form and calls API', async () => {
    const items = [makeItemPool({ id: 1, status: '待分配' })]
    mockListItemPool.mockResolvedValue({ items, total: 1, page: 1, pageSize: 20 })
    mockRejectItemPool.mockResolvedValue(makeItemPool({ id: 1, status: '已拒绝' }))

    const user = userEvent.setup()
    renderPage()

    await waitFor(() => {
      expect(screen.getByTestId('reject-btn-1')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('reject-btn-1'))

    await waitFor(() => {
      expect(screen.getByTestId('reject-modal')).toBeInTheDocument()
    })

    await user.type(screen.getByTestId('form-reject-reason'), 'Not suitable')
    await user.click(screen.getByTestId('reject-ok'))

    await waitFor(() => {
      expect(mockRejectItemPool).toHaveBeenCalledWith(1, 1, { reason: 'Not suitable' })
    })
  })

  it('shows validation error for empty reject reason', async () => {
    const items = [makeItemPool({ id: 1, status: '待分配' })]
    mockListItemPool.mockResolvedValue({ items, total: 1, page: 1, pageSize: 20 })

    const user = userEvent.setup()
    renderPage()

    await waitFor(() => {
      expect(screen.getByTestId('reject-btn-1')).toBeInTheDocument()
    })
    await user.click(screen.getByTestId('reject-btn-1'))

    await waitFor(() => {
      expect(screen.getByTestId('reject-modal')).toBeInTheDocument()
    })

    await user.click(screen.getByTestId('reject-ok'))

    await waitFor(() => {
      expect(screen.getByText('请填写拒绝原因')).toBeInTheDocument()
    })
  })

  // --- Card body expandable text ---

  it('renders background and expected output in card body', async () => {
    const items = [makeItemPool({
      id: 1,
      background: 'Some background text here',
      expected_output: 'Expected output text here',
    })]
    mockListItemPool.mockResolvedValue({ items, total: 1, page: 1, pageSize: 20 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('Some background text here')).toBeInTheDocument()
      expect(screen.getByText('Expected output text here')).toBeInTheDocument()
    })
  })

  // --- Multiple cards rendering ---

  it('renders multiple pool item cards', async () => {
    const items = [
      makeItemPool({ id: 1, title: 'Item A', status: '待分配' }),
      makeItemPool({ id: 2, title: 'Item B', status: '已分配' }),
      makeItemPool({ id: 3, title: 'Item C', status: '已拒绝', reject_reason: 'No' }),
    ]
    mockListItemPool.mockResolvedValue({ items, total: 3, page: 1, pageSize: 20 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('Item A')).toBeInTheDocument()
      expect(screen.getByText('Item B')).toBeInTheDocument()
      expect(screen.getByText('Item C')).toBeInTheDocument()
    })
  })
})
