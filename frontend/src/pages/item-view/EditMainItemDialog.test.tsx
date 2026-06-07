import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import EditMainItemDialog from './EditMainItemDialog'
import type { EditMainItemFormState } from './EditMainItemDialog'

const defaultForm: EditMainItemFormState = {
  title: 'Test Item',
  priority: 'P2',
  assigneeKey: 'user1',
  startDate: '2026-01-01',
  expectedEndDate: '2026-03-01',
  description: 'A test item',
  milestoneKey: '',
}

const milestones = [
  { bizKey: 'ms-1', teamKey: 't1', milestoneMapKey: 'map-1', milestoneName: 'MVP Release', milestoneDesc: '', expectedEndDate: null, milestoneStatus: 'in_progress', statusName: '进行中', completion: 50, relatedMICount: 3, createTime: '', dbUpdateTime: '' },
  { bizKey: 'ms-2', teamKey: 't1', milestoneMapKey: 'map-1', milestoneName: 'Phase 2', milestoneDesc: '', expectedEndDate: null, milestoneStatus: 'not_started', statusName: '未开始', completion: 0, relatedMICount: 0, createTime: '', dbUpdateTime: '' },
]

function renderDialog(overrides: Partial<Parameters<typeof EditMainItemDialog>[0]> = {}) {
  const onFormChange = vi.fn()
  const onOpenChange = vi.fn()
  const onSubmit = vi.fn()
  const result = render(
    <EditMainItemDialog
      open={true}
      onOpenChange={onOpenChange}
      form={defaultForm}
      onFormChange={onFormChange}
      members={[{ userKey: 'user1', displayName: 'User One' }]}
      milestones={[...milestones]}
      milestonesError={false}
      onSubmit={onSubmit}
      isPending={false}
      {...overrides}
    />,
  )
  return { onFormChange, onOpenChange, onSubmit, ...result }
}

describe('EditMainItemDialog (item-view) — milestone selector', () => {
  it('renders 所属里程碑 label and select', () => {
    renderDialog()
    expect(screen.getByText('所属里程碑')).toBeInTheDocument()
  })

  it('shows 未分配 as default when no milestone selected', () => {
    renderDialog({ form: { ...defaultForm, milestoneKey: '' } })
    expect(screen.getByText('未分配')).toBeInTheDocument()
  })

  it('shows milestone options including 未分配 and milestone names', async () => {
    const user = userEvent.setup()
    renderDialog()
    // Find the milestone combobox: it's preceded by the 所属里程碑 label
    const milestoneLabel = screen.getByText('所属里程碑')
    const milestoneWrapper = milestoneLabel.parentElement!
    const trigger = milestoneWrapper.querySelector('[role="combobox"]') as HTMLElement
    await user.click(trigger)
    expect(screen.getByRole('option', { name: '未分配' })).toBeInTheDocument()
    expect(screen.getByRole('option', { name: 'MVP Release' })).toBeInTheDocument()
    expect(screen.getByRole('option', { name: 'Phase 2' })).toBeInTheDocument()
  })

  it('pre-fills current milestone value when dialog opens', () => {
    renderDialog({ form: { ...defaultForm, milestoneKey: 'ms-1' } })
    expect(screen.getByText('MVP Release')).toBeInTheDocument()
  })

  it('shows 未分配 when milestoneKey is null/empty', () => {
    renderDialog({ form: { ...defaultForm, milestoneKey: '' } })
    expect(screen.getByText('未分配')).toBeInTheDocument()
  })

  it('calls onFormChange with empty string when 未分配 is selected', async () => {
    const user = userEvent.setup()
    const { onFormChange } = renderDialog({ form: { ...defaultForm, milestoneKey: 'ms-1' } })
    const milestoneLabel = screen.getByText('所属里程碑')
    const milestoneWrapper = milestoneLabel.parentElement!
    const trigger = milestoneWrapper.querySelector('[role="combobox"]') as HTMLElement
    await user.click(trigger)
    const unassigned = screen.getByRole('option', { name: '未分配' })
    await user.click(unassigned)
    expect(onFormChange).toHaveBeenCalled()
    const updater = onFormChange.mock.calls[0][0]
    const result = updater(defaultForm)
    expect(result.milestoneKey).toBe('')
  })

  it('calls onFormChange with milestone bizKey when a milestone is selected', async () => {
    const user = userEvent.setup()
    const { onFormChange } = renderDialog()
    const milestoneLabel = screen.getByText('所属里程碑')
    const milestoneWrapper = milestoneLabel.parentElement!
    const trigger = milestoneWrapper.querySelector('[role="combobox"]') as HTMLElement
    await user.click(trigger)
    const option = screen.getByRole('option', { name: 'MVP Release' })
    await user.click(option)
    expect(onFormChange).toHaveBeenCalled()
    const updater = onFormChange.mock.calls[0][0]
    const result = updater(defaultForm)
    expect(result.milestoneKey).toBe('ms-1')
  })

  it('shows only 未分配 option when no milestones exist', async () => {
    const user = userEvent.setup()
    renderDialog({ milestones: [] })
    const milestoneLabel = screen.getByText('所属里程碑')
    const milestoneWrapper = milestoneLabel.parentElement!
    const trigger = milestoneWrapper.querySelector('[role="combobox"]') as HTMLElement
    await user.click(trigger)
    expect(screen.getByRole('option', { name: '未分配' })).toBeInTheDocument()
    expect(screen.queryByRole('option', { name: 'MVP Release' })).not.toBeInTheDocument()
  })

  it('shows 加载失败 and disables select when API fails', () => {
    renderDialog({ milestonesError: true })
    expect(screen.getByText('加载失败')).toBeInTheDocument()
    const milestoneLabel = screen.getByText('所属里程碑')
    const milestoneWrapper = milestoneLabel.parentElement!
    const trigger = milestoneWrapper.querySelector('[role="combobox"]') as HTMLElement
    expect(trigger).toBeDisabled()
  })

  it('positions milestone selector below 负责人 field', () => {
    renderDialog()
    const labels = screen.getAllByText(/负责人|所属里程碑/)
    // 负责人 should appear before 所属里程碑 in DOM order
    const assigneeIdx = labels.findIndex((el) => el.textContent === '负责人')
    const milestoneIdx = labels.findIndex((el) => el.textContent === '所属里程碑')
    expect(assigneeIdx).toBeLessThan(milestoneIdx)
  })
})
