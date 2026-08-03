import { ImageResponse } from "next/og";

export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default function OpengraphImage() {
  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          alignItems: "center",
          backgroundColor: "#0E2640",
          color: "#ffffff",
          fontFamily: "sans-serif",
        }}
      >
        <div
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            width: 88,
            height: 88,
            borderRadius: "50%",
            border: "4px solid #0EEBFD",
            marginBottom: 32,
            fontSize: 40,
            fontWeight: 700,
            color: "#0EEBFD",
          }}
        >
          L
        </div>
        <div style={{ fontSize: 64, fontWeight: 700, letterSpacing: -1 }}>LastcoinVision</div>
        <div style={{ fontSize: 28, color: "#7DABD5", marginTop: 16 }}>
          The Layer-1 for Governments and the Developers Who Serve Them
        </div>
      </div>
    ),
    { ...size },
  );
}
