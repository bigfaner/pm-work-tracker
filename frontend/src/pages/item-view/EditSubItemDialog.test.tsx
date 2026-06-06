import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import EditSubItemDialog from './EditSubItemDialog'
import type { EditSubItemFormState } from './EditSubItemDialog'

const defaultForm: EditSubItemFormState = {
  title: 'Test Sub Item',
  priority: 'P2',
  assigneeKey: '',
  startDate: '2026-01-01',
  expectedEndDate: '2026-03-01',
  description: 'A test sub item',
}

function renderDialog(overrides: Partial<Parameters<typeof EditSubItemDialog>[0]> = {}) {
  const onFormChange = vi.fn()
  const onOpenChange = vi.fn()
  const onSubmit = vi.fn()
  const result = render(
    <EditSubItemDialog
      open={true}
      onOpenChange={onOpenChange}
      form={defaultForm}
      onFormChange={onFormChange}
      members={[]}
      onSubmit={onSubmit}
      isPending={false}
      {...overrides}
    />,
  )
  return { onFormChange, onOpenChange, onSubmit, ...result }
}

describe('EditSubItemDialog (item-view)', () => {
  it('renders startDate field with current value', () => {
    renderDialog()
    const labels = screen.getAllByText('开始时间')
    expect(labels.length).toBeGreaterThanOrEqual(1)

    // Find the date input associated with startDate by checking all date inputs
    const dateInputs = screen.getAllByDisplayValue('2026-01-01')
    expect(dateInputs.length).toBeGreaterThanOrEqual(1)
  })

  it('calls onFormChange when startDate is changed', async () => {
    const user = userEvent.setup()
    const { onFormChange } = renderDialog()

    // Find the startDate input - it's the one with value '2026-01-01'
    const startDateInput = screen.getAllByDisplayValue('2026-01-01')[0]
    await user.clear(startDateInput)
    await user.type(startDateInput, '2026-02-01')

    expect(onFormChange).toHaveBeenCalled()
  })

  it('disables save when startDate is after expectedEndDate', () => {
    renderDialog({
      form: {
        ...defaultForm,
        startDate: '2026-06-01',
        expectedEndDate: '2026-03-01',
      },
    })

    const saveButton = screen.getByText('保存')
    expect(saveButton).toBeDisabled()
    expect(screen.getByText('开始时间不得晚于结束时间')).toBeInTheDocument()
  })

  it('allows save when startDate is before expectedEndDate', () => {
    renderDialog({
      form: {
        ...defaultForm,
        startDate: '2026-01-01',
        expectedEndDate: '2026-03-01',
      },
    })

    const saveButton = screen.getByText('保存')
    expect(saveButton).not.toBeDisabled()
  })

  it('allows save when startDate is empty', () => {
    renderDialog({
      form: {
        ...defaultForm,
        startDate: '',
        expectedEndDate: '2026-03-01',
      },
    })

    const saveButton = screen.getByText('保存')
    expect(saveButton).not.toBeDisabled()
  })

  it('allows save when expectedEndDate is empty', () => {
    renderDialog({
      form: {
        ...defaultForm,
        startDate: '2026-01-01',
        expectedEndDate: '',
      },
    })

    const saveButton = screen.getByText('保存')
    expect(saveButton).not.toBeDisabled()
  })

  it('allows save when startDate equals expectedEndDate', () => {
    renderDialog({
      form: {
        ...defaultForm,
        startDate: '2026-03-01',
        expectedEndDate: '2026-03-01',
      },
    })

    const saveButton = screen.getByText('保存')
    expect(saveButton).not.toBeDisabled()
  })

  it('does not show date error message when dates are valid', () => {
    renderDialog()
    expect(screen.queryByText('开始时间不得晚于结束时间')).not.toBeInTheDocument()
  })

  it('calls onFormChange when expectedEndDate is changed', async () => {
    const user = userEvent.setup()
    const { onFormChange } = renderDialog()

    const endDateInput = screen.getAllByDisplayValue('2026-03-01')[0]
    await user.clear(endDateInput)
    await user.type(endDateInput, '2026-04-01')

    expect(onFormChange).toHaveBeenCalled()
  })
})
