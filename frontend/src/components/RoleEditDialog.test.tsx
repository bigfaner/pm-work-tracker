import { describe, it, expect, vi, beforeEach, beforeAll, afterAll, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { server } from '@/mocks/server'
import { http, HttpResponse } from 'msw'
import RoleEditDialog from './RoleEditDialog'

beforeAll(() => server.listen({ onUnhandledRequest: 'bypass' }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

function createQueryClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
}

function renderDialog(props: { open: boolean; onOpenChange: (v: boolean) => void; roleId?: number | null }) {
  const qc = createQueryClient()
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <RoleEditDialog {...props} />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

function setupHandlers() {
  server.use(
    http.get('/v1/admin/roles/:id', ({ params }) => {
      const id = Number(params.id)
      if (id === 1) {
        return HttpResponse.json({
          code: 0,
          data: {
            id: 1,
            name: 'superadmin',
            description: '超级管理员',
            isPreset: true,
            permissions: [
              { code: 'team:create', description: '创建团队' },
              { code: 'team:read', description: '查看团队信息' },
            ],
            memberCount: 2,
            createdAt: '2026-04-01T00:00:00Z',
          },
        })
      }
      if (id === 4) {
        return HttpResponse.json({
          code: 0,
          data: {
            id: 4,
            name: 'viewer',
            description: '只读查看者',
            isPreset: false,
            permissions: [
              { code: 'team:read', description: '查看团队信息' },
              { code: 'main_item:read', description: '查看主事项' },
            ],
            memberCount: 0,
            createdAt: '2026-04-15T12:00:00Z',
          },
        })
      }
      return HttpResponse.json(
        { code: 'ERR_ROLE_NOT_FOUND', message: 'not found' },
        { status: 404 },
      )
    }),

    http.post('/v1/admin/roles', async ({ request }) => {
      const body = (await request.json()) as Record<string, unknown>
      const name = body.name as string
      if (name === 'pm') {
        return HttpResponse.json(
          { code: 'ERR_ROLE_NAME_EXISTS', message: '角色名称已存在' },
          { status: 409 },
        )
      }
      return HttpResponse.json({
        code: 0,
        data: {
          id: 100,
          name,
          description: body.description || '',
          isPreset: false,
          permissionCount: (body.permissionCodes as string[]).length,
          memberCount: 0,
          createdAt: '2026-04-19T12:00:00Z',
        },
      })
    }),

    http.put('/v1/admin/roles/:id', async ({ params }) => {
      const id = Number(params.id)
      return HttpResponse.json({
        code: 0,
        data: {
          id,
          name: 'viewer',
          description: 'updated',
          isPreset: false,
          permissions: [{ code: 'team:read', description: '查看团队信息' }],
          memberCount: 0,
          createdAt: '2026-04-15T12:00:00Z',
        },
      })
    }),
  )
}

describe('RoleEditDialog', () => {
  beforeEach(() => {
    setupHandlers()
  })

  it('shows create mode title', () => {
    renderDialog({ open: true, onOpenChange: vi.fn() })
    expect(screen.getByText('创建角色')).toBeInTheDocument()
  })

  it('shows edit mode title with role name', async () => {
    renderDialog({ open: true, onOpenChange: vi.fn(), roleId: 4 })
    await waitFor(() => {
      expect(screen.getByText(/编辑角色.*viewer/)).toBeInTheDocument()
    })
  })

  it('pre-fills form in edit mode', async () => {
    renderDialog({ open: true, onOpenChange: vi.fn(), roleId: 4 })
    await waitFor(() => {
      const input = screen.getByPlaceholderText('请输入角色名称（2-50 个字符）') as HTMLInputElement
      expect(input.value).toBe('viewer')
    })
  })

  it('disables name input for preset roles', async () => {
    renderDialog({ open: true, onOpenChange: vi.fn(), roleId: 1 })
    await waitFor(() => {
      const input = screen.getByPlaceholderText('请输入角色名称（2-50 个字符）') as HTMLInputElement
      expect(input.disabled).toBe(true)
    })
  })

  it('validates name length on save', async () => {
    const user = userEvent.setup()
    const onOpenChange = vi.fn()
    renderDialog({ open: true, onOpenChange })

    const nameInput = screen.getByPlaceholderText('请输入角色名称（2-50 个字符）')
    await user.type(nameInput, 'a')
    await user.click(screen.getByText('保存'))

    await waitFor(() => {
      expect(screen.getByText('角色名称需要 2-50 个字符')).toBeInTheDocument()
    })
  })

  it('shows error for duplicate role name', async () => {
    const user = userEvent.setup()
    const onOpenChange = vi.fn()
    renderDialog({ open: true, onOpenChange })

    const nameInput = screen.getByPlaceholderText('请输入角色名称（2-50 个字符）')
    await user.type(nameInput, 'pm')

    // Select at least one permission - click the first checkbox option "创建团队"
    await user.click(screen.getByText('创建团队'))

    await user.click(screen.getByText('保存'))

    await waitFor(() => {
      expect(screen.getByText('角色名称已存在')).toBeInTheDocument()
    })
  })

  it('has description textarea', () => {
    renderDialog({ open: true, onOpenChange: vi.fn() })
    expect(screen.getByPlaceholderText('请输入角色描述（最多 200 字符）')).toBeInTheDocument()
  })
})
