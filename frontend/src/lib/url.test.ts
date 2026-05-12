import { describe, it, expect, vi, beforeEach } from "vitest";
import { buildFullUrl } from "./url";

describe("buildFullUrl", () => {
  beforeEach(() => {
    vi.stubGlobal("window", {
      ...window,
      location: {
        ...window.location,
        origin: "https://example.com",
      },
    });
  });

  it("returns URL without base path when not provided", () => {
    expect(buildFullUrl("/items/1")).toBe("https://example.com/items/1");
  });

  it("bug: includes base path when deployed under sub-path", () => {
    expect(buildFullUrl("/items/1", "/tracker")).toBe(
      "https://example.com/tracker/items/1",
    );
  });

  it("handles sub-item paths with base path", () => {
    expect(buildFullUrl("/items/1/sub/11", "/pm-tracker")).toBe(
      "https://example.com/pm-tracker/items/1/sub/11",
    );
  });
});
