import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { TooltipProvider } from '@/components/ui/tooltip'
import MilestoneDetailPanel from './MilestoneDetailPanel'
import type { Milestone } from '@/types'

// Mock API modules
vi.mock('@/api/milestones', () => ({
  getMilestoneApi: vi.fn(),
  deleteMilestoneApi: vi.fn(),
  getMilestoneTransitionsApi: vi.fn(() => Promise.resolve({ transitions: [] })),
  changeMilestoneStatusApi: vi.fn(),
}))

vi.mock('@/api/mainItems', () => ({
  listMainItemsApi: vi.fn(),
  updateMainItemApi: vi.fn(),
}))

vi.mock('@/store/team', () => ({
  useTeamStore: vi.fn(
    (selector: (s: { currentTeamId: string | null }) => unknown) =>
      selector({ currentTeamId: 'team-1' }),
  ),
}))

vi.mock('@/components/ui/toast', () => ({
  useToast: vi.fn(() => ({ addToast: vi.fn() })),
}))

const mockNavigate = vi.fn()
vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}))

import { getMilestoneApi, deleteMilestoneApi } from '@/api/milestones'
import { listMainItemsApi, updateMainItemApi } from '@/api/mainItems'

const mockMilestone: Milestone = {
  bizKey: 'ms-1',
  teamKey: 'team-1',
  milestoneMapKey: 'map-1',
  milestoneName: 'MVP 发布',
  milestoneDesc: '完成产品MVP版本的核心功能开发',
  expectedEndDate: '2026-06-30',
  milestoneStatus: 'not_started',
  statusName: '未开始',
  completion: 80,
  relatedMICount: 3,
  createTime: '2026-01-01',
  dbUpdateTime: '2026-01-01',
}

const mockMIs = {
  items: [
    {
      bizKey: 'mi-1',
      teamKey: 'team-1',
      code: 'MI-0001',
      title: '需求分析',
      priority: 'P1',
      proposerKey: 'user-1',
      assigneeKey: 'user-1',
      planStartDate: '2026-01-01',
      expectedEndDate: '2026-03-01',
      actualEndDate: null,
      itemStatus: 'progressing',
      statusName: '进行中',
      completion: 60,
      createTime: '2026-01-01',
      dbUpdateTime: '2026-01-01',
    },
    {
      bizKey: 'mi-2',
      teamKey: 'team-1',
      code: 'MI-0003',
      title: 'UI设计',
      priority: 'P1',
      proposerKey: 'user-1',
      assigneeKey: 'user-2',
      planStartDate: '2026-02-01',
      expectedEndDate: '2026-04-01',
      actualEndDate: '2026-04-01',
      itemStatus: 'completed',
      statusName: '已完成',
      completion: 100,
      createTime: '2026-01-15',
      dbUpdateTime: '2026-04-01',
    },
    {
      bizKey: 'mi-3',
      teamKey: 'team-1',
      code: 'MI-0005',
      title: 'API开发',
      priority: 'P2',
      proposerKey: 'user-1',
      assigneeKey: 'user-1',
      planStartDate: '2026-03-01',
      expectedEndDate: '2026-06-30',
      actualEndDate: null,
      itemStatus: 'progressing',
      statusName: '进行中',
      completion: 80,
      createTime: '2026-02-01',
      dbUpdateTime: '2026-02-01',
    },
  ],
  total: 3,
  page: 1,
  size: 200,
}

function createQueryClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
}

function renderPanel(
  overrides: Partial<Parameters<typeof MilestoneDetailPanel>[0]> = {},
  milestoneOverride?: Milestone,
) {
  const onClose = vi.fn()
  const onEdit = vi.fn()
  const onQuickAdd = vi.fn()
  const onDeleted = vi.fn()

  vi.mocked(getMilestoneApi).mockResolvedValue(
    milestoneOverride ?? mockMilestone,
  )
  vi.mocked(listMainItemsApi).mockResolvedValue(mockMIs)

  const result = render(
    <QueryClientProvider client={createQueryClient()}>
      <TooltipProvider>
        <MilestoneDetailPanel
          open={true}
          onClose={onClose}
          milestoneId="ms-1"
          onEdit={onEdit}
          onQuickAdd={onQuickAdd}
          onDeleted={onDeleted}
          {...overrides}
        />
      </TooltipProvider>
    </QueryClientProvider>,
  )

  return { onClose, onEdit, onQuickAdd, onDeleted, ...result }
}

