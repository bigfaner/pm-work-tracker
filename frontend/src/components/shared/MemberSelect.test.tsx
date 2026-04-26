import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemberSelect } from './MemberSelect'

const members = [
  { bizKey: 'alice-key', displayName: 'Alice' },
  { bizKey: 'bob-key', displayName: 'Bob' },
  { bizKey: 'charlie-key', displayName: 'Charlie' },
]

describe('MemberSelect', () => {
  it('renders with placeholder text when no value and allowEmpty is false', () => {
    render(
      <MemberSelect
        members={members}
        selectedId=""
        onSelect={vi.fn()}
        placeholder="选择负责人"
        allowEmpty={false}
      />,
    )
    const trigger = screen.getByRole('combobox')
    // Radix SelectValue placeholder is rendered as span content
    expect(trigger).toHaveTextContent('选择负责人')
  })

  it('renders "不指定" when selectedId is empty and allowEmpty is true', () => {
    render(
      <MemberSelect
        members={members}
        selectedId=""
        onSelect={vi.fn()}
      />,
    )
    expect(screen.getByText('不指定')).toBeInTheDocument()
  })

  it('renders member options when opened', async () => {
    render(
      <MemberSelect
        members={members}
        selectedId=""
        onSelect={vi.fn()}
      />,
    )
    const trigger = screen.getByRole('combobox')
    await userEvent.click(trigger)
    expect(screen.getByText('Alice')).toBeInTheDocument()
    expect(screen.getByText('Bob')).toBeInTheDocument()
    expect(screen.getByText('Charlie')).toBeInTheDocument()
  })

  it('calls onSelect when a member is selected', async () => {
    const onSelect = vi.fn()
    render(
      <MemberSelect
        members={members}
        selectedId=""
        onSelect={onSelect}
      />,
    )
    const trigger = screen.getByRole('combobox')
    await userEvent.click(trigger)
    await userEvent.click(screen.getByText('Bob'))
    expect(onSelect).toHaveBeenCalledWith('bob-key')
  })

  it('shows "不指定" option when allowEmpty is true', async () => {
    render(
      <MemberSelect
        members={members}
        selectedId="alice-key"
        onSelect={vi.fn()}
        allowEmpty
      />,
    )
    const trigger = screen.getByRole('combobox')
    await userEvent.click(trigger)
    expect(screen.getByText('不指定')).toBeInTheDocument()
  })

  it('does not show "不指定" option when allowEmpty is false', async () => {
    render(
      <MemberSelect
        members={members}
        selectedId=""
        onSelect={vi.fn()}
        allowEmpty={false}
      />,
    )
    const trigger = screen.getByRole('combobox')
    await userEvent.click(trigger)
    expect(screen.queryByText('不指定')).not.toBeInTheDocument()
  })

  it('calls onSelect with empty string when "不指定" is selected', async () => {
    const onSelect = vi.fn()
    render(
      <MemberSelect
        members={members}
        selectedId="alice-key"
        onSelect={onSelect}
        allowEmpty
      />,
    )
    const trigger = screen.getByRole('combobox')
    await userEvent.click(trigger)
    await userEvent.click(screen.getByText('不指定'))
    expect(onSelect).toHaveBeenCalledWith('')
  })

  it('displays selected member name when selectedId is set', () => {
    render(
      <MemberSelect
        members={members}
        selectedId="bob-key"
        onSelect={vi.fn()}
      />,
    )
    expect(screen.getByText('Bob')).toBeInTheDocument()
  })
})
