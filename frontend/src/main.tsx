import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "@/App";

// index.html の <div id="root"> を取得する。
const rootElement = document.getElementById("root");

// 万一 root が無ければ、原因不明の動作になる前にここで明示的に止める。
// TypeScript 的にも null チェックを通すことで、この先 rootElement を安全に使える。
if (!rootElement) {
  throw new Error("#root が見つかりません");
}

// React 18+ の新方式。root を作り、その中に <App /> を描画する。
createRoot(rootElement).render(
  // StrictMode は開発時だけ働く「お行儀チェック係」。
  // 潜在的な問題を見つけるため、一部処理をわざと2回実行する（本番では無効）。
  <StrictMode>
    <App />
  </StrictMode>,
);
