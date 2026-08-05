import { describe, expect, it } from "vitest";

import { accent, navy } from "../src/colors.js";

// Regression test against web/tailwind.config.ts's literal values (2026-08-06).
// If that file's palette ever changes, this test should fail until this
// package's colors.ts is updated to match by hand — see the comment in
// colors.ts on why these aren't generated from one another.
describe("brand color parity with web/tailwind.config.ts", () => {
  it("navy scale matches", () => {
    expect(navy).toEqual({
      50: "#EAF1F8",
      100: "#D4E3F1",
      200: "#A9C7E3",
      300: "#7DABD5",
      400: "#528FC7",
      500: "#1F4E79",
      600: "#1A4166",
      700: "#153453",
      800: "#102740",
      900: "#0E2640",
      950: "#081627",
    });
  });

  it("accent scale matches", () => {
    expect(accent).toEqual({
      50: "#E6FDFF",
      100: "#CCFBFE",
      200: "#99F7FD",
      300: "#66F2FC",
      400: "#33EEFB",
      500: "#0EEBFD",
      600: "#0BC2D1",
      700: "#0899A5",
      800: "#067179",
      900: "#03484C",
    });
  });
});