describe('MilestoneDetailPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockNavigate.mockClear()
  })

  // AC-1: Panel slides in; closes on overlay click, Escape, or X button
  it('renders nothing when open is false', () => {
    render(
      <QueryClientProvider client={createQueryClient()}>
        <TooltipProvider>
          <MilestoneDetailPanel
            open={false}
            onClose={vi.fn()}
            milestoneId="ms-1"
            onEdit={vi.fn()}
            onQuickAdd={vi.fn()}
          />
        </TooltipProvider>
      </QueryClientProvider>,
    )
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('renders panel with w-[540px] when open', () => {
    renderPanel()
    const dialog = screen.getByRole('dialog')
    expect(dialog).toBeInTheDocument()
    expect(dialog.className).toContain('w-[540px]')
  })

  it('closes on overlay click', async () => {
    const user = userEvent.setup()
    const { onClose } = renderPanel()
    const overlay = document.querySelector('.bg-black\\/20')
    expect(overlay).toBeInTheDocument()
    await user.click(overlay!)
    expect(onClose).toHaveBeenCalled()
  })

  it('closes on X button click', async () => {
    const user = userEvent.setup()
    const { onClose } = renderPanel()
    await user.click(screen.getByLabelText('关闭面板'))
    expect(onClose).toHaveBeenCalled()
  })

  it('closes on Escape key', async () => {
    const { onClose } = renderPanel()
    fireEvent.keyDown(document, { key: 'Escape' })
    expect(onClose).toHaveBeenCalled()
  })

  // AC-2: Displays name, description, expected end date, progress, MI list
  it('displays milestone name', async () => {
    renderPanel()
    await waitFor(() => {
      expect(screen.getByText('MVP 发布')).toBeInTheDocument()
    })
  })

  it('displays description with line-clamp', async () => {
    renderPanel()
    await waitFor(() => {
      expect(
        screen.getByText('完成产品MVP版本的核心功能开发'),
      ).toBeInTheDocument()
    })
  })

  it('displays expected end date', async () => {
    renderPanel()
    await waitFor(() => {
      expect(screen.getByText('2026-06-30')).toBeInTheDocument()
    })
  })

  it('displays progress percentage', async () => {
    renderPanel()
    await waitFor(() => {
      // Multiple elements show "80%" (progress label + MI completion), use getAllByText
      const elements = screen.getAllByText('80%')
      expect(elements.length).toBeGreaterThanOrEqual(1)
    })
  })

  it('displays description label', async () => {
    renderPanel()
    await waitFor(() => {
      expect(screen.getByText('描述')).toBeInTheDocument()
    })
  })

  it('displays plan end date label', async () => {
    renderPanel()
    await waitFor(() => {
      expect(screen.getByText('计划完成时间')).toBeInTheDocument()
    })
  })

  it('displays progress label', async () => {
    renderPanel()
    await waitFor(() => {
      expect(screen.getByText('进度')).toBeInTheDocument()
    })
  })

  // AC-3: Status switch via StatusTransitionDropdown
  it('renders StatusTransitionDropdown', async () => {
    renderPanel()
    await waitFor(() => {
      expect(screen.getByText('未开始')).toBeInTheDocument()
    })
  })

  // AC-4: MI list with unbind
  it('shows related MI list with count', async () => {
    renderPanel()
    await waitFor(() => {
      expect(screen.getByText('关联事项 (3)')).toBeInTheDocument()
    })
  })

  it('shows unbind button on hover for non-terminal MI', async () => {
    renderPanel()
    await waitFor(() => {
      expect(screen.getByText('MI-0001')).toBeInTheDocument()
    })
    // The X button should exist in DOM but be opacity-0
    const unbindButtons = screen.getAllByLabelText(/解绑事项/)
    expect(unbindButtons.length).toBeGreaterThan(0)
  })

  it('hides unbind button for terminal MI (completed)', async () => {
    renderPanel()
    await waitFor(() => {
      expect(screen.getByText('MI-0003')).toBeInTheDocument()
    })
    // MI-0003 is completed, should not have unbind button
    expect(screen.queryByLabelText('解绑事项 MI-0003')).not.toBeInTheDocument()
  })

  it('calls unbind API when X is clicked', async () => {
    vi.mocked(updateMainItemApi).mockResolvedValue({} as never)
    const user = userEvent.setup()
    renderPanel()
    await waitFor(() => {
      expect(screen.getByText('MI-0001')).toBeInTheDocument()
    })
    const unbindBtn = screen.getByLabelText('解绑事项 MI-0001')
    await user.click(unbindBtn)
    expect(updateMainItemApi).toHaveBeenCalledWith('team-1', 'mi-1', {
      milestoneKey: '',
    })
  })

  // AC-5: Quick add
  it('renders + 添加 button', async () => {
    renderPanel()
    await waitFor(() => {
      expect(screen.getByText('添加')).toBeInTheDocument()
    })
  })

  it('calls onQuickAdd when + 添加 is clicked', async () => {
    const user = userEvent.setup()
    const { onQuickAdd } = renderPanel()
    await waitFor(() => {
      expect(screen.getByText('添加')).toBeInTheDocument()
    })
    await user.click(screen.getByText('添加'))
    expect(onQuickAdd).toHaveBeenCalledWith(mockMilestone)
  })

  // AC-6: Delete button visible only for not_started/cancelled
  it('shows delete button for not_started status', async () => {
    renderPanel()
    await waitFor(() => {
      expect(screen.getByText('删除里程碑')).toBeInTheDocument()
    })
  })

  it('hides delete button for in_progress status', async () => {
    renderPanel(undefined, {
      ...mockMilestone,
      milestoneStatus: 'in_progress',
      statusName: '进行中',
    })
    await waitFor(() => {
      // Wait for milestone name to appear (confirms data loaded)
      expect(screen.getByText('MVP 发布')).toBeInTheDocument()
    })
    expect(screen.queryByText('删除里程碑')).not.toBeInTheDocument()
  })

  it('shows delete button for cancelled status', async () => {
    renderPanel(undefined, {
      ...mockMilestone,
      milestoneStatus: 'cancelled',
      statusName: '已取消',
    })
    await waitFor(() => {
      // Wait for milestone name to confirm data loaded
      expect(screen.getByText('MVP 发布')).toBeInTheDocument()
    })
    expect(screen.getByText('删除里程碑')).toBeInTheDocument()
  })

  it('opens ConfirmDialog when delete is clicked', async () => {
    const user = userEvent.setup()
    renderPanel()
    await waitFor(() => {
      expect(screen.getByText('删除里程碑')).toBeInTheDocument()
    })
    await user.click(screen.getByText('删除里程碑'))
    await waitFor(() => {
      expect(screen.getByText(/确定删除里程碑 MVP 发布/)).toBeInTheDocument()
    })
    expect(screen.getByText(/关联的 3 个事项将解除绑定/)).toBeInTheDocument()
  })

  it('calls delete API when confirm is clicked', async () => {
    vi.mocked(deleteMilestoneApi).mockResolvedValue({ message: 'ok' })
    const user = userEvent.setup()
    renderPanel()
    await waitFor(() => {
      expect(screen.getByText('删除里程碑')).toBeInTheDocument()
    })
    await user.click(screen.getByText('删除里程碑'))
    await waitFor(() => {
      expect(screen.getByText('确认删除')).toBeInTheDocument()
    })
    await user.click(screen.getByText('确认删除'))
    await waitFor(() => {
      expect(deleteMilestoneApi).toHaveBeenCalledWith('team-1', 'ms-1')
    })
  })

  // Cancelled milestone: grey styling, no MI list, no + 添加
  it('shows cancelled styling and hides MI list when cancelled', async () => {
    renderPanel(undefined, {
      ...mockMilestone,
      milestoneStatus: 'cancelled',
      statusName: '已取消',
    })
    await waitFor(() => {
      // Wait for name to confirm data loaded
      expect(screen.getByText('MVP 发布')).toBeInTheDocument()
    })
    // Name should have line-through
    const name = screen.getByText('MVP 发布')
    expect(name.className).toContain('line-through')
    // MI list should be hidden
    expect(screen.queryByText(/关联事项/)).not.toBeInTheDocument()
    // + 添加 should be hidden
    expect(screen.queryByText('添加')).not.toBeInTheDocument()
  })

  // Terminal milestone: no edit button, status dropdown disabled
  it('hides edit button for terminal milestone (completed)', async () => {
    renderPanel(undefined, {
      ...mockMilestone,
      milestoneStatus: 'completed',
      statusName: '已完成',
    })
    await waitFor(() => {
      expect(screen.getByText('MVP 发布')).toBeInTheDocument()
    })
    expect(screen.queryByLabelText('编辑里程碑')).not.toBeInTheDocument()
  })

  // Loading state
  it('shows skeleton while loading', () => {
    vi.mocked(getMilestoneApi).mockReturnValue(new Promise(() => {})) // never resolves
    renderPanel()
    expect(document.querySelector('.animate-pulse')).toBeInTheDocument()
  })

  // Shows loading skeleton when fetching
  it('shows loading skeleton before data loads', () => {
    vi.mocked(getMilestoneApi).mockReturnValue(new Promise(() => {}))
    renderPanel()
    expect(document.querySelector('.animate-pulse')).toBeInTheDocument()
  })

  // DnD: MI rows are draggable for non-terminal items
  describe('MI drag support', () => {
    it('non-terminal MI row is draggable and has grab cursor', async () => {
      renderPanel()
      await waitFor(() => {
        expect(screen.getByTestId('mi-drag-mi-1')).toBeInTheDocument()
      })
      const miRow = screen.getByTestId('mi-drag-mi-1')
      expect(miRow.getAttribute('draggable')).toBe('true')
      expect(miRow.className).toContain('cursor-grab')
    })

    it('terminal MI row (completed) is not draggable', async () => {
      renderPanel()
      await waitFor(() => {
        expect(screen.getByTestId('mi-drag-mi-2')).toBeInTheDocument()
      })
      const miRow = screen.getByTestId('mi-drag-mi-2')
      expect(miRow.getAttribute('draggable')).toBe('false')
    })

    it('sets window.__dragMI on dragStart', async () => {
      renderPanel()
      await waitFor(() => {
        expect(screen.getByTestId('mi-drag-mi-1')).toBeInTheDocument()
      })
      const miRow = screen.getByTestId('mi-drag-mi-1')
      fireEvent.dragStart(miRow)
      expect((window as unknown as Record<string, unknown>).__dragMI).toEqual({
        miBizKey: 'mi-1',
        miCode: 'MI-0001',
        sourceMilestoneKey: 'ms-1',
      })
    })

    it('clears window.__dragMI after dragEnd', async () => {
      renderPanel()
      await waitFor(() => {
        expect(screen.getByTestId('mi-drag-mi-1')).toBeInTheDocument()
      })
      const miRow = screen.getByTestId('mi-drag-mi-1')
      fireEvent.dragStart(miRow)
      expect((window as unknown as Record<string, unknown>).__dragMI).toBeTruthy()

      fireEvent.dragEnd(miRow)
      // Wait for the 200ms setTimeout to clear the drag data
      await waitFor(
        () => {
          expect((window as unknown as Record<string, unknown>).__dragMI).toBeUndefined()
        },
        { timeout: 1000 },
      )
    })
  })

  // AC-4: Click MI item navigates to /items/:mainItemId
  it('navigates to MI detail when MI row is clicked', async () => {
    const user = userEvent.setup()
    renderPanel()
    await waitFor(() => {
      expect(screen.getByText('MI-0001')).toBeInTheDocument()
    })
    const miRow = screen.getByTestId('mi-drag-mi-1')
    await user.click(miRow)
    expect(mockNavigate).toHaveBeenCalledWith('/items/mi-1')
  })

  it('does not navigate when unbind X button is clicked', async () => {
    vi.mocked(updateMainItemApi).mockResolvedValue({} as never)
    const user = userEvent.setup()
    renderPanel()
    await waitFor(() => {
      expect(screen.getByText('MI-0001')).toBeInTheDocument()
    })
    const unbindBtn = screen.getByLabelText('解绑事项 MI-0001')
    await user.click(unbindBtn)
    expect(mockNavigate).not.toHaveBeenCalled()
    expect(updateMainItemApi).toHaveBeenCalled()
  })
})
