import client from './client'
import type { LoginReq, LoginResp } from '@/types'

export function loginApi(req: LoginReq): Promise<LoginResp> {
  return client.post<never, LoginResp>('/auth/login', req)
}

export function logoutApi(): Promise<void> {
  return client.post('/auth/logout')
}
