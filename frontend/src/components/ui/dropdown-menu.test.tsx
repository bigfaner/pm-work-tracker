import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuLabel,
} from './dropdown-menu'
import { Button } from './button'

function DropdownWrapper() {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button>Menu</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        <DropdownMenuLabel>Actions</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem>Edit</DropdownMenuItem>
        <DropdownMenuItem danger>Delete</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

describe('DropdownMenu', () => {
  it('renders trigger', () => {
    render(<DropdownWrapper />)
    expect(screen.getByRole('button', { name: 'Menu' })).toBeInTheDocument()
  })

  it('opens on click and shows items', async () => {
    render(<DropdownWrapper />)
    await userEvent.click(screen.getByRole('button', { name: 'Menu' }))
    expect(screen.getByText('Edit')).toBeInTheDocument()
    expect(screen.getByText('Delete')).toBeInTheDocument()
    expect(screen.getByText('Actions')).toBeInTheDocument()
  })

  it('danger item has error styling', async () => {
    render(<DropdownWrapper />)
    await userEvent.click(screen.getByRole('button'))
    expect(screen.getByText('Delete').className).toContain('text-error')
  })
})
