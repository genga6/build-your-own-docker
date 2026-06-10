// 画面本体のコンポーネント。今は静的な見出しだけの最小構成。
// 関数が JSX（HTMLのような記法）を return すると、それが画面になる。
export function App() {
  return (
    <main>
      <h1>react-go-lab</h1>
      <p>frontend (React + Vite + TypeScript) の雛形です。</p>
      {/* TODO: backend 完成後、ここで /api/health を fetch して
          結果を useState / useEffect で表示する練習をする */}
    </main>
  );
}
