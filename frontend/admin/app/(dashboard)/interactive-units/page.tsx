"use client";

import * as React from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";

export default function InteractiveUnitsPage() {
  const router = useRouter();

  return (
    <div className="mx-auto flex max-w-3xl flex-col gap-4 p-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">交互单元</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          当前入口需要具体单元 ID 才能进入编辑页。请从题库管理列表中选择一个交互单元打开。
        </p>
      </div>
      <div className="rounded-xl border bg-card p-5">
        <p className="text-sm text-muted-foreground">
          该页面用于兜底处理 `/interactive-units` 根路由，避免直接访问时出现 404。
        </p>
        <div className="mt-4">
          <Button onClick={() => router.push("/exams")}>返回题库管理</Button>
        </div>
      </div>
    </div>
  );
}
