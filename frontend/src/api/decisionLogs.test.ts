import { describe, it, expect, vi, beforeEach } from "vitest";
import client from "./client";
import * as decisionLogsApi from "./decisionLogs";

vi.mock("./client", () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
    patch: vi.fn(),
  },
}));

const mockClient = vi.mocked(client, true);

describe("decisionLogs API", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("listDecisionLogsApi", () => {
    it("should GET decision logs with pagination params", async () => {
      mockClient.get.mockResolvedValue({
        items: [],
        total: 0,
        page: 1,
        size: 20,
      });
      await decisionLogsApi.listDecisionLogsApi("team-bk", "main-bk", 1, 20);
      expect(mockClient.get).toHaveBeenCalledWith(
        "/teams/team-bk/main-items/main-bk/decision-logs",
        { params: { page: 1, pageSize: 20 } },
      );
    });

    it("should GET decision logs without pagination params", async () => {
      mockClient.get.mockResolvedValue({
        items: [],
        total: 0,
        page: 1,
        size: 20,
      });
      await decisionLogsApi.listDecisionLogsApi("team-bk", "main-bk");
      expect(mockClient.get).toHaveBeenCalledWith(
        "/teams/team-bk/main-items/main-bk/decision-logs",
        { params: {} },
      );
    });
  });

  describe("createDecisionLogApi", () => {
    it("should POST to decision-logs endpoint with request body", async () => {
      mockClient.post.mockResolvedValue({});
      await decisionLogsApi.createDecisionLogApi("team-bk", "main-bk", {
        category: "technical",
        tags: ["tag1"],
        content: "decision content",
        logStatus: "draft",
      });
      expect(mockClient.post).toHaveBeenCalledWith(
        "/teams/team-bk/main-items/main-bk/decision-logs",
        {
          category: "technical",
          tags: ["tag1"],
          content: "decision content",
          logStatus: "draft",
        },
      );
    });
  });

  describe("updateDecisionLogApi", () => {
    it("should PUT to decision-logs/:bizKey endpoint with request body", async () => {
      mockClient.put.mockResolvedValue({});
      await decisionLogsApi.updateDecisionLogApi(
        "team-bk",
        "main-bk",
        "log-bk",
        {
          category: "risk",
          tags: ["updated"],
          content: "updated content",
        },
      );
      expect(mockClient.put).toHaveBeenCalledWith(
        "/teams/team-bk/main-items/main-bk/decision-logs/log-bk",
        {
          category: "risk",
          tags: ["updated"],
          content: "updated content",
        },
      );
    });
  });

  describe("publishDecisionLogApi", () => {
    it("should PATCH to decision-logs/:bizKey/publish endpoint", async () => {
      mockClient.patch.mockResolvedValue({});
      await decisionLogsApi.publishDecisionLogApi(
        "team-bk",
        "main-bk",
        "log-bk",
      );
      expect(mockClient.patch).toHaveBeenCalledWith(
        "/teams/team-bk/main-items/main-bk/decision-logs/log-bk/publish",
      );
    });
  });
});
