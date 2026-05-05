import client from "./client";
import type { ReportPreviewResp } from "@/types";

export function getWeeklyReportPreviewApi(
  teamBizKey: string,
  weekStart: string,
): Promise<ReportPreviewResp> {
  return client.get<never, ReportPreviewResp>(
    `/teams/${teamBizKey}/reports/weekly/preview`,
    { params: { weekStart } },
  );
}

export function exportWeeklyReportApi(
  teamBizKey: string,
  weekStart: string,
): Promise<Blob> {
  return client.get<never, Blob>(`/teams/${teamBizKey}/reports/weekly/export`, {
    params: { weekStart },
    responseType: "blob",
  });
}
