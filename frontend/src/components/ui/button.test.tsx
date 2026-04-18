import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Button } from './button'

describe('Button', () => {
  it('renders with default primary variant', () => {
    render(<Button>Click me</Button>)
    const btn = screen.getByRole('button', { name: 'Click me' })
    expect(btn).toBeInTheDocument()
    expect(btn.className).toContain('bg-primary-600')
  })

  it('renders secondary variant', () => {
    render(<Button variant="secondary">Secondary</Button>)
    const btn = screen.getByRole('button')
    expect(btn.className).toContain('border-border-dark')
  })

  it('renders warning variant', () => {
    render(<Button variant="warning">Warn</Button>)
    const btn = screen.getByRole('button')
    expect(btn.className).toContain('text-warning')
  })

  it('renders danger variant', () => {
    render(<Button variant="danger">Delete</Button>)
    const btn = screen.getByRole('button')
    expect(btn.className).toContain('text-error')
  })

  it('renders ghost variant', () => {
    render(<Button variant="ghost">Ghost</Button>)
    const btn = screen.getByRole('button')
    expect(btn.className).toContain('bg-transparent')
  })

  it('renders icon variant', () => {
    render(<Button variant="icon">X</Button>)
    const btn = screen.getByRole('button')
    expect(btn.className).toContain('w-9')
  })

  it('renders small size', () => {
    render(<Button size="sm">Small</Button>)
    const btn = screen.getByRole('button')
    expect(btn.className).toContain('h-8')
  })

  it('renders large size', () => {
    render(<Button size="lg">Large</Button>)
    const btn = screen.getByRole('button')
    expect(btn.className).toContain('h-11')
  })

  it('handles disabled state', () => {
    render(<Button disabled>Disabled</Button>)
    const btn = screen.getByRole('button')
    expect(btn).toBeDisabled()
    expect(btn.className).toContain('disabled:opacity-50')
  })

  it('handles click', async () => {
    let clicked = false
    render(<Button onClick={() => { clicked = true }}>Click</Button>)
    await userEvent.click(screen.getByRole('button'))
    expect(clicked).toBe(true)
  })

  it('merges custom className', () => {
    render(<Button className="my-custom">Test</Button>)
    expect(screen.getByRole('button').className).toContain('my-custom')
  })
})
