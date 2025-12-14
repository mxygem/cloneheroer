import "./globals.css";
import type { Metadata } from "next";
import { ReactNode } from "react";

export const metadata: Metadata = {
  title: "Cloneheroer Scores",
  description: "Visualize Clone Hero scores"
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body>
        <header className="app-header">
          <div className="container">
            <h1>Cloneheroer</h1>
            <p>Score tracker dashboard</p>
          </div>
        </header>
        <main className="container">{children}</main>
      </body>
    </html>
  );
}

