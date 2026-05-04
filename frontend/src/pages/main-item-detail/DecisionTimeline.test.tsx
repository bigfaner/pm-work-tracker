import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import { DecisionTimeline } from "./DecisionTimeline";
import * as decisionLogsApi from "@/api/decisionLogs";
import { usePermission } from "@/hooks/usePermission";
import { useAuthStore } from "@/store/auth";

// Mock the API module
vi.mock("@/api/decisionLogs", () => ({
  listDecisionLogsApi: vi.fn(),
}));

// Mock usePermission hook
vi.mock("@/hooks/usePermission", () => ({
  usePermission: vi.fn().mockReturnValue(true),
}));

// Mock auth store
vi.mock("@/store/auth", () => ({
  useAuthStore: vi.fn(),
}));

const mockListDecisionLogsApi = vi.mocked(decisionLogsApi.listDecisionLogsApi);
const mockUseAuthStore = vi.mocked(useAuthStore);

function mockAuthStore(userBizKey: string) {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  mockUseAuthStore.mockImplementation((selector: (...args: any[]) => any) => {
    const state = {
      token: "test-token",
      user: {
        bizKey: userBizKey,
        username: "testuser",
        displayName: "Test User",
        isSuperAdmin: false,
        createTime: "2026-01-01T00:00:00Z",
      },
      isAuthenticated: true,
      isSuperAdmin: false,
      permissions: null,
      permissionsLoadedAt: null,
      _hasHydrated: true,
      setAuth: vi.fn(),
      clearAuth: vi.fn(),
      setPermissions: vi.fn(),
      fetchPermissions: vi.fn(),
      hasPermission: vi.fn().mockReturnValue(true),
      _setHasHydrated: vi.fn(),
    };
    return selector(state);
  });
}

const defaultProps = {
  teamId: "team-bk",
  mainItemId: "main-bk",
  mainStatus: "progressing",
  onAdd: vi.fn(),
  onEdit: vi.fn(),
};

const sampleLogs = [
  {
    bizKey: "log-1",
    mainItemKey: "main-bk",
    category: "technical",
    tags: ["架构", "性能优化"],
    content: "采用微服务架构进行系统拆分，提升系统可维护性和扩展性",
    logStatus: "published" as const,
    createdBy: "user-a",
    creatorName: "张三",
    createTime: "2026-04-27T14:30:00Z",
    updateTime: "2026-04-27T14:30:00Z",
  },
  {
    bizKey: "log-2",
    mainItemKey: "main-bk",
    category: "risk",
    tags: ["安全", "数据泄露"],
    content: "识别到数据传输安全风险，决定引入端到端加密方案",
    logStatus: "draft" as const,
    createdBy: "user-b",
    creatorName: "李四",
    createTime: "2026-04-28T10:00:00Z",
    updateTime: "2026-04-28T10:00:00Z",
  },
];

function makePage(items: typeof sampleLogs, total?: number) {
  return {
    items,
    total: total ?? items.length,
    page: 1,
    size: 20,
  };
}

