import { describe, it, expect, beforeEach, vi } from "vitest";
import { useAuthStore } from "./auth";
import type { User, PermissionData } from "@/types";

const mockUser: User = {
  bizKey: "1",
  username: "testuser",
  displayName: "Test User",
  createTime: "",
};

describe("useAuthStore", () => {
  beforeEach(() => {
    // Reset store state before each test
    useAuthStore.getState().clearAuth();
  });

  it("starts with null token and user", () => {
    const state = useAuthStore.getState();
    expect(state.token).toBeNull();
    expect(state.user).toBeNull();
  });

  it("setAuth populates token and user", () => {
    useAuthStore.getState().setAuth("jwt-token-123", mockUser);

    const state = useAuthStore.getState();
    expect(state.token).toBe("jwt-token-123");
    expect(state.user).toEqual(mockUser);
  });

  it("clearAuth resets token and user to null", () => {
    useAuthStore.getState().setAuth("jwt-token-123", mockUser);
    useAuthStore.getState().clearAuth();

    const state = useAuthStore.getState();
    expect(state.token).toBeNull();
    expect(state.user).toBeNull();
  });

  it("isAuthenticated returns false when no token", () => {
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
  });

  it("isAuthenticated returns true when token is set", () => {
    useAuthStore.getState().setAuth("jwt-token-123", mockUser);
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
  });

  it("isAuthenticated returns false after clearAuth", () => {
    useAuthStore.getState().setAuth("jwt-token-123", mockUser);
    useAuthStore.getState().clearAuth();
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
  });
});

describe("useAuthStore permissions", () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth();
  });

  describe("permissions state", () => {
    it("starts with null permissions", () => {
      const state = useAuthStore.getState();
      expect(state.permissions).toBeNull();
      expect(state.permissionsLoadedAt).toBeNull();
    });

    it("clearAuth resets permissions", () => {
      const perms: PermissionData = {
        teamPermissions: { "1": ["team:read"] },
      };
      useAuthStore.getState().setPermissions(perms);
      expect(useAuthStore.getState().permissions).not.toBeNull();

      useAuthStore.getState().clearAuth();
      expect(useAuthStore.getState().permissions).toBeNull();
      expect(useAuthStore.getState().permissionsLoadedAt).toBeNull();
    });
  });

  describe("hasPermission", () => {
    it("returns false when no permissions loaded", () => {
      expect(useAuthStore.getState().hasPermission("team:read")).toBe(false);
    });

    it("returns true for user with all codes (superadmin equivalent)", () => {
      const perms: PermissionData = {
        teamPermissions: { "1": ["any:permission"] },
      };
      useAuthStore.getState().setPermissions(perms);
      expect(useAuthStore.getState().hasPermission("any:permission")).toBe(
        true,
      );
    });

    it("returns true when teamId matches and team has the permission", () => {
      const perms: PermissionData = {
        teamPermissions: { "1": ["team:read", "team:write"] },
      };
      useAuthStore.getState().setPermissions(perms);
      expect(useAuthStore.getState().hasPermission("team:write", "1")).toBe(
        true,
      );
    });

    it("returns false when teamId matches but team lacks the permission", () => {
      const perms: PermissionData = {
        teamPermissions: { "1": ["team:read"] },
      };
      useAuthStore.getState().setPermissions(perms);
      expect(useAuthStore.getState().hasPermission("team:write", "1")).toBe(
        false,
      );
    });

    it("returns false when teamId is not in teamPermissions", () => {
      const perms: PermissionData = {
        teamPermissions: { "1": ["team:read"] },
      };
      useAuthStore.getState().setPermissions(perms);
      expect(useAuthStore.getState().hasPermission("team:read", "99")).toBe(
        false,
      );
    });

    it("returns true when no teamId and any team has the permission", () => {
      const perms: PermissionData = {
        teamPermissions: {
          "1": ["team:read"],
          "2": ["team:write", "item:create"],
        },
      };
      useAuthStore.getState().setPermissions(perms);
      expect(useAuthStore.getState().hasPermission("item:create")).toBe(true);
    });

    it("returns false when no teamId and no team has the permission", () => {
      const perms: PermissionData = {
        teamPermissions: { "1": ["team:read"] },
      };
      useAuthStore.getState().setPermissions(perms);
      expect(useAuthStore.getState().hasPermission("admin:manage")).toBe(false);
    });

    it("returns false when teamPermissions is empty", () => {
      const perms: PermissionData = {
        teamPermissions: {},
      };
      useAuthStore.getState().setPermissions(perms);
      expect(useAuthStore.getState().hasPermission("team:read")).toBe(false);
    });
  });

  describe("fetchPermissions", () => {
    it("calls API and stores permissions", async () => {
      const mockPerms: PermissionData = {
        teamPermissions: { "1": ["team:read", "team:write"] },
      };
      vi.spyOn(
        await import("@/api/permissions"),
        "getPermissionsApi",
      ).mockResolvedValue(mockPerms);

      await useAuthStore.getState().fetchPermissions();

      const state = useAuthStore.getState();
      expect(state.permissions).toEqual(mockPerms);
      expect(state.permissionsLoadedAt).toBeGreaterThan(0);
    });

    it("clears permissions on fetch error", async () => {
      vi.spyOn(
        await import("@/api/permissions"),
        "getPermissionsApi",
      ).mockRejectedValue(new Error("fail"));

      // Set some existing permissions first
      useAuthStore.getState().setPermissions({
        teamPermissions: { "1": ["team:read"] },
      });
      expect(useAuthStore.getState().permissions).not.toBeNull();

      await useAuthStore.getState().fetchPermissions();
      expect(useAuthStore.getState().permissions).toBeNull();
    });
  });
});
