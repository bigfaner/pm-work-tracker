import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Avatar, AvatarImage, AvatarFallback } from './avatar'

describe('Avatar', () => {
  it('renders with fallback initials', () => {
    render(
      <Avatar>
        <AvatarFallback>JD</AvatarFallback>
      </Avatar>
    )
    expect(screen.getByText('JD')).toBeInTheDocument()
  })

  it('fallback has primary styling', () => {
    render(
      <Avatar data-testid="avatar">
        <AvatarFallback>AB</AvatarFallback>
      </Avatar>
    )
    expect(screen.getByText('AB').className).toContain('bg-primary-100')
  })

  it('renders with image and fallback', () => {
    render(
      <Avatar>
        <AvatarImage src="https://example.com/avatar.jpg" alt="User" />
        <AvatarFallback>U</AvatarFallback>
      </Avatar>
    )
    // Fallback is shown when image fails to load in jsdom
    expect(screen.getByText('U')).toBeInTheDocument()
  })

  it('avatar is rounded-full', () => {
    const { container } = render(
      <Avatar data-testid="avatar"><AvatarFallback>X</AvatarFallback></Avatar>
    )
    const avatar = container.firstChild as HTMLElement
    expect(avatar.className).toContain('rounded-full')
  })
})
