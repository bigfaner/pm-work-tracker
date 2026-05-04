import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { Breadcrumb, BreadcrumbItem, BreadcrumbSeparator } from "./breadcrumb";

describe("Breadcrumb", () => {
  it("renders breadcrumb with items and separators", () => {
    render(
      <MemoryRouter>
        <Breadcrumb>
          <BreadcrumbItem href="/">Home</BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem href="/items">Items</BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem isCurrent>Detail</BreadcrumbItem>
        </Breadcrumb>
      </MemoryRouter>,
    );
    expect(screen.getByText("Home")).toBeInTheDocument();
    expect(screen.getByText("Items")).toBeInTheDocument();
    expect(screen.getByText("Detail")).toBeInTheDocument();
  });

  it("renders link items as anchors", () => {
    render(
      <MemoryRouter>
        <Breadcrumb>
          <BreadcrumbItem href="/">Home</BreadcrumbItem>
        </Breadcrumb>
      </MemoryRouter>,
    );
    expect(screen.getByText("Home").tagName).toBe("A");
  });

  it("renders current item as span with text-primary", () => {
    render(
      <MemoryRouter>
        <Breadcrumb>
          <BreadcrumbItem isCurrent>Current</BreadcrumbItem>
        </Breadcrumb>
      </MemoryRouter>,
    );
    const current = screen.getByText("Current");
    expect(current.tagName).toBe("SPAN");
    expect(current.className).toContain("text-primary");
  });

  it("separator renders slash", () => {
    render(<BreadcrumbSeparator />);
    expect(screen.getByText("/")).toBeInTheDocument();
  });
});
