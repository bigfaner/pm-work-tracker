import client from "./client";
import type { AppendProgressReq, ProgressRecord } from "@/types";

export function appendProgressApi(
  teamBizKey: string,
  subBizKey: string,
  req: AppendProgressReq,
): Promise<ProgressRecord> {
  return client.post<never, ProgressRecord>(
    `/teams/${teamBizKey}/sub-items/${subBizKey}/progress`,
    req,
  );
}

export function listProgressApi(
  teamBizKey: string,
  subBizKey: string,
): Promise<ProgressRecord[]> {
  return client.get<never, ProgressRecord[]>(
    `/teams/${teamBizKey}/sub-items/${subBizKey}/progress`,
  );
}
