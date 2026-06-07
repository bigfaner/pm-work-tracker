import client from './client'
import type {
  CreateMilestoneMapReq,
  UpdateMilestoneMapReq,
  MilestoneMapFilter,
  MilestoneMap,
  CreateMilestoneReq,
  UpdateMilestoneReq,
  MilestoneListFilter,
  Milestone,
  PageResult,
} from '@/types'

// --- MilestoneMap ---

export function createMilestoneMapApi(
  teamBizKey: string,
  req: CreateMilestoneMapReq,
): Promise<MilestoneMap> {
  return client.post<never, MilestoneMap>(
    `/teams/${teamBizKey}/milestone-maps`,
    req,
  )
}

export function listMilestoneMapsApi(
  teamBizKey: string,
  filter?: MilestoneMapFilter,
): Promise<PageResult<MilestoneMap>> {
  return client.get<never, PageResult<MilestoneMap>>(
    `/teams/${teamBizKey}/milestone-maps`,
    { params: filter },
  )
}

export function getMilestoneMapApi(
  teamBizKey: string,
  mapId: string,
): Promise<MilestoneMap> {
  return client.get<never, MilestoneMap>(
    `/teams/${teamBizKey}/milestone-maps/${mapId}`,
  )
}

export function updateMilestoneMapApi(
  teamBizKey: string,
  mapId: string,
  req: UpdateMilestoneMapReq,
): Promise<MilestoneMap> {
  return client.put<never, MilestoneMap>(
    `/teams/${teamBizKey}/milestone-maps/${mapId}`,
    req,
  )
}

export function deleteMilestoneMapApi(
  teamBizKey: string,
  mapId: string,
): Promise<{ message: string }> {
  return client.delete<never, { message: string }>(
    `/teams/${teamBizKey}/milestone-maps/${mapId}`,
  )
}

export function changeMilestoneMapStatusApi(
  teamBizKey: string,
  mapId: string,
  req: { status: string },
): Promise<MilestoneMap> {
  return client.put<never, MilestoneMap>(
    `/teams/${teamBizKey}/milestone-maps/${mapId}/status`,
    req,
  )
}

export function getMilestoneMapTransitionsApi(
  teamBizKey: string,
  mapId: string,
): Promise<string[]> {
  return client
    .get<never, { transitions: string[] }>(
      `/teams/${teamBizKey}/milestone-maps/${mapId}/available-transitions`,
    )
    .then((res) => res.transitions ?? [])
}

// --- Milestone ---

export function createMilestoneApi(
  teamBizKey: string,
  mapId: string,
  req: CreateMilestoneReq,
): Promise<Milestone> {
  return client.post<never, Milestone>(
    `/teams/${teamBizKey}/milestone-maps/${mapId}/milestones`,
    req,
  )
}

export function listMilestonesByMapApi(
  teamBizKey: string,
  mapId: string,
): Promise<{ items: Milestone[]; total: number }> {
  return client.get<never, { items: Milestone[]; total: number }>(
    `/teams/${teamBizKey}/milestone-maps/${mapId}/milestones`,
  )
}

export function listMilestonesByTeamApi(
  teamBizKey: string,
  filter?: MilestoneListFilter,
): Promise<{ items: Milestone[]; total: number }> {
  return client.get<never, { items: Milestone[]; total: number }>(
    `/teams/${teamBizKey}/milestones`,
    { params: filter },
  )
}

export function getMilestoneApi(
  teamBizKey: string,
  milestoneId: string,
): Promise<Milestone> {
  return client.get<never, Milestone>(
    `/teams/${teamBizKey}/milestones/${milestoneId}`,
  )
}

export function updateMilestoneApi(
  teamBizKey: string,
  milestoneId: string,
  req: UpdateMilestoneReq,
): Promise<Milestone> {
  return client.put<never, Milestone>(
    `/teams/${teamBizKey}/milestones/${milestoneId}`,
    req,
  )
}

export function deleteMilestoneApi(
  teamBizKey: string,
  milestoneId: string,
): Promise<{ message: string }> {
  return client.delete<never, { message: string }>(
    `/teams/${teamBizKey}/milestones/${milestoneId}`,
  )
}

export function changeMilestoneStatusApi(
  teamBizKey: string,
  milestoneId: string,
  req: { status: string },
): Promise<Milestone> {
  return client.put<never, Milestone>(
    `/teams/${teamBizKey}/milestones/${milestoneId}/status`,
    req,
  )
}

export function getMilestoneTransitionsApi(
  teamBizKey: string,
  milestoneId: string,
): Promise<string[]> {
  return client
    .get<never, { transitions: string[] }>(
      `/teams/${teamBizKey}/milestones/${milestoneId}/available-transitions`,
    )
    .then((res) => res.transitions ?? [])
}
