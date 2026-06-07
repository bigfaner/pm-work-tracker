import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import CreateMilestoneDialog from './CreateMilestoneDialog'
import type { MilestoneFormState } from './CreateMilestoneDialog'

function createDefaultForm(
  overrides: Partial<MilestoneFormState> = {},
): MilestoneFormState {
  return {
    milestoneName: '',
    expectedEndDate: '',
    milestoneDesc: '',
    ...overrides,
  }
}

function renderDialog(
  overrides: Partial<Parameters<typeof CreateMilestoneDialog>[0]> = {},
) {
  const onFormChange = vi.fn()
  const onOpenChange = vi.fn()
  const onSubmit = vi.fn()
  const form = createDefaultForm(overrides.form)
  const result = render(
    <CreateMilestoneDialog
      open={true}
      onOpenChange={onOpenChange}
      form={form}
      onFormChange={onFormChange}
      onSubmit={onSubmit}
      isPending={false}
      {...overrides}
    />,
  )
  return { onFormChange, onOpenChange, onSubmit, ...result }
}

describe('CreateMilestoneDialog', () => {
  // AC-1: Dialog renders form with name, expected end date, description
  it('renders name input with placeholder', () => {
    renderDialog()
    expect(
      screen.getByPlaceholderText('请输入里程碑名称'),
    ).toBeInTheDocument()
  })

  it('renders name input with maxLength 100', () => {
    renderDialog()
    const input = screen.getByPlaceholderText('请输入里程碑名称')
    expect(input).toHaveAttribute('maxlength', '100')
  })

  it('renders expected end date picker', () => {
    renderDialog()
    expect(screen.getByText('计划完成时间')).toBeInTheDocument()
  })

  it('renders description textarea', () => {
    renderDialog()
    expect(
      screen.getByPlaceholderText('请输入描述（可选）'),
    ).toBeInTheDocument()
  })

  it('renders required markers on name and expected end date', () => {
    renderDialog()
    const markers = screen.getAllByText('*')
    // Name and expectedEndDate each have one required marker
    expect(markers.length).toBeGreaterThanOrEqual(2)
  })

  // AC-2: Create mode
  it('shows create mode title "创建里程碑"', () => {
    renderDialog()
    expect(screen.getByText('创建里程碑')).toBeInTheDocument()
  })

  it('disables confirm when name is empty', () => {
    renderDialog({
      form: createDefaultForm({
        milestoneName: '',
        expectedEndDate: '2026-12-31',
      }),
    })
    expect(screen.getByText('确认')).toBeDisabled()
  })

  it('disables confirm when expectedEndDate is empty', () => {
    renderDialog({
      form: createDefaultForm({
        milestoneName: 'MVP 发布',
        expectedEndDate: '',
      }),
    })
    expect(screen.getByText('确认')).toBeDisabled()
  })

  it('disables confirm when name is whitespace only', () => {
    renderDialog({
      form: createDefaultForm({
        milestoneName: '   ',
        expectedEndDate: '2026-12-31',
      }),
    })
    expect(screen.getByText('确认')).toBeDisabled()
  })

  it('enables confirm when name and expectedEndDate are filled', () => {
    renderDialog({
      form: createDefaultForm({
        milestoneName: 'MVP 发布',
        expectedEndDate: '2026-12-31',
      }),
    })
    expect(screen.getByText('确认')).not.toBeDisabled()
  })

  // AC-3: Edit mode
  it('shows edit mode title "编辑里程碑" when milestone is provided', () => {
    renderDialog({
      milestone: {
        bizKey: 'ms-1',
        teamKey: 'team-1',
        milestoneMapKey: 'map-1',
        milestoneName: 'MVP 发布',
        milestoneDesc: '描述内容',
        expectedEndDate: '2026-06-30',
        milestoneStatus: 'not_started',
        statusName: '未开始',
        completion: 0,
        relatedMICount: 0,
        createTime: '2026-01-01',
        dbUpdateTime: '2026-01-01',
      },
    })
    expect(screen.getByText('编辑里程碑')).toBeInTheDocument()
  })

  it('pre-fills form with milestone values in edit mode', () => {
    renderDialog({
      form: createDefaultForm({
        milestoneName: 'MVP 发布',
        expectedEndDate: '2026-06-30',
        milestoneDesc: '描述内容',
      }),
      milestone: {
        bizKey: 'ms-1',
        teamKey: 'team-1',
        milestoneMapKey: 'map-1',
        milestoneName: 'MVP 发布',
        milestoneDesc: '描述内容',
        expectedEndDate: '2026-06-30',
        milestoneStatus: 'not_started',
        statusName: '未开始',
        completion: 0,
        relatedMICount: 0,
        createTime: '2026-01-01',
        dbUpdateTime: '2026-01-01',
      },
    })
    expect(
      (screen.getByPlaceholderText('请输入里程碑名称') as HTMLInputElement)
        .value,
    ).toBe('MVP 发布')
    expect(
      (screen.getByPlaceholderText('请输入描述（可选）') as HTMLTextAreaElement)
        .value,
    ).toBe('描述内容')
  })

  // AC-4: Submitting state
  it('disables confirm when isPending is true', () => {
    renderDialog({
      form: createDefaultForm({
        milestoneName: 'Test',
        expectedEndDate: '2026-12-31',
      }),
      isPending: true,
    })
    expect(screen.getByText('确认')).toBeDisabled()
  })

  it('disables name input when isPending', () => {
    renderDialog({
      form: createDefaultForm({
        milestoneName: 'Test',
        expectedEndDate: '2026-12-31',
      }),
      isPending: true,
    })
    expect(screen.getByPlaceholderText('请输入里程碑名称')).toBeDisabled()
  })

  it('disables date input when isPending', () => {
    renderDialog({
      form: createDefaultForm({
        milestoneName: 'Test',
        expectedEndDate: '2026-12-31',
      }),
      isPending: true,
    })
    // DateInput renders <input type="date"> inside a wrapper; the inner input has the value
    expect(screen.getByDisplayValue('2026-12-31')).toBeDisabled()
  })

  it('disables description textarea when isPending', () => {
    renderDialog({
      form: createDefaultForm({
        milestoneName: 'Test',
        expectedEndDate: '2026-12-31',
        milestoneDesc: 'desc',
      }),
      isPending: true,
    })
    expect(screen.getByPlaceholderText('请输入描述（可选）')).toBeDisabled()
  })

  it('disables cancel button when isPending', () => {
    renderDialog({
      form: createDefaultForm({
        milestoneName: 'Test',
        expectedEndDate: '2026-12-31',
      }),
      isPending: true,
    })
    expect(screen.getByText('取消')).toBeDisabled()
  })

  // Form interactions
  it('calls onFormChange when name input changes', async () => {
    const user = userEvent.setup()
    const { onFormChange } = renderDialog()
    const input = screen.getByPlaceholderText('请输入里程碑名称')
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

  it('calls onFormChange when date changes', async () => {
    const user = userEvent.setup()
    const { onFormChange } = renderDialog()
    const dateInput = screen.getByLabelText('计划完成时间')
    await user.type(dateInput, '2026-12-31')
    expect(onFormChange).toHaveBeenCalled()
  })

  it('calls onSubmit when confirm button is clicked', async () => {
    const user = userEvent.setup()
    const { onSubmit } = renderDialog({
      form: createDefaultForm({
        milestoneName: 'MVP 发布',
        expectedEndDate: '2026-12-31',
      }),
    })
    await user.click(screen.getByText('确认'))
    expect(onSubmit).toHaveBeenCalled()
  })

  it('calls onOpenChange(false) when cancel is clicked', async () => {
    const user = userEvent.setup()
    const { onOpenChange } = renderDialog({
      form: createDefaultForm({
        milestoneName: 'MVP 发布',
        expectedEndDate: '2026-12-31',
      }),
    })
    await user.click(screen.getByText('取消'))
    expect(onOpenChange).toHaveBeenCalledWith(false)
  })
})
