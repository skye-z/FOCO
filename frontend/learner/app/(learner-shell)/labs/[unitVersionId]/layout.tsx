"use client"

import * as React from "react"

export default function LabsDetailLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <div className="fixed inset-0 z-40 overflow-auto bg-[var(--surface)]">
      {children}
    </div>
  )
}
