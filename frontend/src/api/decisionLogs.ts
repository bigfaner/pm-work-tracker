import client from "./client";
import type { PageResult } from "@/types";

// Decision Log types

export interface DecisionLog {
  bizKey: string;
  mainItemKey: string;
  category: string;
  tags: string[];
  content: string;
  logStatus: "draft" | "published";
  createdBy: string;
  creatorName: string;
  createTime: string;
  updateTime: string;
}

export interface CreateDecisionLogReq {
  category: string;
  tags: string[];
  content: string;
  logStatus: "draft" | "published";
}

export interface UpdateDecisionLogReq {
  category: string;
  tags: string[];
  content: string;
}

// API functions

export function listDecisionLogsApi(
  teamBizKey: string,
  mainBizKey: string,
  page?: number,
  pageSize?: number,
): Promise<PageResult<DecisionLog>> {
  return client.get<never, PageResult<DecisionLog>>(
    `/teams/${teamBizKey}/main-items/${mainBizKey}/decision-logs`,
    { params: { page, pageSize } },
  );
}

export function createDecisionLogApi(
  teamBizKey: string,
  mainBizKey: string,
  req: CreateDecisionLogReq,
): Promise<DecisionLog> {
  return client.post<never, DecisionLog>(
    `/teams/${teamBizKey}/main-items/${mainBizKey}/decision-logs`,
    req,
  );
}

export function updateDecisionLogApi(
  teamBizKey: string,
  mainBizKey: string,
  bizKey: string,
  req: UpdateDecisionLogReq,
): Promise<DecisionLog> {
  return client.put<never, DecisionLog>(
    `/teams/${teamBizKey}/main-items/${mainBizKey}/decision-logs/${bizKey}`,
    req,
  );
}

export function publishDecisionLogApi(
  teamBizKey: string,
  mainBizKey: string,
  bizKey: string,
): Promise<DecisionLog> {
  return client.patch<never, DecisionLog>(
    `/teams/${teamBizKey}/main-items/${mainBizKey}/decision-logs/${bizKey}/publish`,
  );
}
