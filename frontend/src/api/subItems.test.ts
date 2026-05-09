import { describe, it, expect, vi, beforeEach } from 'vitest'
import client from './client'
import { createSubItemApi } from './subItems'

vi.mock('./client', () => ({
  default: { post: vi.fn(), get: vi.fn(), put: vi.fn() },
}))

const mockClient = client as unknown as { post: ReturnType<typeof vi.fn> }

beforeEach(() => {
  vi.clearAllMocks()
})

describe('createSubItemApi', () => {
  it('bug: sends mainItemKey (not mainItemBizKey) in request body', async () => {
    mockClient.post.mockResolvedValue({ bizKey: 'sub-1' })

    await createSubItemApi('team-1', 'main-1', {
      title: 'Test sub item',
      priority: 'P1',
      assigneeKey: 'user-1',
      startDate: '2026-04-29',
      expectedEndDate: '2026-05-10',
    })

    const body = mockClient.post.mock.calls[0][1]
    expect(body).toHaveProperty('mainItemKey', 'main-1')
    expect(body).not.toHaveProperty('mainItemBizKey')
  })
})
