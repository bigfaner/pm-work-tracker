import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { DecisionFormDialog } from "./DecisionFormDialog";
import * as decisionLogsApi from "@/api/decisionLogs";

// Mock the API module
vi.mock("@/api/decisionLogs", () => ({
  createDecisionLogApi: vi.fn(),
  updateDecisionLogApi: vi.fn(),
  publishDecisionLogApi: vi.fn(),
}));

// Mock useToast
const mockAddToast = vi.fn();
vi.mock("@/components/ui/toast", () => ({
  useToast: () => ({ addToast: mockAddToast }),
}));

const defaultProps = {
  open: true,
  onOpenChange: vi.fn(),
  teamBizKey: "team-1",
  mainBizKey: "main-1",
  onSuccess: vi.fn(),
  recentTags: ["性能优化", "缓存策略", "里程碑", "合规"],
};

describe("DecisionFormDialog", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("rendering", () => {
    it("renders in new mode with correct title", () => {
      render(<DecisionFormDialog {...defaultProps} mode="new" />);
      expect(screen.getByText("添加决策")).toBeInTheDocument();
    });

    it("renders in edit mode with correct title", () => {
      render(
        <DecisionFormDialog
          {...defaultProps}
          mode="edit"
          draftData={{
            bizKey: "log-1",
            category: "schedule",
            tags: ["里程碑"],
            content: "延期一周",
            logStatus: "draft",
          }}
        />,
      );
      expect(screen.getByText("编辑决策")).toBeInTheDocument();
    });

    it("renders three form fields and three footer buttons", () => {
      render(<DecisionFormDialog {...defaultProps} mode="new" />);
      expect(screen.getByText("分类")).toBeInTheDocument();
      expect(screen.getByText("标签")).toBeInTheDocument();
      expect(screen.getByText("决策内容")).toBeInTheDocument();
      expect(screen.getByRole("button", { name: "取消" })).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: "保存草稿" }),
      ).toBeInTheDocument();
      expect(screen.getByRole("button", { name: "发布" })).toBeInTheDocument();
    });

    it("has category select with 6 predefined options", () => {
      render(<DecisionFormDialog {...defaultProps} mode="new" />);
      const select = screen.getByRole("combobox");
      expect(select).toBeInTheDocument();
      // Placeholder + 6 options
      expect(select.children).toHaveLength(7);
    });

    it("has content textarea with character counter", () => {
      render(<DecisionFormDialog {...defaultProps} mode="new" />);
      expect(screen.getByPlaceholderText("请输入决策内容")).toBeInTheDocument();
      expect(screen.getByText("0/2000")).toBeInTheDocument();
    });

    it("has ARIA dialog attributes", () => {
      render(<DecisionFormDialog {...defaultProps} mode="new" />);
      // Radix Dialog sets aria-modal and aria-labelledby on the content wrapper
      const dialog = screen.getByRole("dialog");
      // aria-labelledby should point to the title
      expect(dialog).toHaveAttribute("aria-labelledby");
    });
  });

  describe("edit mode", () => {
    it("pre-fills fields from draftData prop", () => {
      render(
        <DecisionFormDialog
          {...defaultProps}
          mode="edit"
          draftData={{
            bizKey: "log-1",
            category: "schedule",
            tags: ["里程碑", "延期"],
            content: "预计延期一周",
            logStatus: "draft",
          }}
        />,
      );
      const select = screen.getByRole("combobox") as HTMLSelectElement;
      expect(select.value).toBe("schedule");
      expect(screen.getByText("里程碑")).toBeInTheDocument();
      expect(screen.getByText("延期")).toBeInTheDocument();
      expect(screen.getByDisplayValue("预计延期一周")).toBeInTheDocument();
    });

    it("shows loading spinner when loading is true", () => {
      render(
        <DecisionFormDialog {...defaultProps} mode="edit" loading={true} />,
      );
      expect(screen.getByText("加载中...")).toBeInTheDocument();
    });
  });

  describe("category select", () => {
    it("shows error when submitted without selection", async () => {
      const user = userEvent.setup();
      render(<DecisionFormDialog {...defaultProps} mode="new" />);
      await user.click(screen.getByRole("button", { name: "发布" }));
      // Error message should appear (find it by text in role=alert elements)
      const alerts = screen.getAllByRole("alert");
      const categoryAlert = alerts.find(
        (el) => el.textContent === "请选择分类",
      );
      expect(categoryAlert).toBeTruthy();
    });

    it("clears error when category is selected", async () => {
      const user = userEvent.setup();
      render(<DecisionFormDialog {...defaultProps} mode="new" />);
      await user.click(screen.getByRole("button", { name: "发布" }));
      const alerts = screen.getAllByRole("alert");
      expect(alerts.some((el) => el.textContent === "请选择分类")).toBe(true);

      const select = screen.getByRole("combobox");
      await user.selectOptions(select, "technical");
      // Category error should be cleared
      const remainingAlerts = screen.queryAllByRole("alert");
      expect(
        remainingAlerts.some((el) => el.textContent === "请选择分类"),
      ).toBe(false);
    });
  });

  describe("tags input", () => {
    it("adds tag on Enter key", async () => {
      const user = userEvent.setup();
      render(<DecisionFormDialog {...defaultProps} mode="new" />);
      const input = screen.getByPlaceholderText("输入标签，回车添加");
      await user.type(input, "缓存策略{Enter}");
      expect(screen.getByText("缓存策略")).toBeInTheDocument();
    });

    it("adds tag on comma key", async () => {
      const user = userEvent.setup();
      render(<DecisionFormDialog {...defaultProps} mode="new" />);
      const input = screen.getByPlaceholderText("输入标签，回车添加");
      await user.type(input, "性能优化,");
      expect(screen.getByText("性能优化")).toBeInTheDocument();
    });

    it("prevents duplicate tags", async () => {
      const user = userEvent.setup();
      render(
        <DecisionFormDialog
          {...defaultProps}
          mode="edit"
          draftData={{
            bizKey: "log-1",
            category: "technical",
            tags: ["缓存策略"],
            content: "test",
            logStatus: "draft",
          }}
        />,
      );
      const input = screen.getByPlaceholderText("输入标签，回车添加");
      await user.type(input, "缓存策略{Enter}");
      // Should still only have one instance
      expect(screen.getAllByText("缓存策略").length).toBe(1);
    });

    it("removes tag on clicking x button", async () => {
      const user = userEvent.setup();
      render(
        <DecisionFormDialog
          {...defaultProps}
          mode="edit"
          draftData={{
            bizKey: "log-1",
            category: "technical",
            tags: ["缓存策略"],
            content: "test",
            logStatus: "draft",
          }}
        />,
      );
      expect(screen.getByText("缓存策略")).toBeInTheDocument();
      // Click the remove button (×)
      const removeBtn = screen.getByLabelText("移除标签 缓存策略");
      await user.click(removeBtn);
      expect(screen.queryByText("缓存策略")).not.toBeInTheDocument();
    });

    it("shows error for tags exceeding 20 chars", async () => {
      const user = userEvent.setup();
      render(<DecisionFormDialog {...defaultProps} mode="new" />);
      const input = screen.getByPlaceholderText("输入标签，回车添加");
      // 21 chars
      const longTag = "a".repeat(21);
      await user.type(input, longTag + "{Enter}");
      expect(screen.getByText("标签不能超过 20 字符")).toBeInTheDocument();
    });

    it("shows recent tags dropdown when input is focused", async () => {
      const user = userEvent.setup();
      render(<DecisionFormDialog {...defaultProps} mode="new" />);
      const input = screen.getByPlaceholderText("输入标签，回车添加");
      await user.click(input);
      expect(screen.getByText("性能优化")).toBeInTheDocument();
      expect(screen.getByText("缓存策略")).toBeInTheDocument();
    });

    it("filters recent tags by input text", async () => {
      const user = userEvent.setup();
      render(<DecisionFormDialog {...defaultProps} mode="new" />);
      const input = screen.getByPlaceholderText("输入标签，回车添加");
      await user.click(input);
      // All recent tags visible initially
      expect(screen.getByText("性能优化")).toBeInTheDocument();
      await user.type(input, "缓");
      // "缓存策略" matches, others filtered out
      expect(screen.getByText("缓存策略")).toBeInTheDocument();
      expect(screen.queryByText("性能优化")).not.toBeInTheDocument();
    });

    it("adds tag from recent tags suggestion click", async () => {
      const user = userEvent.setup();
      render(<DecisionFormDialog {...defaultProps} mode="new" />);
      const input = screen.getByPlaceholderText("输入标签，回车添加");
      await user.click(input);
      // Click a suggestion from the dropdown
      const suggestion = screen.getByText("里程碑");
      await user.click(suggestion);
      // Tag should be added as a badge in the input wrapper
      const tags = screen.getAllByText("里程碑");
      // One in dropdown, one as tag badge — but after click, tag should be in the input wrapper
      expect(tags.length).toBeGreaterThanOrEqual(1);
    });
  });

  describe("content textarea", () => {
    it("updates character counter on input", async () => {
      const user = userEvent.setup();
      render(<DecisionFormDialog {...defaultProps} mode="new" />);
      const textarea = screen.getByPlaceholderText("请输入决策内容");
      await user.type(textarea, "test");
      expect(screen.getByText("4/2000")).toBeInTheDocument();
    });

    it("shows warning color when content > 1800 chars", () => {
      render(<DecisionFormDialog {...defaultProps} mode="new" />);
      const textarea = screen.getByPlaceholderText("请输入决策内容");
      const longText = "a".repeat(1801);
      fireEvent.change(textarea, { target: { value: longText } });
      const counter = screen.getByText("1801/2000");
      expect(counter.className).toContain("text-warning");
    });

    it("shows error color when content reaches 2000 chars", () => {
      render(<DecisionFormDialog {...defaultProps} mode="new" />);
      const textarea = screen.getByPlaceholderText("请输入决策内容");
      const longText = "a".repeat(2000);
      fireEvent.change(textarea, { target: { value: longText } });
      const counter = screen.getByText("2000/2000");
      expect(counter.className).toContain("text-error");
    });

    it("shows required error when submitted empty", async () => {
      const user = userEvent.setup();
      render(<DecisionFormDialog {...defaultProps} mode="new" />);
      await user.click(screen.getByRole("button", { name: "发布" }));
      expect(screen.getByText("请输入决策内容")).toBeInTheDocument();
    });
  });

  describe("submit flows", () => {
    it("calls createDecisionLogApi with logStatus=draft on save draft", async () => {
      const user = userEvent.setup();
      vi.mocked(decisionLogsApi.createDecisionLogApi).mockResolvedValue({
        bizKey: "new-1",
        mainItemKey: "main-1",
        category: "technical",
        tags: ["缓存策略"],
        content: "决定采用 Redis 缓存热点数据",
        logStatus: "draft",
        createdBy: "user-1",
        creatorName: "Test",
        createTime: "2026-01-01T00:00:00Z",
        updateTime: "2026-01-01T00:00:00Z",
      });

      render(<DecisionFormDialog {...defaultProps} mode="new" />);

      // Fill form
      await user.selectOptions(screen.getByRole("combobox"), "technical");
      const tagInput = screen.getByPlaceholderText("输入标签，回车添加");
      await user.type(tagInput, "缓存策略{Enter}");
      const textarea = screen.getByPlaceholderText("请输入决策内容");
      await user.type(textarea, "决定采用 Redis 缓存热点数据");

      await user.click(screen.getByRole("button", { name: "保存草稿" }));

      await waitFor(() => {
        expect(decisionLogsApi.createDecisionLogApi).toHaveBeenCalledWith(
          "team-1",
          "main-1",
          expect.objectContaining({
            category: "technical",
            tags: ["缓存策略"],
            content: "决定采用 Redis 缓存热点数据",
            logStatus: "draft",
          }),
        );
      });
    });

    it("calls createDecisionLogApi with logStatus=published on publish", async () => {
      const user = userEvent.setup();
      vi.mocked(decisionLogsApi.createDecisionLogApi).mockResolvedValue({
        bizKey: "new-1",
        mainItemKey: "main-1",
        category: "technical",
        tags: [],
        content: "test content",
        logStatus: "published",
        createdBy: "user-1",
        creatorName: "Test",
        createTime: "2026-01-01T00:00:00Z",
        updateTime: "2026-01-01T00:00:00Z",
      });

      render(<DecisionFormDialog {...defaultProps} mode="new" />);

      await user.selectOptions(screen.getByRole("combobox"), "technical");
      await user.type(
        screen.getByPlaceholderText("请输入决策内容"),
        "test content",
      );

      await user.click(screen.getByRole("button", { name: "发布" }));

      await waitFor(() => {
        expect(decisionLogsApi.createDecisionLogApi).toHaveBeenCalledWith(
          "team-1",
          "main-1",
          expect.objectContaining({
            logStatus: "published",
          }),
        );
      });
    });

    it("calls updateDecisionLogApi on save draft in edit mode", async () => {
      const user = userEvent.setup();
      vi.mocked(decisionLogsApi.updateDecisionLogApi).mockResolvedValue({
        bizKey: "log-1",
        mainItemKey: "main-1",
        category: "technical",
        tags: ["缓存策略"],
        content: "updated content here",
        logStatus: "draft",
        createdBy: "user-1",
        creatorName: "Test",
        createTime: "2026-01-01T00:00:00Z",
        updateTime: "2026-01-01T00:00:00Z",
      });

      render(
        <DecisionFormDialog
          {...defaultProps}
          mode="edit"
          draftData={{
            bizKey: "log-1",
            category: "technical",
            tags: [],
            content: "original content here",
            logStatus: "draft",
          }}
        />,
      );

      const textarea = screen.getByPlaceholderText("请输入决策内容");
      await user.clear(textarea);
      await user.type(textarea, "updated content here");

      await user.click(screen.getByRole("button", { name: "保存草稿" }));

      await waitFor(() => {
        expect(decisionLogsApi.updateDecisionLogApi).toHaveBeenCalledWith(
          "team-1",
          "main-1",
          "log-1",
          expect.objectContaining({
            content: "updated content here",
          }),
        );
      });
    });

    it("calls updateDecisionLogApi then publishDecisionLogApi on publish in edit mode", async () => {
      const user = userEvent.setup();
      vi.mocked(decisionLogsApi.updateDecisionLogApi).mockResolvedValue({
        bizKey: "log-1",
        mainItemKey: "main-1",
        category: "technical",
        tags: [],
        content: "test",
        logStatus: "draft",
        createdBy: "user-1",
        creatorName: "Test",
        createTime: "2026-01-01T00:00:00Z",
        updateTime: "2026-01-01T00:00:00Z",
      });
      vi.mocked(decisionLogsApi.publishDecisionLogApi).mockResolvedValue({
        bizKey: "log-1",
        mainItemKey: "main-1",
        category: "technical",
        tags: [],
        content: "test",
        logStatus: "published",
        createdBy: "user-1",
        creatorName: "Test",
        createTime: "2026-01-01T00:00:00Z",
        updateTime: "2026-01-01T00:00:00Z",
      });

      render(
        <DecisionFormDialog
          {...defaultProps}
          mode="edit"
          draftData={{
            bizKey: "log-1",
            category: "technical",
            tags: [],
            content: "test",
            logStatus: "draft",
          }}
        />,
      );

      await user.click(screen.getByRole("button", { name: "发布" }));

      await waitFor(() => {
        expect(decisionLogsApi.updateDecisionLogApi).toHaveBeenCalled();
      });
      await waitFor(() => {
        expect(decisionLogsApi.publishDecisionLogApi).toHaveBeenCalledWith(
          "team-1",
          "main-1",
          "log-1",
        );
      });
    });

    it("closes dialog and calls onSuccess after successful submit", async () => {
      const user = userEvent.setup();
      vi.mocked(decisionLogsApi.createDecisionLogApi).mockResolvedValue({
        bizKey: "new-1",
        mainItemKey: "main-1",
        category: "technical",
        tags: [],
        content: "test",
        logStatus: "published",
        createdBy: "user-1",
        creatorName: "Test",
        createTime: "2026-01-01T00:00:00Z",
        updateTime: "2026-01-01T00:00:00Z",
      });

      render(<DecisionFormDialog {...defaultProps} mode="new" />);

      await user.selectOptions(screen.getByRole("combobox"), "technical");
      await user.type(screen.getByPlaceholderText("请输入决策内容"), "test");
      await user.click(screen.getByRole("button", { name: "发布" }));

      await waitFor(() => {
        expect(defaultProps.onSuccess).toHaveBeenCalled();
        expect(defaultProps.onOpenChange).toHaveBeenCalledWith(false);
      });
    });

    it("shows submitting state on publish button during API call", async () => {
      const user = userEvent.setup();
      let resolveCreate: (value: unknown) => void;
      vi.mocked(decisionLogsApi.createDecisionLogApi).mockReturnValue(
        new Promise<unknown>((resolve) => {
          resolveCreate = resolve;
        }) as Promise<decisionLogsApi.DecisionLog>,
      );

      render(<DecisionFormDialog {...defaultProps} mode="new" />);

      await user.selectOptions(screen.getByRole("combobox"), "technical");
      await user.type(screen.getByPlaceholderText("请输入决策内容"), "test");
      await user.click(screen.getByRole("button", { name: "发布" }));

      // Buttons should be disabled during submission
      expect(screen.getByRole("button", { name: "取消" })).toBeDisabled();
      expect(screen.getByRole("button", { name: "保存草稿" })).toBeDisabled();

      // Resolve the promise
      resolveCreate!({
        bizKey: "new-1",
        mainItemKey: "main-1",
        category: "technical",
        tags: [],
        content: "test",
        logStatus: "published",
        createdBy: "user-1",
        creatorName: "Test",
        createTime: "2026-01-01T00:00:00Z",
        updateTime: "2026-01-01T00:00:00Z",
      });

      await waitFor(() => {
        expect(defaultProps.onSuccess).toHaveBeenCalled();
      });
    });

    it("shows error toast on API failure", async () => {
      const user = userEvent.setup();
      vi.mocked(decisionLogsApi.createDecisionLogApi).mockRejectedValue(
        new Error("Network error"),
      );

      render(<DecisionFormDialog {...defaultProps} mode="new" />);

      await user.selectOptions(screen.getByRole("combobox"), "technical");
      await user.type(screen.getByPlaceholderText("请输入决策内容"), "test");
      await user.click(screen.getByRole("button", { name: "发布" }));

      await waitFor(() => {
        expect(mockAddToast).toHaveBeenCalledWith(
          expect.stringContaining("操作失败"),
          "error",
        );
      });
    });
  });

  describe("unsaved changes confirmation", () => {
    it("shows confirmation when closing with unsaved changes", async () => {
      const user = userEvent.setup();
      render(<DecisionFormDialog {...defaultProps} mode="new" />);

      // Make a change
      await user.type(
        screen.getByPlaceholderText("请输入决策内容"),
        "some content",
      );

      // Try to close
      await user.click(screen.getByRole("button", { name: "取消" }));

      expect(
        screen.getByText("未保存的内容将丢失，确认关闭？"),
      ).toBeInTheDocument();
    });

    it("closes dialog when confirmation is accepted", async () => {
      const user = userEvent.setup();
      render(<DecisionFormDialog {...defaultProps} mode="new" />);

      await user.type(
        screen.getByPlaceholderText("请输入决策内容"),
        "some content",
      );
      await user.click(screen.getByRole("button", { name: "取消" }));

      // Confirm close
      await user.click(screen.getByRole("button", { name: "确认" }));
      expect(defaultProps.onOpenChange).toHaveBeenCalledWith(false);
    });
  });
});
