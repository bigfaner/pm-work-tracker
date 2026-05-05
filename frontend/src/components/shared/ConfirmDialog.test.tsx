import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import ConfirmDialog from "./ConfirmDialog";

describe("ConfirmDialog", () => {
  it("renders title and description when open", () => {
    render(
      <ConfirmDialog
        open={true}
        onOpenChange={() => {}}
        title="确认操作"
        description="确定要执行此操作吗？"
        onConfirm={() => {}}
      />,
    );
    expect(screen.getByText("确认操作")).toBeInTheDocument();
    expect(screen.getByText("确定要执行此操作吗？")).toBeInTheDocument();
  });

  it("does not render when closed", () => {
    render(
      <ConfirmDialog
        open={false}
        onOpenChange={() => {}}
        title="确认操作"
        onConfirm={() => {}}
      />,
    );
    expect(screen.queryByText("确认操作")).not.toBeInTheDocument();
  });

  it("calls onConfirm when confirm button is clicked", async () => {
    const onConfirm = vi.fn();
    const user = userEvent.setup();
    render(
      <ConfirmDialog
        open={true}
        onOpenChange={() => {}}
        title="确认操作"
        onConfirm={onConfirm}
      />,
    );
    await user.click(screen.getByRole("button", { name: "确认" }));
    expect(onConfirm).toHaveBeenCalledOnce();
  });

  it("calls onOpenChange(false) when cancel is clicked", async () => {
    const onOpenChange = vi.fn();
    const user = userEvent.setup();
    render(
      <ConfirmDialog
        open={true}
        onOpenChange={onOpenChange}
        title="确认操作"
        onConfirm={() => {}}
      />,
    );
    await user.click(screen.getByRole("button", { name: "取消" }));
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("renders text input when placeholder is provided", () => {
    render(
      <ConfirmDialog
        open={true}
        onOpenChange={() => {}}
        title="解散团队"
        description="请输入团队名称确认"
        confirmPlaceholder="请输入团队名称"
        onConfirm={() => {}}
      />,
    );
    expect(screen.getByPlaceholderText("请输入团队名称")).toBeInTheDocument();
  });

  it("does not render text input when placeholder is omitted", () => {
    render(
      <ConfirmDialog
        open={true}
        onOpenChange={() => {}}
        title="确认操作"
        onConfirm={() => {}}
      />,
    );
    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
  });

  it("uses custom confirm label", () => {
    render(
      <ConfirmDialog
        open={true}
        onOpenChange={() => {}}
        title="删除"
        confirmLabel="删除"
        onConfirm={() => {}}
      />,
    );
    expect(screen.getByRole("button", { name: "删除" })).toBeInTheDocument();
  });

  it("shows destructive variant when confirmVariant is danger", () => {
    render(
      <ConfirmDialog
        open={true}
        onOpenChange={() => {}}
        title="删除"
        confirmLabel="删除"
        confirmVariant="danger"
        onConfirm={() => {}}
      />,
    );
    const btn = screen.getByRole("button", { name: "删除" });
    // danger variant applies text-error class
    expect(btn.className).toContain("text-error");
  });
});
