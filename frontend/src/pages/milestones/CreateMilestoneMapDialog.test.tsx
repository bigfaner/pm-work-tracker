import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import CreateMilestoneMapDialog from './CreateMilestoneMapDialog'
import type { MilestoneMapFormState } from './CreateMilestoneMapDialog'

const members = [
  { userKey: 'user-1', displayName: '张三' },
  { userKey: 'user-2', displayName: '李四' },
]

function createDefaultForm(
  overrides: Partial<MilestoneMapFormState> = {},
): MilestoneMapFormState {
  return {
    mapName: '',
    assigneeKey: '',
    planStartDate: '',
    expectedEndDate: '',
    mapDesc: '',
    ...overrides,
  }
}

function renderDialog(overrides: Partial<Parameters<typeof CreateMilestoneMapDialog>[0]> = {}) {
  const onFormChange = vi.fn()
  const onOpenChange = vi.fn()
  const onSubmit = vi.fn()
  const form = createDefaultForm(overrides.form)
  const result = render(
    <CreateMilestoneMapDialog
      open={true}
      onOpenChange={onOpenChange}
      form={form}
      onFormChange={onFormChange}
      members={members}
      onSubmit={onSubmit}
      isPending={false}
      {...overrides}
    />,
  )
  return { onFormChange, onOpenChange, onSubmit, ...result }
}

describe('CreateMilestoneMapDialog', () => {
  // AC-1: Dialog renders form with all required fields
  it('renders name input', () => {
    renderDialog()
    expect(screen.getByPlaceholderText('请输入里程碑图名称')).toBeInTheDocument()
  })

  it('renders assignee member select', () => {
    renderDialog()
    expect(screen.getByText('选择负责人')).toBeInTheDocument()
  })

  it('renders plan start date picker', () => {
    renderDialog()
    expect(screen.getByText('计划开始时间')).toBeInTheDocument()
  })

  it('renders expected end date picker', () => {
    renderDialog()
    expect(screen.getByText('计划完成时间')).toBeInTheDocument()
  })

  it('renders description textarea', () => {
    renderDialog()
    expect(screen.getByPlaceholderText('请输入描述（可选）')).toBeInTheDocument()
  })

  // AC-2: Create mode
  it('shows create mode title "创建里程碑图"', () => {
    renderDialog()
    expect(screen.getByText('创建里程碑图')).toBeInTheDocument()
  })

  it('disables confirm when name is empty', () => {
    renderDialog({ form: createDefaultForm({ mapName: '', assigneeKey: 'user-1' }) })
    expect(screen.getByText('确认')).toBeDisabled()
  })

  it('disables confirm when assignee is empty', () => {
    renderDialog({ form: createDefaultForm({ mapName: 'Test', assigneeKey: '' }) })
    expect(screen.getByText('确认')).toBeDisabled()
  })

  it('enables confirm when name and assignee are filled', () => {
    renderDialog({
      form: createDefaultForm({ mapName: '产品MVP', assigneeKey: 'user-1' }),
    })
    expect(screen.getByText('确认')).not.toBeDisabled()
  })

  // AC-3: Edit mode
  it('shows edit mode title "编辑里程碑图" when milestoneMap is provided', () => {
    renderDialog({
      milestoneMap: {
        bizKey: 'map-1',
        teamKey: 'team-1',
        creatorKey: 'user-1',
        creatorName: '张三',
        assigneeKey: 'user-1',
        assigneeName: '张三',
        mapName: '产品MVP',
        mapDesc: '描述内容',
        mapStatus: 'planning',
        statusName: '规划中',
        planStartDate: '2026-05-01',
        expectedEndDate: '2026-12-31',
        milestoneCount: 0,
        itemCount: 0,
        overallProgress: 0,
        createTime: '2026-01-01',
        dbUpdateTime: '2026-01-01',
      },
    })
    expect(screen.getByText('编辑里程碑图')).toBeInTheDocument()
  })

  // AC-4: Submitting state
  it('shows loading and disables confirm when isPending is true', () => {
    renderDialog({
      form: createDefaultForm({ mapName: 'Test', assigneeKey: 'user-1' }),
      isPending: true,
    })
    const confirmButton = screen.getByText('确认')
    expect(confirmButton).toBeDisabled()
  })

  it('disables all inputs when isPending is true', () => {
    renderDialog({
      form: createDefaultForm({
        mapName: 'Test',
        assigneeKey: 'user-1',
        planStartDate: '2026-05-01',
        expectedEndDate: '2026-12-31',
        mapDesc: 'desc',
      }),
      isPending: true,
    })
    // Name input should be disabled
    expect(screen.getByPlaceholderText('请输入里程碑图名称')).toBeDisabled()
    // Description textarea should be disabled
    expect(screen.getByPlaceholderText('请输入描述（可选）')).toBeDisabled()
    // Date inputs should be disabled — they are <input type="date">, query by display value
    expect(screen.getByDisplayValue('2026-05-01')).toBeDisabled()
    expect(screen.getByDisplayValue('2026-12-31')).toBeDisabled()
  })

  // AC-5: Date validation
  it('shows error when expected end date is before plan start date', () => {
    renderDialog({
      form: createDefaultForm({
        mapName: 'Test',
        assigneeKey: 'user-1',
        planStartDate: '2026-06-01',
        expectedEndDate: '2026-05-01',
      }),
    })
    expect(screen.getByText('计划完成时间不得早于计划开始时间')).toBeInTheDocument()
    expect(screen.getByText('确认')).toBeDisabled()
  })

  it('does not show error when dates are valid', () => {
    renderDialog({
      form: createDefaultForm({
        mapName: 'Test',
        assigneeKey: 'user-1',
        planStartDate: '2026-05-01',
        expectedEndDate: '2026-12-31',
      }),
    })
    expect(
      screen.queryByText('计划完成时间不得早于计划开始时间'),
    ).not.toBeInTheDocument()
  })

  it('does not show error when plan start date is empty', () => {
    renderDialog({
      form: createDefaultForm({
        mapName: 'Test',
        assigneeKey: 'user-1',
        expectedEndDate: '2026-12-31',
      }),
    })
    expect(
      screen.queryByText('计划完成时间不得早于计划开始时间'),
    ).not.toBeInTheDocument()
  })

  it('does not show error when expected end date is empty', () => {
    renderDialog({
      form: createDefaultForm({
        mapName: 'Test',
        assigneeKey: 'user-1',
        planStartDate: '2026-05-01',
      }),
    })
    expect(
      screen.queryByText('计划完成时间不得早于计划开始时间'),
    ).not.toBeInTheDocument()
  })

  it('allows submit when dates are equal', () => {
    renderDialog({
      form: createDefaultForm({
        mapName: 'Test',
        assigneeKey: 'user-1',
        planStartDate: '2026-06-01',
        expectedEndDate: '2026-06-01',
      }),
    })
    expect(screen.getByText('确认')).not.toBeDisabled()
  })

  // Form interactions
  it('calls onFormChange when name input changes', async () => {
    const user = userEvent.setup()
    const { onFormChange } = renderDialog()
    const input = screen.getByPlaceholderText('请输入里程碑图名称')
    await user.type(input, 'A')
    expect(onFormChange).toHaveBeenCalled()
  })

  it('calls onFormChange when description changes', async () => {
    const user = userEvent.setup()
    const { onFormChange } = renderDialog()
    const textarea = screen.getByPlaceholderText('请输入描述（可选）')
    await user.type(textarea, 'A')
    expect(onFormChange).toHaveBeenCalled()
  })
})
