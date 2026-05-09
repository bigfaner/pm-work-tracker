import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { CheckboxGroup } from './checkbox-group'

const options = [
  { value: 'team:create', label: '创建团队' },
  { value: 'team:read', label: '查看团队信息' },
  { value: 'team:update', label: '编辑团队信息' },
]

describe('CheckboxGroup', () => {
  it('renders all options', () => {
    render(
      <CheckboxGroup
        options={options}
        selected={[]}
        onChange={() => {}}
      />,
    )
    expect(screen.getByText('创建团队')).toBeInTheDocument()
    expect(screen.getByText('查看团队信息')).toBeInTheDocument()
    expect(screen.getByText('编辑团队信息')).toBeInTheDocument()
  })

  it('renders title with selection count', () => {
    render(
      <CheckboxGroup
        options={options}
        selected={['team:read']}
        onChange={() => {}}
        title="团队管理"
      />,
    )
    expect(screen.getByText('团队管理')).toBeInTheDocument()
    expect(screen.getByText('(1/3)')).toBeInTheDocument()
  })

  it('checks selected options', () => {
    render(
      <CheckboxGroup
        options={options}
        selected={['team:read']}
        onChange={() => {}}
      />,
    )
    const checkboxes = screen.getAllByRole('checkbox')
    expect(checkboxes[0]).not.toBeChecked()
    expect(checkboxes[1]).toBeChecked()
    expect(checkboxes[2]).not.toBeChecked()
  })

  it('calls onChange when toggling an option', async () => {
    const user = userEvent.setup()
    const handleChange = vi.fn()
    render(
      <CheckboxGroup
        options={options}
        selected={[]}
        onChange={handleChange}
      />,
    )

    await user.click(screen.getByText('创建团队'))
    expect(handleChange).toHaveBeenCalledWith(['team:create'])
  })

  it('calls onChange to remove when unchecking', async () => {
    const user = userEvent.setup()
    const handleChange = vi.fn()
    render(
      <CheckboxGroup
        options={options}
        selected={['team:read']}
        onChange={handleChange}
      />,
    )

    await user.click(screen.getByText('查看团队信息'))
    expect(handleChange).toHaveBeenCalledWith([])
  })

  it('select all checkbox toggles all options', async () => {
    const user = userEvent.setup()
    const handleChange = vi.fn()

    const { rerender } = render(
      <CheckboxGroup
        options={options}
        selected={[]}
        onChange={handleChange}
        title="团队管理"
      />,
    )

    // Click the title (select all)
    await user.click(screen.getByText('团队管理'))
    expect(handleChange).toHaveBeenCalledWith(['team:create', 'team:read', 'team:update'])
  })

  it('deselect all when all are selected', async () => {
    const user = userEvent.setup()
    const handleChange = vi.fn()
    render(
      <CheckboxGroup
        options={options}
        selected={['team:create', 'team:read', 'team:update']}
        onChange={handleChange}
        title="团队管理"
      />,
    )

    await user.click(screen.getByText('团队管理'))
    expect(handleChange).toHaveBeenCalledWith([])
  })

  it('shows indeterminate state for partial selection', () => {
    render(
      <CheckboxGroup
        options={options}
        selected={['team:read']}
        onChange={() => {}}
        title="团队管理"
      />,
    )

    // The "all" checkbox should exist but not be fully checked
    const checkboxes = screen.getAllByRole('checkbox')
    // First checkbox is the "select all" in title
    expect(checkboxes[0]).not.toBeChecked()
    // But it should have indeterminate state (tested via ref)
    expect((checkboxes[0] as HTMLInputElement).indeterminate).toBe(true)
  })

  it('disables all checkboxes when disabled', () => {
    render(
      <CheckboxGroup
        options={options}
        selected={[]}
        onChange={() => {}}
        disabled
      />,
    )
    const checkboxes = screen.getAllByRole('checkbox')
    checkboxes.forEach((cb) => expect(cb).toBeDisabled())
  })
})
