import type { Metadata } from "next";
import "./globals.css";
import { Toaster } from "@/components/ui/sonner";

export const metadata: Metadata = {
  title: "FOCO Learner",
  description: "Learner experience for the intelligent exam prep platform",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="zh-CN" className="font-sans">
      <body>
        {children}
        <Toaster position="top-right" />
      </body>
    </html>
  );
}
