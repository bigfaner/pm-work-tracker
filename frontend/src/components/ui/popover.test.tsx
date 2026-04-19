import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Popover, PopoverTrigger, PopoverContent } from './popover'
import { Button } from './button'

describe('Popover', () => {
  it('renders trigger', () => {
    render(
      <Popover>
        <PopoverTrigger asChild>
          <Button>Open Popover</Button>
        </PopoverTrigger>
        <PopoverContent>Popover content here</PopoverContent>
      </Popover>
    )
    expect(screen.getByRole('button', { name: 'Open Popover' })).toBeInTheDocument()
  })

  it('shows content on click', async () => {
    render(
      <Popover>
        <PopoverTrigger asChild>
          <Button>Open</Button>
        </PopoverTrigger>
        <PopoverContent>Popover content</PopoverContent>
      </Popover>
    )
    await userEvent.click(screen.getByRole('button'))
    expect(screen.getByText('Popover content')).toBeInTheDocument()
  })
})
