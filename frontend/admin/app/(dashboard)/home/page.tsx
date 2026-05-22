"use client";

import * as React from "react";
import {
  BarChart3,
  BookOpen,
  Calendar,
  LoaderCircle,
  Users,
} from "lucide-react";
import {
  readBrowserAccessToken,
  clearStoredSession,
  authFetch,
} from "@/lib/auth-session";
import { useRouter } from "next/navigation";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

type Stats = {
  total_exams: number;
  total_users: number;
  active_users_7d: number;
  platform_version: string;
  last_updated: string;
};

const updateRecords = [
  {
    version: "v0.2.0",
    title: "学习与交付能力更新",
    description: "接入 Redis + Go 二级缓存，完善学习首页、交互单元、错题本闪卡和用户画像图谱。",
  },
  {
    version: "v0.1.1",
    title: "题库与交互体验增强",
    description: "补齐题型筛选、错题选项展示、交互单元会话保留和标注题逐字选择。",
  },
  {
    version: "v0.1.0",
    title: "初始版本发布",
    description: "完成管理员登录、题库管理、用户管理、平台概览和基础学习端流程。",
  },
];

export default function HomePage() {
  const router = useRouter();
  const [stats, setStats] = React.useState<Stats | null>(null);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState("");

  React.useEffect(() => {
    let cancelled = false;
    async function load() {
      const token = readBrowserAccessToken();
      if (!token) {
        clearStoredSession();
        router.replace("/");
        return;
      }
      try {
        const resp = await authFetch("/api/v1/admin/stats", {
          headers: { Authorization: `Bearer ${token}` },
          cache: "no-store",
        });
        if (!resp.ok) {
          if (resp.status === 401) {
            clearStoredSession();
            router.replace("/");
            return;
          }
          throw new Error("请求失败");
        }
        const payload = await resp.json();
        if (!cancelled) setStats(payload.data as Stats);
      } catch {
        if (!cancelled) setError("加载统计数据失败");
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    void load();
    return () => {
      cancelled = true;
    };
  }, [router]);

  if (loading) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <LoaderCircle className="size-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <p className="text-destructive">{error}</p>
      </div>
    );
  }

  const cards = [
    {
      title: "考试总数",
      value: stats?.total_exams ?? 0,
      icon: BookOpen,
      color: "text-blue-600",
      bg: "bg-blue-50",
    },
    {
      title: "注册用户",
      value: stats?.total_users ?? 0,
      icon: Users,
      color: "text-emerald-600",
      bg: "bg-emerald-50",
    },
    {
      title: "近7日活跃",
      value: stats?.active_users_7d ?? 0,
      icon: BarChart3,
      color: "text-amber-600",
      bg: "bg-amber-50",
    },
  ];

  return (
    <main className="mx-auto max-w-6xl px-6 py-8">
      <div className="mb-8">
        <h1 className="text-2xl font-bold tracking-tight">平台概览</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          查看 FOCO 平台运行状态和关键数据
        </p>
      </div>

      <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {cards.map((c) => (
          <Card key={c.title} className="py-2">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {c.title}
              </CardTitle>
              <div className={`rounded-lg p-2 ${c.bg}`}>
                <c.icon className={`size-4 ${c.color}`} />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold">{c.value}</div>
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="mt-8 grid gap-6 lg:grid-cols-2">
        <Card className="py-2">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <Calendar className="size-4 text-muted-foreground" />
              平台信息
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">平台版本</span>
              <Badge variant="outline">{stats?.platform_version ?? "-"}</Badge>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">缓存机制</span>
              <span className="font-medium">Go L1 + Redis L2</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">设置存储</span>
              <span className="font-medium">admin_settings</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">最近同步</span>
              <span className="font-medium">
                {stats?.last_updated
                  ? new Date(stats.last_updated).toLocaleString("zh-CN")
                  : "-"}
              </span>
            </div>
          </CardContent>
        </Card>

        <Card className="py-2">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <BookOpen className="size-4 text-muted-foreground" />
              更新记录
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3 text-sm">
              {updateRecords.map((record) => (
                <div key={record.version} className="flex items-start gap-3">
                  <div className="mt-1.5 size-2 rounded-full bg-primary" />
                  <div>
                    <p className="font-medium">
                      {record.version} — {record.title}
                    </p>
                    <p className="text-muted-foreground">
                      {record.description}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </main>
  );
}
