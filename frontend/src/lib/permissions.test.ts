import { describe, it, expect } from 'vitest'
import { PERMISSION_GROUPS, ALL_PERMISSION_CODES } from './permissions'

describe('PERMISSION_GROUPS', () => {
  describe('user group', () => {
    it('has 4 permission codes', () => {
      const userGroup = PERMISSION_GROUPS.find((g) => g.key === 'user')
      expect(userGroup).toBeDefined()
      expect(userGroup!.permissions).toHaveLength(4)
    })

    it('contains user:list', () => {
      const userGroup = PERMISSION_GROUPS.find((g) => g.key === 'user')!
      expect(userGroup.permissions.some((p) => p.value === 'user:list')).toBe(true)
    })

    it('contains user:read with detail label', () => {
      const userGroup = PERMISSION_GROUPS.find((g) => g.key === 'user')!
      const readPerm = userGroup.permissions.find((p) => p.value === 'user:read')
      expect(readPerm).toBeDefined()
      expect(readPerm!.label).toContain('详情')
    })

    it('contains user:assign_role', () => {
      const userGroup = PERMISSION_GROUPS.find((g) => g.key === 'user')!
      expect(userGroup.permissions.some((p) => p.value === 'user:assign_role')).toBe(true)
    })

    it('does not contain user:manage_role', () => {
      const userGroup = PERMISSION_GROUPS.find((g) => g.key === 'user')!
      expect(userGroup.permissions.some((p) => p.value === 'user:manage_role')).toBe(false)
    })
  })

  describe('role group', () => {
    it('exists', () => {
      const roleGroup = PERMISSION_GROUPS.find((g) => g.key === 'role')
      expect(roleGroup).toBeDefined()
    })

    it('has 4 permission codes', () => {
      const roleGroup = PERMISSION_GROUPS.find((g) => g.key === 'role')
      expect(roleGroup!.permissions).toHaveLength(4)
    })

    it('contains role:read, role:create, role:update, role:delete', () => {
      const roleGroup = PERMISSION_GROUPS.find((g) => g.key === 'role')!
      const codes = roleGroup.permissions.map((p) => p.value)
      expect(codes).toContain('role:read')
      expect(codes).toContain('role:create')
      expect(codes).toContain('role:update')
      expect(codes).toContain('role:delete')
    })
  })
})

describe('ALL_PERMISSION_CODES', () => {
  it('includes all new user codes', () => {
    expect(ALL_PERMISSION_CODES).toContain('user:list')
    expect(ALL_PERMISSION_CODES).toContain('user:read')
    expect(ALL_PERMISSION_CODES).toContain('user:update')
    expect(ALL_PERMISSION_CODES).toContain('user:assign_role')
  })

  it('includes all role codes', () => {
    expect(ALL_PERMISSION_CODES).toContain('role:read')
    expect(ALL_PERMISSION_CODES).toContain('role:create')
    expect(ALL_PERMISSION_CODES).toContain('role:update')
    expect(ALL_PERMISSION_CODES).toContain('role:delete')
  })

  it('does not include removed user:manage_role', () => {
    expect(ALL_PERMISSION_CODES).not.toContain('user:manage_role')
  })
})
