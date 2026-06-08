import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { DateInput } from './date-input'

describe('DateInput', () => {
  it('renders date input with type="date"', () => {
    render(<DateInput />)
    const input = document.querySelector('input[type="date"]')
    expect(input).toHaveAttribute('type', 'date')
  })

  it('bug: clicking the input should call showPicker to open calendar', async () => {
    const showPickerMock = vi.fn()
    HTMLInputElement.prototype.showPicker = showPickerMock

    render(<DateInput />)
    const input = document.querySelector('input[type="date"]')!
    await userEvent.click(input)

    expect(showPickerMock).toHaveBeenCalled()

    HTMLInputElement.prototype.showPicker = undefined as unknown as () => void
  })

  it('does not call showPicker when disabled', async () => {
    const showPickerMock = vi.fn()
    HTMLInputElement.prototype.showPicker = showPickerMock

    render(<DateInput disabled />)
    const input = document.querySelector('input[type="date"]')!
    expect(input).toBeDisabled()

    HTMLInputElement.prototype.showPicker = undefined as unknown as () => void
  })
})
