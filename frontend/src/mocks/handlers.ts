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
  bizKey: '1',
  username: 'testuser',
  displayName: 'Test User',
  isSuperAdmin: false,
  createTime: '',
}

export const seedMembers: TeamMemberResp[] = [
  { id: 1, bizKey: '1', teamKey: '1', userBizKey: 'U001', displayName: 'Test User', username: 'testuser', role: 'pm', roleId: 1, roleName: 'pm', joinedAt: '2024-01-01' },
]

export function makeMainItem(overrides: Partial<MainItem> = {}): MainItem {
  return {
    bizKey: '1',
    teamKey: '1',
    code: 'MI-0001',
    title: 'Test Main Item',
    priority: 'P2',
    proposerKey: 'U001',
    assigneeKey: 'U001',
    planStartDate: null,
    expectedEndDate: null,
    actualEndDate: null,
    itemStatus: 'progressing',
    completion: 50,
    createTime: '2024-06-01T10:00:00Z',
    dbUpdateTime: '2024-06-01T10:00:00Z',
    ...overrides,
  }
}

export function makeItemPool(overrides: Partial<ItemPool> = {}): ItemPool {
  return {
    bizKey: '1',
    teamKey: '1',
    title: 'Test Pool Item',
    background: '',
    expectedOutput: '',
    submitterKey: 'U001',
    poolStatus: '待分配',
    assignedMainKey: null,
    assignedSubKey: null,
    assignedMainCode: '',
    assignedMainTitle: '',
    assigneeKey: null,
    rejectReason: '',
    reviewedAt: null,
    reviewerKey: null,
    createTime: '2024-06-01T10:00:00Z',
    dbUpdateTime: '2024-06-01T10:00:00Z',
    ...overrides,
  }
}

// --- Handlers ---

export const handlers = [
  // Auth: login
  http.post('/v1/auth/login', async ({ request }) => {
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
  http.get('/v1/teams/:teamId/members', () => {
    return HttpResponse.json({ code: 0, data: seedMembers })
  }),

  // Main items: list
  http.get('/v1/teams/:teamId/main-items', ({ request }) => {
    const url = new URL(request.url)
    const priority = url.searchParams.get('priority')
    const archived = url.searchParams.get('archived')

    let items = [
      makeMainItem({ bizKey: '1', code: 'MI-0001', title: 'Alpha', priority: 'P1', itemStatus: 'progressing', completion: 50 }),
      makeMainItem({ bizKey: '2', code: 'MI-0002', title: 'Beta', priority: 'P2', itemStatus: 'pending', completion: 0 }),
      makeMainItem({ bizKey: '3', code: 'MI-0003', title: 'Gamma', priority: 'P3', itemStatus: 'completed', completion: 100 }),
    ]

    if (priority) {
      items = items.filter((i) => i.priority === priority)
    }
    if (archived === 'false') {
      // seed data has no archived items, keep all
    }

    const page: PageResult<MainItem> = { items, total: items.length, page: 1, size: 20 }
    return HttpResponse.json({ code: 0, data: page })
  }),

  // Item pool: list
  http.get('/v1/teams/:teamId/item-pool', ({ request }) => {
    const url = new URL(request.url)
    const status = url.searchParams.get('status')

    let items = [
      makeItemPool({ bizKey: '1', title: 'Pending Item', poolStatus: '待分配' }),
      makeItemPool({ bizKey: '2', title: 'Assigned Item', poolStatus: '已分配', assignedMainKey: '1', assigneeKey: 'U001' }),
      makeItemPool({ bizKey: '3', title: 'Rejected Item', poolStatus: '已拒绝', rejectReason: 'Not suitable' }),
    ]

    if (status) {
      items = items.filter((i) => i.poolStatus === status)
    }

    const page: PageResult<ItemPool> = { items, total: items.length, page: 1, size: 20 }
    return HttpResponse.json({ code: 0, data: page })
  }),

  // Item pool: assign
  http.post('/v1/teams/:teamId/item-pool/:poolId/assign', async ({ params }) => {
    const poolId = Number(params.poolId)
    const resp: AssignItemPoolResp = { mainItemBizKey: `mi-${100 + poolId}`, subItemBizKey: `si-${1000 + poolId}` }
    return HttpResponse.json({ code: 0, data: resp })
  }),

  // Item pool: reject
  http.post('/v1/teams/:teamId/item-pool/:poolId/reject', async ({ params }) => {
    const updated = makeItemPool({ bizKey: String(params.poolId), poolStatus: '已拒绝' })
    return HttpResponse.json({ code: 0, data: updated })
  }),

  // Generic 401 handler (for testing auth-clear-on-401)
  http.get('/v1/trigger-401', () => {
    return HttpResponse.json(
      { code: 'UNAUTHORIZED', message: 'token expired' },
      { status: 401 },
    )
  }),
]
