import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import {
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogBody,
  DialogFooter,
} from "./dialog";
import { Button } from "./button";

function DialogWrapper({ size }: { size?: "sm" | "md" | "lg" }) {
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button>Open</Button>
      </DialogTrigger>
      <DialogContent size={size}>
        <DialogHeader>
          <DialogTitle>Dialog Title</DialogTitle>
          <DialogDescription>Description text</DialogDescription>
        </DialogHeader>
        <DialogBody>Body content</DialogBody>
        <DialogFooter>
          <Button variant="secondary">Cancel</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

describe("Dialog", () => {
  it("renders trigger button", () => {
    render(<DialogWrapper />);
    expect(screen.getByRole("button", { name: "Open" })).toBeInTheDocument();
  });

  it("opens dialog on trigger click", async () => {
    render(<DialogWrapper />);
    await userEvent.click(screen.getByRole("button", { name: "Open" }));
    expect(screen.getByText("Dialog Title")).toBeInTheDocument();
    expect(screen.getByText("Body content")).toBeInTheDocument();
  });

  it("renders sm size", async () => {
    render(<DialogWrapper size="sm" />);
    await userEvent.click(screen.getByRole("button"));
    const content =
      screen.getByText("Dialog Title").closest('[role="dialog"]') ||
      screen.getByText("Dialog Title").parentElement?.parentElement;
    expect(content?.className).toContain("max-w-[480px]");
  });

  it("renders lg size", async () => {
    render(<DialogWrapper size="lg" />);
    await userEvent.click(screen.getByRole("button"));
    const content =
      screen.getByText("Dialog Title").closest('[role="dialog"]') ||
      screen.getByText("Dialog Title").parentElement?.parentElement;
    expect(content?.className).toContain("max-w-[680px]");
  });

  it("has close button", async () => {
    render(<DialogWrapper />);
    await userEvent.click(screen.getByRole("button", { name: "Open" }));
    expect(screen.getByText("Close")).toBeInTheDocument();
  });
});
