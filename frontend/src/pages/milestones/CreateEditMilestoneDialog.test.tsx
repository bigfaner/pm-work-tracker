import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import CreateEditMilestoneDialog from './CreateEditMilestoneDialog'

function renderDialog(ui: React.ReactElement) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={qc}>{ui}</QueryClientProvider>,
  )
}

describe('CreateEditMilestoneDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders create mode with empty form', () => {
    renderDialog(
      <CreateEditMilestoneDialog
        open={true}
        onOpenChange={() => {}}
        mode="create"
        onSubmit={() => {}}
        isPending={false}
      />,
    )

    expect(screen.getByText('创建里程碑')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('请输入里程碑名称')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '确认' })).toBeInTheDocument()
  })

  it('renders edit mode with pre-filled data', () => {
    renderDialog(
      <CreateEditMilestoneDialog
        open={true}
        onOpenChange={() => {}}
        mode="edit"
        milestone={{
          bizKey: 'ms-1',
          teamKey: 'team-1',
          milestoneMapKey: 'map-1',
          milestoneName: 'M1 需求分析',
          expectedEndDate: '2026-06-30',
          milestoneStatus: 'in_progress',
          statusName: '进行中',
          completion: 60,
          relatedMICount: 3,
          createTime: '2026-05-01T00:00:00Z',
          dbUpdateTime: '2026-05-01T00:00:00Z',
        }}
        onSubmit={() => {}}
        isPending={false}
      />,
    )

    expect(screen.getByText('编辑里程碑')).toBeInTheDocument()
  })

  it('validates name is required', async () => {
    const user = userEvent.setup()
    renderDialog(
      <CreateEditMilestoneDialog
        open={true}
        onOpenChange={() => {}}
        mode="create"
        onSubmit={() => {}}
        isPending={false}
      />,
    )

    await user.click(screen.getByRole('button', { name: '确认' }))

    await waitFor(() => {
      expect(screen.getByText('请输入里程碑名称')).toBeInTheDocument()
    })
  })

  it('calls onSubmit with form data', async () => {
    const onSubmit = vi.fn()
    const user = userEvent.setup()
    renderDialog(
      <CreateEditMilestoneDialog
        open={true}
        onOpenChange={() => {}}
        mode="create"
        onSubmit={onSubmit}
        isPending={false}
      />,
    )

    await user.type(screen.getByPlaceholderText('请输入里程碑名称'), 'M1 需求分析')
    const dateInput = screen.getByDisplayValue('') // DateInput
    await user.type(dateInput, '2026-06-30')

    await user.click(screen.getByRole('button', { name: '确认' }))

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith({
        milestoneName: 'M1 需求分析',
        expectedEndDate: '2026-06-30',
      })
    })
  })

  it('shows loading state when pending', () => {
    renderDialog(
      <CreateEditMilestoneDialog
        open={true}
        onOpenChange={() => {}}
        mode="create"
        onSubmit={() => {}}
        isPending={true}
      />,
    )

    expect(screen.getByText('保存中...')).toBeInTheDocument()
  })

  it('calls onOpenChange(false) on cancel', async () => {
    const onOpenChange = vi.fn()
    const user = userEvent.setup()
    renderDialog(
      <CreateEditMilestoneDialog
        open={true}
        onOpenChange={onOpenChange}
        mode="create"
        onSubmit={() => {}}
        isPending={false}
      />,
    )

    await user.click(screen.getByRole('button', { name: '取消' }))
    expect(onOpenChange).toHaveBeenCalledWith(false)
  })
})
