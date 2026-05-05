declare module "frappe-gantt" {
  interface GanttTask {
    id: string;
    name: string;
    start: string;
    end: string;
    progress: number;
    dependencies?: string | string[];
    custom_class?: string;
  }

  interface GanttOptions {
    header_height?: number;
    column_width?: number;
    step?: number;
    view_modes?: string[];
    bar_height?: number;
    bar_corner_radius?: number;
    arrow_curve?: number;
    padding?: number;
    view_mode?: string;
    date_format?: string;
    popup_trigger?: string;
    custom_popup_html?: (task: GanttTask) => string;
    language?: string;
    on_click?: (task: GanttTask) => void;
    on_date_change?: (task: GanttTask, start: Date, end: Date) => void;
    on_progress_change?: (task: GanttTask, progress: number) => void;
    on_view_change?: (mode: string) => void;
  }

  class Gantt {
    constructor(
      wrapper: HTMLElement | string,
      tasks: GanttTask[],
      options?: GanttOptions,
    );
    change_view_mode(mode?: string): void;
    refresh(tasks: GanttTask[]): void;
  }

  export default Gantt;
}

declare module "frappe-gantt/dist/frappe-gantt.css";
