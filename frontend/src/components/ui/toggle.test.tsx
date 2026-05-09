import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Toggle } from './toggle'

describe('Toggle', () => {
  it('renders toggle button', () => {
    render(<Toggle>Toggle</Toggle>)
    expect(screen.getByRole('button', { name: 'Toggle' })).toBeInTheDocument()
  })

  it('toggles pressed state on click', async () => {
    render(<Toggle>Toggle</Toggle>)
    const btn = screen.getByRole('button')
    expect(btn.getAttribute('data-state')).toBe('off')
    await userEvent.click(btn)
    expect(btn.getAttribute('data-state')).toBe('on')
  })

  it('renders outline variant', () => {
    render(<Toggle variant="outline">Outline</Toggle>)
    expect(screen.getByRole('button').className).toContain('border')
  })

  it('renders small size', () => {
    render(<Toggle size="sm">Small</Toggle>)
    expect(screen.getByRole('button').className).toContain('h-8')
  })
})
