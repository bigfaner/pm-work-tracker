import client from './client'
import type {
  MilestoneMap,
  Milestone,
  CreateMilestoneMapReq,
  UpdateMilestoneMapReq,
  MilestoneMapFilter,
  CreateMilestoneReq,
  UpdateMilestoneReq,
  MilestoneTeamFilter,
  PageResult,
  ChangeStatusReq,
} from '@/types'

// --- MilestoneMap endpoints ---

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
  bizKey: string,
): Promise<MilestoneMap> {
  return client.get<never, MilestoneMap>(
    `/teams/${teamBizKey}/milestone-maps/${bizKey}`,
  )
}

export function updateMilestoneMapApi(
  teamBizKey: string,
  bizKey: string,
  req: UpdateMilestoneMapReq,
): Promise<MilestoneMap> {
  return client.put<never, MilestoneMap>(
    `/teams/${teamBizKey}/milestone-maps/${bizKey}`,
    req,
  )
}

export function deleteMilestoneMapApi(
  teamBizKey: string,
  bizKey: string,
): Promise<void> {
  return client.delete<never, void>(
    `/teams/${teamBizKey}/milestone-maps/${bizKey}`,
  )
}

export function changeMilestoneMapStatusApi(
  teamBizKey: string,
  bizKey: string,
  req: ChangeStatusReq,
): Promise<MilestoneMap> {
  return client.put<never, MilestoneMap>(
    `/teams/${teamBizKey}/milestone-maps/${bizKey}/status`,
    req,
  )
}

export function getMilestoneMapTransitionsApi(
  teamBizKey: string,
  bizKey: string,
): Promise<string[]> {
  return client
    .get<
      never,
      { transitions: string[] }
    >(`/teams/${teamBizKey}/milestone-maps/${bizKey}/available-transitions`)
    .then((res) => res.transitions ?? [])
}

// --- Milestone endpoints ---

export function createMilestoneApi(
  teamBizKey: string,
  mapBizKey: string,
  req: CreateMilestoneReq,
): Promise<Milestone> {
  return client.post<never, Milestone>(
    `/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`,
    req,
  )
}

export function listMilestonesByMapApi(
  teamBizKey: string,
  mapBizKey: string,
): Promise<{ items: Milestone[], total: number }> {
  return client.get<
    never,
    { items: Milestone[], total: number }
  >(`/teams/${teamBizKey}/milestone-maps/${mapBizKey}/milestones`)
}

export function listMilestonesByTeamApi(
  teamBizKey: string,
  filter?: MilestoneTeamFilter,
): Promise<{ items: Milestone[], total: number }> {
  return client.get<
    never,
    { items: Milestone[], total: number }
  >(`/teams/${teamBizKey}/milestones`, { params: filter })
}

export function getMilestoneApi(
  teamBizKey: string,
  bizKey: string,
): Promise<Milestone> {
  return client.get<never, Milestone>(
    `/teams/${teamBizKey}/milestones/${bizKey}`,
  )
}

export function updateMilestoneApi(
  teamBizKey: string,
  bizKey: string,
  req: UpdateMilestoneReq,
): Promise<Milestone> {
  return client.put<never, Milestone>(
    `/teams/${teamBizKey}/milestones/${bizKey}`,
    req,
  )
}

export function deleteMilestoneApi(
  teamBizKey: string,
  bizKey: string,
): Promise<void> {
  return client.delete<never, void>(
    `/teams/${teamBizKey}/milestones/${bizKey}`,
  )
}

export function changeMilestoneStatusApi(
  teamBizKey: string,
  bizKey: string,
  req: ChangeStatusReq,
): Promise<Milestone> {
  return client.put<never, Milestone>(
    `/teams/${teamBizKey}/milestones/${bizKey}/status`,
    req,
  )
}

export function getMilestoneTransitionsApi(
  teamBizKey: string,
  bizKey: string,
): Promise<string[]> {
  return client
    .get<
      never,
      { transitions: string[] }
    >(`/teams/${teamBizKey}/milestones/${bizKey}/available-transitions`)
    .then((res) => res.transitions ?? [])
}
