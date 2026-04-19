import { describe, it, expect, beforeEach } from 'vitest'
import { render } from '@testing-library/react'
import { PermissionGuard } from './PermissionGuard'
import { useAuthStore } from '@/store/auth'
import type { PermissionData } from '@/types'

describe('PermissionGuard', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth()
  })

  it('renders children when permission is granted', () => {
    const perms: PermissionData = {
      isSuperadmin: false,
      teamPermissions: { 1: ['team:read'] },
    }
    useAuthStore.getState().setPermissions(perms)

    const { getByText } = render(
      <PermissionGuard code="team:read" teamId={1}>
        <div>Protected Content</div>
      </PermissionGuard>,
    )
    expect(getByText('Protected Content')).toBeDefined()
  })

  it('returns null when permission is denied', () => {
    const perms: PermissionData = {
      isSuperadmin: false,
      teamPermissions: { 1: ['team:read'] },
    }
    useAuthStore.getState().setPermissions(perms)

    const { container } = render(
      <PermissionGuard code="team:write" teamId={1}>
        <div>Protected Content</div>
      </PermissionGuard>,
    )
    expect(container.innerHTML).toBe('')
  })

  it('renders children for superadmin regardless of code', () => {
    const perms: PermissionData = {
      isSuperadmin: true,
      teamPermissions: {},
    }
    useAuthStore.getState().setPermissions(perms)

    const { getByText } = render(
      <PermissionGuard code="admin:manage">
        <div>Admin Content</div>
      </PermissionGuard>,
    )
    expect(getByText('Admin Content')).toBeDefined()
  })

  it('returns null when no permissions loaded', () => {
    const { container } = render(
      <PermissionGuard code="team:read">
        <div>Protected Content</div>
      </PermissionGuard>,
    )
    expect(container.innerHTML).toBe('')
  })

  it('checks any team when teamId is omitted', () => {
    const perms: PermissionData = {
      isSuperadmin: false,
      teamPermissions: {
        1: ['team:read'],
        2: ['team:write'],
      },
    }
    useAuthStore.getState().setPermissions(perms)

    const { getByText } = render(
      <PermissionGuard code="team:write">
        <div>Has Write</div>
      </PermissionGuard>,
    )
    expect(getByText('Has Write')).toBeDefined()
  })
})
