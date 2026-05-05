import client from "./client";
import type {
  CreateMainItemReq,
  UpdateMainItemReq,
  ChangeStatusReq,
  MainItemFilter,
  MainItem,
  SubItem,
  PageResult,
} from "@/types";

export function createMainItemApi(
  teamBizKey: string,
  req: CreateMainItemReq,
): Promise<MainItem> {
  return client.post<never, MainItem>(`/teams/${teamBizKey}/main-items`, req);
}

export function listMainItemsApi(
  teamBizKey: string,
  filter?: MainItemFilter,
): Promise<PageResult<MainItem>> {
  return client.get<never, PageResult<MainItem>>(
    `/teams/${teamBizKey}/main-items`,
    { params: filter },
  );
}

export type MainItemDetailResp = MainItem & {
  subItems: SubItem[];
  achievements?: string[];
  blockers?: string[];
};

export function getMainItemApi(
  teamBizKey: string,
  bizKey: string,
): Promise<MainItemDetailResp> {
  return client.get<never, MainItemDetailResp>(
    `/teams/${teamBizKey}/main-items/${bizKey}`,
  );
}

export function updateMainItemApi(
  teamBizKey: string,
  bizKey: string,
  req: UpdateMainItemReq,
): Promise<MainItem> {
  return client.put<never, MainItem>(
    `/teams/${teamBizKey}/main-items/${bizKey}`,
    req,
  );
}

export function changeMainItemStatusApi(
  teamBizKey: string,
  bizKey: string,
  req: ChangeStatusReq,
): Promise<MainItem> {
  return client.put<never, MainItem>(
    `/teams/${teamBizKey}/main-items/${bizKey}/status`,
    req,
  );
}

export function getMainItemTransitionsApi(
  teamBizKey: string,
  bizKey: string,
): Promise<string[]> {
  return client
    .get<
      never,
      { transitions: string[] }
    >(`/teams/${teamBizKey}/main-items/${bizKey}/available-transitions`)
    .then((res) => res.transitions ?? []);
}
