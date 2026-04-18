import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from './select'

function SelectWrapper({ defaultValue = '' }: { defaultValue?: string }) {
  return (
    <Select defaultValue={defaultValue}>
      <SelectTrigger data-testid="trigger">
        <SelectValue placeholder="Select..." />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="opt1">Option 1</SelectItem>
        <SelectItem value="opt2">Option 2</SelectItem>
      </SelectContent>
    </Select>
  )
}

describe('Select', () => {
  it('renders trigger with placeholder', () => {
    render(<SelectWrapper />)
    expect(screen.getByText('Select...')).toBeInTheDocument()
  })

  it('renders with default value', () => {
    render(<SelectWrapper defaultValue="opt1" />)
    expect(screen.getByText('Option 1')).toBeInTheDocument()
  })

  it('trigger has proper styling', () => {
    render(<SelectWrapper />)
    const trigger = screen.getByTestId('trigger')
    expect(trigger.className).toContain('rounded-md')
  })

  it('displays selected default value in trigger', () => {
    render(<SelectWrapper defaultValue="opt1" />)
    expect(screen.getByText('Option 1')).toBeInTheDocument()
  })
})
