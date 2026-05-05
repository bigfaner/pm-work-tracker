import client from "./client";
import type {
  TableFilter,
  WeeklyViewResponse,
  GanttViewResp,
  TableRow,
  PageResult,
} from "@/types";

export function getWeeklyViewApi(
  teamBizKey: string,
  weekStart: string,
): Promise<WeeklyViewResponse> {
  return client.get<never, WeeklyViewResponse>(
    `/teams/${teamBizKey}/views/weekly`,
    { params: { weekStart } },
  );
}

export function getGanttViewApi(
  teamBizKey: string,
  status?: string,
): Promise<GanttViewResp> {
  return client.get<never, GanttViewResp>(`/teams/${teamBizKey}/views/gantt`, {
    params: { status },
  });
}

export function getTableViewApi(
  teamBizKey: string,
  filter?: TableFilter,
): Promise<PageResult<TableRow>> {
  return client.get<never, PageResult<TableRow>>(
    `/teams/${teamBizKey}/views/table`,
    { params: filter },
  );
}

export function exportTableCsvApi(
  teamBizKey: string,
  filter?: TableFilter,
): Promise<Blob> {
  return client.get<never, Blob>(`/teams/${teamBizKey}/views/table/export`, {
    params: filter,
    responseType: "blob",
  });
}
