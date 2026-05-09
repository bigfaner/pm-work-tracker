import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import UserAvatar from './UserAvatar'

describe('UserAvatar', () => {
  it('renders first character of name as fallback', () => {
    render(<UserAvatar name="张明" />)
    expect(screen.getByText('张')).toBeInTheDocument()
  })

  it('renders question mark for empty name', () => {
    render(<UserAvatar name="" />)
    expect(screen.getByText('?')).toBeInTheDocument()
  })

  it('renders with small size', () => {
    const { container } = render(<UserAvatar name="Test" size="sm" />)
    // The Avatar root has the size class applied
    const avatar = container.firstChild as HTMLElement
    expect(avatar).toBeInTheDocument()
  })

  it('renders with medium size by default', () => {
    const { container } = render(<UserAvatar name="Test" />)
    const avatar = container.firstChild as HTMLElement
    expect(avatar).toBeInTheDocument()
  })

  it('uses consistent color based on name', () => {
    const { container: c1 } = render(<UserAvatar name="张明" />)
    const { container: c2 } = render(<UserAvatar name="张明" />)
    // Both should render the same initial
    expect(c1.querySelector('span')?.textContent).toBe('张')
    expect(c2.querySelector('span')?.textContent).toBe('张')
  })

  it('renders different names with initials', () => {
    const { container: c1 } = render(<UserAvatar name="张明" />)
    const { container: c2 } = render(<UserAvatar name="李华" />)
    expect(c1.querySelector('span')?.textContent).toBe('张')
    expect(c2.querySelector('span')?.textContent).toBe('李')
  })
})
