import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from './table'

describe('Table', () => {
  it('renders table with header and body', () => {
    render(
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Status</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow>
            <TableCell>Item A</TableCell>
            <TableCell>Active</TableCell>
          </TableRow>
        </TableBody>
      </Table>
    )
    expect(screen.getByText('Name')).toBeInTheDocument()
    expect(screen.getByText('Item A')).toBeInTheDocument()
  })

  it('TableHead has uppercase text styling', () => {
    render(<Table><TableHeader><TableRow><TableHead data-testid="th">Col</TableHead></TableRow></TableHeader></Table>)
    expect(screen.getByTestId('th').className).toContain('uppercase')
  })

  it('Table wraps in overflow container', () => {
    const { container } = render(<Table><TableBody><TableRow><TableCell>X</TableCell></TableRow></TableBody></Table>)
    expect(container.firstChild).toHaveClass('overflow-auto')
  })
})
