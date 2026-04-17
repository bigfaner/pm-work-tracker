export interface User {
  id: number;
  username: string;
  display_name: string;
  is_super_admin: boolean;
  can_create_team: boolean;
  created_at: string;
  updated_at: string;
}

export interface Team {
  id: number;
  name: string;
  description: string;
  pm_id: number;
  created_at: string;
  updated_at: string;
}

export interface TeamMember {
  id: number;
  team_id: number;
  user_id: number;
  role: string;
  joined_at: string;
  created_at: string;
  updated_at: string;
}

export interface MainItem {
  id: number;
  team_id: number;
  code: string;
  title: string;
  priority: string;
  proposer_id: number;
  assignee_id: number | null;
  start_date: string | null;
  expected_end_date: string | null;
  actual_end_date: string | null;
  status: string;
  completion: number;
  is_key_item: boolean;
  delay_count: number;
  archived_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface SubItem {
  id: number;
  team_id: number;
  main_item_id: number;
  title: string;
  description: string;
  priority: string;
  assignee_id: number | null;
  start_date: string | null;
  expected_end_date: string | null;
  actual_end_date: string | null;
  status: string;
  completion: number;
  is_key_item: boolean;
  delay_count: number;
  weight: number;
  created_at: string;
  updated_at: string;
}

export interface ProgressRecord {
  id: number;
  sub_item_id: number;
  team_id: number;
  author_id: number;
  completion: number;
  achievement: string;
  blocker: string;
  lesson: string;
  is_pm_correct: boolean;
  created_at: string;
}

export interface ItemPool {
  id: number;
  team_id: number;
  title: string;
  background: string;
  expected_output: string;
  submitter_id: number;
  status: string;
  assigned_main_id: number | null;
  assigned_sub_id: number | null;
  assignee_id: number | null;
  reject_reason: string;
  reviewed_at: string | null;
  reviewer_id: number | null;
  created_at: string;
  updated_at: string;
}
