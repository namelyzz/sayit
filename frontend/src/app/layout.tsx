import type { Metadata } from "next";
import "./globals.css";
import Layout from "@/components/layout/Layout";
import { AuthProvider } from "@/context/AuthContext";

export const metadata: Metadata = {
  title: "SayIt - 晒意社区",
  description: "一个清爽、现代的中文社区讨论平台。",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="zh-CN" className="h-full antialiased">
      <body className="min-h-full">
        <AuthProvider>
          <Layout>{children}</Layout>
        </AuthProvider>
      </body>
    </html>
  );
}
