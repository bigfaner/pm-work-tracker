import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Textarea } from './textarea'

describe('Textarea', () => {
  it('renders with default styles', () => {
    render(<Textarea placeholder="Enter text" />)
    const textarea = screen.getByPlaceholderText('Enter text')
    expect(textarea).toBeInTheDocument()
    expect(textarea.className).toContain('border-border-dark')
  })

  it('handles disabled state', () => {
    render(<Textarea disabled placeholder="Disabled" />)
    const textarea = screen.getByPlaceholderText('Disabled')
    expect(textarea).toBeDisabled()
    expect(textarea.className).toContain('disabled:opacity-50')
  })

  it('accepts custom className', () => {
    render(<Textarea className="custom-class" />)
    expect(screen.getByRole('textbox').className).toContain('custom-class')
  })

  it('handles typing', async () => {
    render(<Textarea placeholder="Type here" />)
    const textarea = screen.getByPlaceholderText('Type here')
    await userEvent.type(textarea, 'hello world')
    expect(textarea).toHaveValue('hello world')
  })

  it('includes min-h and resize-y base classes', () => {
    render(<Textarea placeholder="Test" />)
    const textarea = screen.getByPlaceholderText('Test')
    expect(textarea.className).toContain('min-h-[120px]')
    expect(textarea.className).toContain('resize-y')
  })

  it('supports rows attribute', () => {
    render(<Textarea rows={5} placeholder="Rows" />)
    const textarea = screen.getByPlaceholderText('Rows')
    expect(textarea).toHaveAttribute('rows', '5')
  })

  it('renders as textarea element', () => {
    render(<Textarea placeholder="Test" />)
    const textarea = screen.getByPlaceholderText('Test')
    expect(textarea.tagName).toBe('TEXTAREA')
  })

  it('forwards ref', () => {
    const ref = { current: null as HTMLTextAreaElement | null }
    render(<Textarea ref={ref} placeholder="Ref" />)
    expect(ref.current).toBeInstanceOf(HTMLTextAreaElement)
  })
})
