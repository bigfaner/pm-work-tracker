import { http, HttpResponse } from 'msw'
import type {
  User,
  LoginResp,
  MainItem,
  ItemPool,
  TeamMemberResp,
  PageResult,
  AssignItemPoolResp,
} from '@/types'

// --- Seed data ---

export const seedUser: User = {
  id: 1,
  username: 'testuser',
  displayName: 'Test User',
  isSuperAdmin: false,
  canCreateTeam: false,
}

export const seedMembers: TeamMemberResp[] = [
  { id: 1, teamId: 1, userId: 1, displayName: 'Test User', username: 'testuser', role: 'pm', roleId: 1, roleName: 'pm', joinedAt: '2024-01-01' },
]

export function makeMainItem(overrides: Partial<MainItem> = {}): MainItem {
  return {
    id: 1,
    teamId: 1,
    code: 'MI-0001',
    title: 'Test Main Item',
    priority: 'P2',
    proposerId: 1,
    assigneeId: 1,
    startDate: null,
    expectedEndDate: null,
    actualEndDate: null,
    status: '进行中',
    completion: 50,
    isKeyItem: false,
    delayCount: 0,
    archivedAt: null,
    createdAt: '2024-06-01T10:00:00Z',
    updatedAt: '2024-06-01T10:00:00Z',
    ...overrides,
  }
}

export function makeItemPool(overrides: Partial<ItemPool> = {}): ItemPool {
  return {
    id: 1,
    teamId: 1,
    title: 'Test Pool Item',
    background: '',
    expectedOutput: '',
    submitterId: 1,
    status: '待分配',
    assignedMainId: null,
    assignedSubId: null,
    assignedMainCode: '',
    assignedMainTitle: '',
    assigneeId: null,
    rejectReason: '',
    reviewedAt: null,
    reviewerId: null,
    createdAt: '2024-06-01T10:00:00Z',
    updatedAt: '2024-06-01T10:00:00Z',
    ...overrides,
  }
}

// --- Handlers ---

export const handlers = [
  // Auth: login
  http.post('/api/v1/auth/login', async ({ request }) => {
    const body = (await request.json()) as { username: string; password: string }

    if (body.username === 'testuser' && body.password === 'password123') {
      const resp: LoginResp = { token: 'jwt-token-123', user: seedUser }
      return HttpResponse.json({ code: 0, data: resp })
    }

    return HttpResponse.json(
      { code: 'UNAUTHORIZED', message: '账号或密码错误' },
      { status: 401 },
    )
  }),

  // Teams: list members
  http.get('/api/v1/teams/:teamId/members', () => {
    return HttpResponse.json({ code: 0, data: seedMembers })
  }),

  // Main items: list
  http.get('/api/v1/teams/:teamId/main-items', ({ request }) => {
    const url = new URL(request.url)
    const priority = url.searchParams.get('priority')
    const archived = url.searchParams.get('archived')

    let items = [
      makeMainItem({ id: 1, code: 'MI-0001', title: 'Alpha', priority: 'P1', status: '进行中', completion: 50 }),
      makeMainItem({ id: 2, code: 'MI-0002', title: 'Beta', priority: 'P2', status: '未开始', completion: 0 }),
      makeMainItem({ id: 3, code: 'MI-0003', title: 'Gamma', priority: 'P3', status: '已完成', completion: 100 }),
    ]

    if (priority) {
      items = items.filter((i) => i.priority === priority)
    }
    if (archived === 'false') {
      // seed data has no archived items, keep all
    }

    const page: PageResult<MainItem> = { items, total: items.length, page: 1, pageSize: 20 }
    return HttpResponse.json({ code: 0, data: page })
  }),

  // Item pool: list
  http.get('/api/v1/teams/:teamId/item-pool', ({ request }) => {
    const url = new URL(request.url)
    const status = url.searchParams.get('status')

    let items = [
      makeItemPool({ id: 1, title: 'Pending Item', status: '待分配' }),
      makeItemPool({ id: 2, title: 'Assigned Item', status: '已分配', assignedMainId: 1, assigneeId: 1 }),
      makeItemPool({ id: 3, title: 'Rejected Item', status: '已拒绝', rejectReason: 'Not suitable' }),
    ]

    if (status) {
      items = items.filter((i) => i.status === status)
    }

    const page: PageResult<ItemPool> = { items, total: items.length, page: 1, pageSize: 20 }
    return HttpResponse.json({ code: 0, data: page })
  }),

  // Item pool: assign
  http.post('/api/v1/teams/:teamId/item-pool/:poolId/assign', async ({ params }) => {
    const poolId = Number(params.poolId)
    const resp: AssignItemPoolResp = { mainItemId: 100 + poolId, subItemId: 1000 + poolId }
    return HttpResponse.json({ code: 0, data: resp })
  }),

  // Item pool: reject
  http.post('/api/v1/teams/:teamId/item-pool/:poolId/reject', async ({ params }) => {
    const updated = makeItemPool({ id: Number(params.poolId), status: '已拒绝' })
    return HttpResponse.json({ code: 0, data: updated })
  }),

  // Generic 401 handler (for testing auth-clear-on-401)
  http.get('/api/v1/trigger-401', () => {
    return HttpResponse.json(
      { code: 'UNAUTHORIZED', message: 'token expired' },
      { status: 401 },
    )
  }),
]
