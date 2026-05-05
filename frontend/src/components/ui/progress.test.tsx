import { describe, it, expect } from "vitest";
import { render } from "@testing-library/react";
import { Progress } from "./progress";

describe("Progress", () => {
  it("renders progress bar", () => {
    const { container } = render(<Progress value={50} />);
    expect(container.firstChild).toBeInTheDocument();
  });

  it("renders with default size", () => {
    const { container } = render(<Progress value={0} data-testid="progress" />);
    expect(container.firstChild).toBeInTheDocument();
  });

  it("renders small size", () => {
    const { container } = render(
      <Progress value={50} size="sm" data-testid="progress" />,
    );
    const root = container.firstChild as HTMLElement;
    expect(root.className).toContain("h-1.5");
  });

  it("renders large size", () => {
    const { container } = render(
      <Progress value={50} size="lg" data-testid="progress" />,
    );
    const root = container.firstChild as HTMLElement;
    expect(root.className).toContain("h-3");
  });

  it("indicator has width based on value", () => {
    const { container } = render(<Progress value={75} />);
    const indicator = container.querySelector("[style]");
    expect(indicator).toHaveStyle({ width: "75%" });
  });

  it("defaults to 0% when value is undefined", () => {
    const { container } = render(<Progress />);
    const indicator = container.querySelector("[style]");
    expect(indicator).toHaveStyle({ width: "0%" });
  });
});