describe("DecisionTimeline", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockAuthStore("user-b");
    mockListDecisionLogsApi.mockResolvedValue(makePage(sampleLogs));
  });

  afterEach(() => {
    vi.mocked(usePermission).mockReturnValue(true);
  });

  describe("loading state", () => {
    it("shows 3 skeleton rows on initial load", () => {
      mockListDecisionLogsApi.mockReturnValue(new Promise(() => {}));
      render(<DecisionTimeline {...defaultProps} />);
      const skeletons = screen.getAllByTestId("timeline-skeleton");
      expect(skeletons).toHaveLength(3);
    });
  });

  describe("empty state", () => {
    it("shows empty message when no items", async () => {
      mockListDecisionLogsApi.mockResolvedValue(makePage([]));
      render(<DecisionTimeline {...defaultProps} />);
      await waitFor(() => {
        expect(screen.getByText("暂无决策记录")).toBeInTheDocument();
      });
    });

    it("shows add button in empty state when has permission and not terminal", async () => {
      mockListDecisionLogsApi.mockResolvedValue(makePage([]));
      render(<DecisionTimeline {...defaultProps} />);
      await waitFor(() => {
        expect(screen.getAllByText("添加决策").length).toBeGreaterThanOrEqual(
          1,
        );
      });
    });

    it("hides add button in empty state when main item is terminal", async () => {
      mockListDecisionLogsApi.mockResolvedValue(makePage([]));
      render(<DecisionTimeline {...defaultProps} mainStatus="completed" />);
      await waitFor(() => {
        expect(screen.queryByText("添加决策")).not.toBeInTheDocument();
      });
    });
  });

  describe("populated state", () => {
    it("renders timeline items", async () => {
      render(<DecisionTimeline {...defaultProps} />);
      await waitFor(() => {
        expect(
          screen.getByText(
            "采用微服务架构进行系统拆分，提升系统可维护性和扩展性",
          ),
        ).toBeInTheDocument();
      });
    });

    it("renders category badges with correct labels", async () => {
      render(<DecisionTimeline {...defaultProps} />);
      await waitFor(() => {
        expect(screen.getByText("技术")).toBeInTheDocument();
        expect(screen.getByText("风险")).toBeInTheDocument();
      });
    });

    it("renders draft badge for draft items", async () => {
      render(<DecisionTimeline {...defaultProps} />);
      await waitFor(() => {
        expect(screen.getByText("草稿")).toBeInTheDocument();
      });
    });

    it("renders tag badges", async () => {
      render(<DecisionTimeline {...defaultProps} />);
      await waitFor(() => {
        expect(screen.getByText("架构")).toBeInTheDocument();
        expect(screen.getByText("性能优化")).toBeInTheDocument();
      });
    });

    it("renders creator name", async () => {
      render(<DecisionTimeline {...defaultProps} />);
      await waitFor(() => {
        expect(screen.getByText("张三")).toBeInTheDocument();
        expect(screen.getByText("李四")).toBeInTheDocument();
      });
    });

    it("renders time in YYYY-MM-DD HH:mm format", async () => {
      render(<DecisionTimeline {...defaultProps} />);
      await waitFor(() => {
        // createTime "2026-04-27T14:30:00Z" should display as "2026-04-27 14:30" or similar
        expect(screen.getByText(/2026-04-27/)).toBeInTheDocument();
      });
    });

    it("shows header with title and add button", async () => {
      render(<DecisionTimeline {...defaultProps} />);
      await waitFor(() => {
        expect(screen.getByText("决策记录")).toBeInTheDocument();
        expect(screen.getAllByText("添加决策").length).toBeGreaterThanOrEqual(
          1,
        );
      });
    });

    it("hides add button when no permission", async () => {
      vi.mocked(usePermission).mockReturnValue(false);
      render(<DecisionTimeline {...defaultProps} />);
      await waitFor(() => {
        expect(screen.queryByText("添加决策")).not.toBeInTheDocument();
      });
    });

    it("hides add button when main item is terminal", async () => {
      render(<DecisionTimeline {...defaultProps} mainStatus="closed" />);
      await waitFor(() => {
        expect(screen.queryByText("添加决策")).not.toBeInTheDocument();
      });
    });

    it("calls onAdd when add button is clicked", async () => {
      render(<DecisionTimeline {...defaultProps} />);
      await waitFor(() => {
        expect(screen.getAllByText("添加决策").length).toBeGreaterThanOrEqual(
          1,
        );
      });
      fireEvent.click(screen.getAllByText("添加决策")[0]);
      expect(defaultProps.onAdd).toHaveBeenCalledTimes(1);
    });
  });

  describe("edit button", () => {
    it("shows edit button on draft items where createdBy matches currentUser", async () => {
      mockAuthStore("user-b"); // user-b created the draft log-2
      render(<DecisionTimeline {...defaultProps} />);
      await waitFor(() => {
        const editButtons = screen.getAllByText("编辑");
        expect(editButtons).toHaveLength(1);
      });
    });

    it("hides edit button when createdBy does not match currentUser", async () => {
      mockAuthStore("user-c"); // not the creator of any draft
      render(<DecisionTimeline {...defaultProps} />);
      await waitFor(() => {
        expect(screen.queryByText("编辑")).not.toBeInTheDocument();
      });
    });

    it("calls onEdit with bizKey when edit button is clicked", async () => {
      mockAuthStore("user-b");
      render(<DecisionTimeline {...defaultProps} />);
      await waitFor(() => {
        expect(screen.getByText("编辑")).toBeInTheDocument();
      });
      fireEvent.click(screen.getByText("编辑"));
      expect(defaultProps.onEdit).toHaveBeenCalledWith("log-2");
    });
  });

  describe("content expand/collapse", () => {
    it("toggles content expanded state on click", async () => {
      render(<DecisionTimeline {...defaultProps} />);
      await waitFor(() => {
        expect(
          screen.getByText(
            "采用微服务架构进行系统拆分，提升系统可维护性和扩展性",
          ),
        ).toBeInTheDocument();
      });
      const contentEl = screen.getByText(
        "采用微服务架构进行系统拆分，提升系统可维护性和扩展性",
      );
      expect(contentEl.closest("[aria-expanded]")).toHaveAttribute(
        "aria-expanded",
        "false",
      );
      fireEvent.click(contentEl);
      expect(contentEl.closest("[aria-expanded]")).toHaveAttribute(
        "aria-expanded",
        "true",
      );
    });
  });

  describe("footer", () => {
    it("shows loaded count when all items loaded", async () => {
      render(<DecisionTimeline {...defaultProps} />);
      await waitFor(() => {
        expect(screen.getByText("已加载 2 条")).toBeInTheDocument();
      });
    });
  });

  describe("error state", () => {
    it("shows error alert with retry button on first page failure", async () => {
      mockListDecisionLogsApi.mockRejectedValue(new Error("Network error"));
      render(<DecisionTimeline {...defaultProps} />);
      await waitFor(() => {
        expect(screen.getByText("加载失败")).toBeInTheDocument();
        expect(screen.getByText("重试")).toBeInTheDocument();
      });
    });

    it("retries fetching when retry button is clicked", async () => {
      mockListDecisionLogsApi.mockRejectedValueOnce(new Error("Network error"));
      mockListDecisionLogsApi.mockResolvedValue(makePage(sampleLogs));
      render(<DecisionTimeline {...defaultProps} />);
      await waitFor(() => {
        expect(screen.getByText("重试")).toBeInTheDocument();
      });
      fireEvent.click(screen.getByText("重试"));
      await waitFor(() => {
        expect(screen.getByText("技术")).toBeInTheDocument();
      });
    });
  });

  describe("accessibility", () => {
    it("has role=feed on container with aria-label", async () => {
      render(<DecisionTimeline {...defaultProps} />);
      await waitFor(() => {
        const feed = screen.getByRole("feed");
        expect(feed).toHaveAttribute("aria-label", "决策记录");
      });
    });

    it("has role=article on timeline items", async () => {
      render(<DecisionTimeline {...defaultProps} />);
      await waitFor(() => {
        const articles = screen.getAllByRole("article");
        expect(articles).toHaveLength(2);
      });
    });
  });

  describe("tag overflow", () => {
    it("shows +N badge when more than 3 tags", async () => {
      const logWithManyTags = [
        {
          ...sampleLogs[0],
          tags: ["tag1", "tag2", "tag3", "tag4", "tag5"],
        },
      ];
      mockListDecisionLogsApi.mockResolvedValue(makePage(logWithManyTags));
      render(<DecisionTimeline {...defaultProps} />);
      await waitFor(() => {
        expect(screen.getByText("+2")).toBeInTheDocument();
      });
    });
  });

  describe("refreshKey prop", () => {
    it("re-fetches data when refreshKey changes", async () => {
      const { rerender } = render(
        <DecisionTimeline {...defaultProps} refreshKey={0} />,
      );
      await waitFor(() => {
        expect(mockListDecisionLogsApi).toHaveBeenCalledTimes(1);
      });

      // Simulate a new decision being added — parent increments refreshKey
      const newLog = {
        bizKey: "log-3",
        mainItemKey: "main-bk",
        category: "resource" as const,
        tags: [],
        content: "New decision after refresh",
        logStatus: "published" as const,
        createdBy: "user-c",
        creatorName: "王五",
        createTime: "2026-04-29T10:00:00Z",
        updateTime: "2026-04-29T10:00:00Z",
      };
      mockListDecisionLogsApi.mockResolvedValue(
        makePage([...sampleLogs, newLog]),
      );

      rerender(<DecisionTimeline {...defaultProps} refreshKey={1} />);

      await waitFor(() => {
        expect(mockListDecisionLogsApi).toHaveBeenCalledTimes(2);
      });
      await waitFor(() => {
        expect(screen.getByText("New decision after refresh")).toBeInTheDocument();
      });
    });
  });
});
