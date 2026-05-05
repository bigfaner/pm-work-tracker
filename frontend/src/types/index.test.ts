import { describe, it, expect } from "vitest";
import type {
  User,
  Team,
  MainItem,
  SubItem,
  ProgressRecord,
  ItemPool,
  PermissionData,
  Role,
  PermissionGroup,
  PermissionItem,
  AdminTeam,
} from "@/types";

describe("shared TypeScript interfaces", () => {
  it("should define a valid User", () => {
    const user: User = {
      bizKey: "1",
      username: "testuser",
      displayName: "Test User",
      isSuperAdmin: false,
      createTime: "",
    };
    expect(user.username).toBe("testuser");
    expect(user.isSuperAdmin).toBe(false);
  });

  it("should define a valid Team", () => {
    const team: Team = {
      bizKey: "1",
      name: "Team Alpha",
      description: "A team",
      code: "",
      pmKey: "1",
      createdAt: "2024-01-01",
      updatedAt: "2024-01-01",
    };
    expect(team.name).toBe("Team Alpha");
  });

  it("should define a valid MainItem", () => {
    const item: MainItem = {
      bizKey: "1",
      teamKey: "1",
      code: "A001",
      title: "Feature A",
      priority: "P0",
      proposerKey: "1",
      assigneeKey: null,
      planStartDate: null,
      expectedEndDate: null,
      actualEndDate: null,
      itemStatus: "pending",
      completion: 0,
      createTime: "2024-01-01",
      dbUpdateTime: "2024-01-01",
    };
    expect(item.code).toBe("A001");
  });

  it("should define a valid SubItem", () => {
    const sub: SubItem = {
      bizKey: "1",
      teamKey: "1",
      mainItemKey: "1",
      code: "A001-01",
      title: "Sub task",
      itemDesc: "Details",
      priority: "P1",
      assigneeKey: "2",
      planStartDate: "2026-01-01T00:00:00Z",
      expectedEndDate: "2026-01-15T00:00:00Z",
      actualEndDate: null,
      itemStatus: "progressing",
      completion: 50,
      weight: 1,
      createTime: "2024-01-01",
      dbUpdateTime: "2024-01-01",
    };
    expect(sub.weight).toBe(1);
  });

  it("should define a valid ProgressRecord", () => {
    const record: ProgressRecord = {
      subItemKey: "1",
      teamKey: "1",
      authorKey: "1",
      completion: 60,
      achievement: "Done something",
      blocker: "",
      lesson: "",
      isPMCorrect: false,
      createTime: "2024-01-01",
    };
    expect(record.completion).toBe(60);
  });

  it("should define a valid ItemPool", () => {
    const pool: ItemPool = {
      bizKey: "1",
      teamKey: "1",
      title: "Pool item",
      background: "Context",
      expectedOutput: "Result",
      submitterKey: "1",
      poolStatus: "待分配",
      assignedMainKey: null,
      assignedSubKey: null,
      assignedMainCode: "",
      assignedMainTitle: "",
      assigneeKey: null,
      rejectReason: "",
      reviewedAt: null,
      reviewerKey: null,
      createTime: "2024-01-01",
      dbUpdateTime: "2024-01-01",
    };
    expect(pool.poolStatus).toBe("待分配");
  });

  it("should define a valid PermissionData", () => {
    const permData: PermissionData = {
      isSuperAdmin: false,
      teamPermissions: { 1: ["team:read", "team:write"] },
    };
    expect(permData.isSuperAdmin).toBe(false);
    expect(permData.teamPermissions[1]).toContain("team:read");
  });

  it("should define a valid Role", () => {
    const role: Role = {
      bizKey: "1",
      roleName: "PM",
      roleDesc: "Project Manager",
      isPreset: true,
      permissionCount: 5,
      memberCount: 3,
      createTime: "2024-01-01",
    };
    expect(role.roleName).toBe("PM");
    expect(role.isPreset).toBe(true);
  });

  it("should define a valid PermissionGroup", () => {
    const group: PermissionGroup = {
      resource: "team",
      actions: [
        { code: "team:read", description: "View team" },
        { code: "team:write", description: "Edit team" },
      ],
    };
    expect(group.actions).toHaveLength(2);
  });

  it("should define a valid PermissionItem", () => {
    const item: PermissionItem = {
      code: "item:create",
      description: "Create items",
    };
    expect(item.code).toBe("item:create");
  });

  it("should define AdminTeam with flat pmDisplayName", () => {
    const team: AdminTeam = {
      bizKey: "1",
      name: "Team A",
      pmDisplayName: "Alice",
      memberCount: 5,
      mainItemCount: 10,
      createTime: "2024-01-01",
    };
    expect(team.pmDisplayName).toBe("Alice");
  });
});
