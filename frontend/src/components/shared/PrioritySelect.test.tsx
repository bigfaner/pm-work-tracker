import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { PrioritySelectItems } from './PrioritySelect'
import { Select, SelectContent, SelectTrigger } from '@/components/ui/select'

function renderInSelect() {
  return render(
    <Select defaultOpen>
      <SelectTrigger data-testid="trigger" />
      <SelectContent>
        <PrioritySelectItems />
      </SelectContent>
    </Select>
  )
}

describe('PrioritySelectItems', () => {
  it('renders exactly 3 priority options', () => {
    renderInSelect()
    expect(screen.getByText('P1')).toBeInTheDocument()
    expect(screen.getByText('P2')).toBeInTheDocument()
    expect(screen.getByText('P3')).toBeInTheDocument()
  })
})
