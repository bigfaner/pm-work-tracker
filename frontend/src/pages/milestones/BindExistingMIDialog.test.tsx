import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import BindExistingMIDialog from './BindExistingMIDialog'

vi.mock('@/api/mainItems', () => ({
  listMainItemsApi: vi.fn(),
  updateMainItemApi: vi.fn(),
}))

vi.mock('@/components/ui/toast', () => ({
  useToast: vi.fn(() => ({ addToast: vi.fn() })),
}))

import { listMainItemsApi, updateMainItemApi } from '@/api/mainItems'

const mockUnboundItems = {
  items: [
    {
      bizKey: 'mi-10',
      teamKey: 'team-1',
      code: 'MI-0010',
      title: '未绑定事项A',
      priority: 'P1',
      proposerKey: 'user-1',
      assigneeKey: null,
      planStartDate: null,
      expectedEndDate: null,
      actualEndDate: null,
      itemStatus: 'not_started',
      statusName: '未开始',
      completion: 0,
      milestoneKey: null,
      createTime: '2026-01-01',
      dbUpdateTime: '2026-01-01',
    },
    {
      bizKey: 'mi-11',
      teamKey: 'team-1',
      code: 'MI-0011',
      title: '未绑定事项B',
      priority: 'P2',
      proposerKey: 'user-1',
      assigneeKey: null,
      planStartDate: null,
      expectedEndDate: null,
      actualEndDate: null,
      itemStatus: 'progressing',
      statusName: '进行中',
      completion: 30,
      milestoneKey: null,
      createTime: '2026-02-01',
      dbUpdateTime: '2026-02-01',
    },
  ],
  total: 2,
  page: 1,
  size: 200,
}

function createQueryClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
}

function renderDialog(overrides: Record<string, unknown> = {}) {
  const onOpenChange = vi.fn()

  vi.mocked(listMainItemsApi).mockResolvedValue(mockUnboundItems)
  vi.mocked(updateMainItemApi).mockResolvedValue({} as never)

  const result = render(
    <QueryClientProvider client={createQueryClient()}>
      <BindExistingMIDialog
        open={true}
        onOpenChange={onOpenChange}
        teamId="team-1"
        milestoneBizKey="ms-1"
        milestoneName="MVP 发布"
        {...overrides}
      />
    </QueryClientProvider>,
  )

  return { onOpenChange, ...result }
}

describe('BindExistingMIDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders nothing when open is false', () => {
    render(
      <QueryClientProvider client={createQueryClient()}>
        <BindExistingMIDialog
          open={false}
          onOpenChange={vi.fn()}
          teamId="team-1"
          milestoneBizKey="ms-1"
          milestoneName="MVP 发布"
        />
      </QueryClientProvider>,
    )
    expect(screen.queryByText(/关联已有事项/)).not.toBeInTheDocument()
  })

  it('renders dialog with milestone name in title', async () => {
    renderDialog()
    await waitFor(() => {
      expect(screen.getByText(/关联已有事项到「MVP 发布」/)).toBeInTheDocument()
    })
  })

  it('renders unbound items as checkboxes', async () => {
    renderDialog()
    await waitFor(() => {
      expect(screen.getByText('MI-0010 - 未绑定事项A')).toBeInTheDocument()
      expect(screen.getByText('MI-0011 - 未绑定事项B')).toBeInTheDocument()
    })
  })

  it('shows empty state when no unbound items', async () => {
    const onOpenChange = vi.fn()
    vi.mocked(listMainItemsApi).mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      size: 200,
    })
    render(
      <QueryClientProvider client={createQueryClient()}>
        <BindExistingMIDialog
          open={true}
          onOpenChange={onOpenChange}
          teamId="team-1"
          milestoneBizKey="ms-1"
          milestoneName="MVP 发布"
        />
      </QueryClientProvider>,
    )
    await waitFor(() => {
      expect(screen.getByText('暂无可关联的事项')).toBeInTheDocument()
    })
  })

  it('disables submit when nothing is selected', async () => {
    renderDialog()
    await waitFor(() => {
      expect(screen.getByText(/关联 \(/)).toBeInTheDocument()
    })
    const submitBtn = screen.getByText(/关联 \(/)
    expect(submitBtn).toBeDisabled()
  })

  it('selects items and submits', async () => {
    const user = userEvent.setup()
    renderDialog()
    await waitFor(() => {
      expect(screen.getByText('MI-0010 - 未绑定事项A')).toBeInTheDocument()
    })
    await user.click(screen.getByText('MI-0010 - 未绑定事项A'))
    await user.click(screen.getByText('MI-0011 - 未绑定事项B'))

    const submitBtn = screen.getByText('关联 (2)')
    expect(submitBtn).not.toBeDisabled()
    await user.click(submitBtn)

    await waitFor(() => {
      expect(updateMainItemApi).toHaveBeenCalledWith('team-1', 'mi-10', {
        milestoneKey: 'ms-1',
      })
      expect(updateMainItemApi).toHaveBeenCalledWith('team-1', 'mi-11', {
        milestoneKey: 'ms-1',
      })
    })
  })

  it('closes dialog on cancel', async () => {
    const user = userEvent.setup()
    const { onOpenChange } = renderDialog()
    await waitFor(() => {
      expect(screen.getByText('取消')).toBeInTheDocument()
    })
    await user.click(screen.getByText('取消'))
    expect(onOpenChange).toHaveBeenCalledWith(false)
  })

  it('fetches unbound items with milestoneKey=unassigned', async () => {
    renderDialog()
    await waitFor(() => {
      expect(listMainItemsApi).toHaveBeenCalledWith('team-1', {
        milestoneKey: 'unassigned',
        pageSize: 200,
      })
    })
  })
})
