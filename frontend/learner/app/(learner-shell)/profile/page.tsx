"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import Link from "next/link";
import cytoscape from "cytoscape";
import {
  Award,
  BarChart3,
  BookOpen,
  BrainCircuit,
  CalendarDays,
  Compass,
  Coins,
  Flame,
  Sparkles,
  Target,
  TrendingUp,
  UserRound,
} from "lucide-react";

import {
  authFetch,
  readBrowserAccessToken,
  readStoredSession,
  type ActiveEnrollmentSnapshot,
} from "@/lib/auth-session";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";

const API_BASE = "/api/v1";

type MePayload = {
  user: {
    id: string;
    email: string;
    roles: string[];
  };
  active_exam_enrollment: ActiveEnrollmentSnapshot;
};

type ProfilePayload = {
  archive: {
    total_sessions: number;
    total_answered: number;
    total_correct: number;
    accuracy: number;
    total_xp: number;
    coins_balance: number;
    current_streak: number;
    best_streak: number;
    streak_status: string;
    first_studied_at?: string;
    last_studied_at?: string;
  };
  records: {
    monthly_sessions: number;
    monthly_answered: number;
    monthly_xp: number;
    recent_sessions: Array<{
      id: string;
      status: string;
      total_count: number;
      answered_count: number;
      correct_count: number;
      accuracy: number;
      xp_earned: number;
      coins_earned: number;
      duration_minutes: number;
      created_at: string;
      completed_at?: string;
    }>;
  };
  portrait: {
    total_attempts: number;
    overall_accuracy: number;
    weak_points: Array<{
      knowledge_point_id: string;
      knowledge_point_name: string;
      attempts: number;
      correct_count: number;
      accuracy: number;
    }>;
    strong_points: Array<{
      knowledge_point_id: string;
      knowledge_point_name: string;
      attempts: number;
      correct_count: number;
      accuracy: number;
    }>;
    knowledge_graph: {
      nodes: Array<{
        id: string;
        type: string;
        ref_id: string;
        label: string;
        description?: string;
        group?: string;
        mastery?: number;
        attempts: number;
      }>;
      edges: Array<{
        id: string;
        source: string;
        target: string;
        type: string;
        label?: string;
      }>;
    };
  };
};

function authHeaders(): Record<string, string> {
  const token = readBrowserAccessToken();
  const headers: Record<string, string> = {};
  if (token) headers.Authorization = `Bearer ${token}`;
  return headers;
}

function formatDate(dateLike?: string) {
  if (!dateLike) return "暂无";
  const date = new Date(dateLike);
  if (Number.isNaN(date.getTime())) return "暂无";
  return date.toLocaleDateString("zh-CN", {
    year: "numeric",
    month: "numeric",
    day: "numeric",
  });
}

function sessionStatusLabel(status: string) {
  switch (status) {
    case "completed":
      return "已完成";
    case "in_progress":
      return "进行中";
    case "active":
      return "进行中";
    default:
      return status;
  }
}

function sessionStatusTone(status: string) {
  switch (status) {
    case "completed":
      return "bg-[var(--secondary)]/10 text-[var(--secondary)]";
    case "in_progress":
    case "active":
      return "bg-[var(--primary)]/10 text-[var(--primary)]";
    default:
      return "bg-[var(--surface-container)] text-[var(--on-surface-variant)]";
  }
}

function PointList({
  title,
  points,
  emptyCopy,
  tone,
}: {
  title: string;
  points: ProfilePayload["portrait"]["weak_points"];
  emptyCopy: string;
  tone: "weak" | "strong";
}) {
  return (
    <Card className="rounded-2xl border-0 bg-[var(--surface-container-lowest)] shadow-sm">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          {tone === "weak" ? (
            <Target className="size-5 text-[var(--secondary)]" />
          ) : (
            <TrendingUp className="size-5 text-[var(--secondary)]" />
          )}
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {points.length === 0 ? (
          <div className="rounded-2xl border border-dashed border-[var(--outline-variant)] bg-[var(--surface)] px-5 py-8 text-center text-sm text-[var(--text-muted)]">
            {emptyCopy}
          </div>
        ) : (
          points.map((point) => (
            <div key={point.knowledge_point_id} className="space-y-2">
              <div className="flex items-center justify-between gap-3 text-sm">
                <span className="font-medium text-[var(--text-main)]">
                  {point.knowledge_point_name}
                </span>
                <span className="text-[var(--text-muted)]">
                  {point.correct_count}/{point.attempts}
                </span>
              </div>
              <Progress
                value={point.accuracy}
                className="h-2 bg-[var(--surface-container-high)] [&>[data-slot=progress-indicator]]:bg-[var(--secondary)]"
              />
              <div className="flex items-center justify-between text-xs text-[var(--text-muted)]">
                <span>正确率 {point.accuracy}%</span>
                <span>练习 {point.attempts} 次</span>
              </div>
            </div>
          ))
        )}
      </CardContent>
    </Card>
  );
}

function LearnerKnowledgeGraph({
  graph,
}: {
  graph?: ProfilePayload["portrait"]["knowledge_graph"];
}) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const [selectedNode, setSelectedNode] = useState<{
    label: string;
    type: string;
    mastery?: number;
    attempts: number;
    description?: string;
  } | null>(null);

  const graphData = useMemo(() => {
    const nodes = (graph?.nodes ?? []).filter(
      (node) => node.type !== "question",
    );
    const nodeIds = new Set(nodes.map((node) => node.id));
    const edges = (graph?.edges ?? []).filter(
      (edge) => nodeIds.has(edge.source) && nodeIds.has(edge.target),
    );

    return { nodes, edges };
  }, [graph]);

  useEffect(() => {
    if (!containerRef.current || graphData.nodes.length === 0) return;

    const nodeColor = (node: (typeof graphData.nodes)[number]) => {
      if (node.type === "exam") return "#f59e0b";
      if (node.type === "subject") return "#7c3aed";
      if (node.type === "chapter") return "#0f766e";
      if (node.mastery == null) return "#94a3b8";
      if (node.mastery < 50) return "#ef4444";
      if (node.mastery < 80) return "#f59e0b";
      return "#22c55e";
    };

    const nodeSize = (type: string) => {
      if (type === "exam") return 42;
      if (type === "subject") return 36;
      if (type === "chapter") return 30;
      return 34;
    };

    const cy = cytoscape({
      container: containerRef.current,
      elements: {
        nodes: graphData.nodes.map((node) => ({
          data: {
            id: node.id,
            label: node.label || "未命名节点",
            type: node.type,
            description: node.description || "",
            mastery: node.mastery,
            attempts: node.attempts ?? 0,
            color: nodeColor(node),
            size: nodeSize(node.type),
            borderColor:
              node.type === "knowledge_point" && node.mastery == null
                ? "#cbd5e1"
                : "#ffffff",
          },
        })),
        edges: graphData.edges.map((edge) => ({
          data: {
            id: edge.id,
            source: edge.source,
            target: edge.target,
            label: edge.label || edge.type,
            type: edge.type,
          },
        })),
      },
      style: [
        {
          selector: "node",
          style: {
            label: "data(label)",
            shape: "ellipse",
            width: "data(size)",
            height: "data(size)",
            "background-color": "data(color)",
            "border-color": "data(borderColor)",
            "border-width": 2,
            color: "#475569",
            "font-size": 11,
            "font-weight": 600,
            "text-wrap": "wrap",
            "text-max-width": "150px",
            "text-valign": "bottom",
            "text-margin-y": 9,
            "overlay-opacity": 0,
          },
        },
        {
          selector: 'node[type = "exam"]',
          style: {
            shape: "triangle",
            "font-size": 12,
            "font-weight": 700,
            "text-max-width": "180px",
          },
        },
        {
          selector: 'node[type = "subject"]',
          style: {
            shape: "diamond",
          },
        },
        {
          selector: 'node[type = "knowledge_point"]',
          style: {
            shape: "round-rectangle",
            "text-max-width": "130px",
          },
        },
        {
          selector: "node:selected",
          style: {
            color: "#0f172a",
            "border-color": "#0f172a",
            "border-width": 3,
            "z-index": 999,
          },
        },
        {
          selector: "edge",
          style: {
            width: 1.4,
            "line-color": "#cbd5e1",
            "target-arrow-color": "#cbd5e1",
            "target-arrow-shape": "triangle",
            "curve-style": "bezier",
            opacity: 0.75,
          },
        },
        {
          selector: 'edge[type = "prerequisite"]',
          style: {
            "line-style": "dashed",
            "line-color": "#94a3b8",
            "target-arrow-color": "#94a3b8",
          },
        },
      ],
      layout: {
        name: "cose",
        animate: false,
        fit: true,
        padding: 60,
        idealEdgeLength: 150,
        nodeRepulsion: 90000,
        nodeOverlap: 18,
        componentSpacing: 100,
        gravity: 0.35,
      },
    });

    cy.on("tap", "node", (event) => {
      const data = event.target.data();
      setSelectedNode({
        label: data.label,
        type: data.type,
        mastery: data.mastery,
        attempts: data.attempts,
        description: data.description,
      });
    });

    requestAnimationFrame(() => {
      cy.resize();
      cy.fit(undefined, 44);
      cy.center();
    });

    return () => {
      cy.destroy();
    };
  }, [graphData]);

  if (graphData.nodes.length === 0) {
    return (
      <div className="rounded-2xl border border-dashed border-[var(--outline-variant)] bg-[var(--surface)] px-5 py-10 text-center">
        <p className="font-medium text-[var(--text-main)]">暂无知识图谱数据</p>
        <p className="mt-2 text-sm text-[var(--text-muted)]">
          激活考试并维护知识点后，这里会展示你的学习覆盖情况。
        </p>
      </div>
    );
  }

  const selectedMastery =
    selectedNode?.type === "knowledge_point"
      ? selectedNode.mastery == null
        ? "未学习"
        : `${selectedNode.mastery}%`
      : null;

  return (
    <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_260px]">
      <div className="overflow-hidden rounded-2xl border border-[var(--outline-variant)] bg-[var(--surface)]">
        <div
          ref={(node) => {
            containerRef.current = node;
          }}
          className="h-[520px] min-h-[420px] w-full"
        />
      </div>

      <div className="space-y-4">
        <div className="rounded-2xl border border-[var(--outline-variant)] bg-[var(--surface)] p-4">
          <p className="text-sm font-medium text-[var(--text-main)]">
            掌握状态
          </p>
          <div className="mt-4 grid gap-3 text-sm text-[var(--text-muted)]">
            {[
              ["#94a3b8", "未学习"],
              ["#ef4444", "薄弱"],
              ["#f59e0b", "巩固中"],
              ["#22c55e", "已掌握"],
            ].map(([color, label]) => (
              <div key={label} className="flex items-center gap-2">
                <span
                  className="size-3 rounded-full"
                  style={{ backgroundColor: color }}
                />
                <span>{label}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="rounded-2xl border border-[var(--outline-variant)] bg-[var(--surface)] p-4">
          <p className="text-sm font-medium text-[var(--text-main)]">
            {selectedNode?.label ?? "点击节点查看详情"}
          </p>
          {selectedNode ? (
            <div className="mt-3 space-y-2 text-sm text-[var(--text-muted)]">
              <p>类型：{nodeTypeLabel(selectedNode.type)}</p>
              {selectedMastery ? <p>掌握度：{selectedMastery}</p> : null}
              {selectedNode.type === "knowledge_point" ? (
                <p>练习次数：{selectedNode.attempts}</p>
              ) : null}
              {selectedNode.description ? (
                <p>{selectedNode.description}</p>
              ) : null}
            </div>
          ) : (
            <p className="mt-3 text-sm text-[var(--text-muted)]">
              图中不展示题目节点，只展示考试、科目、章节与知识点。
            </p>
          )}
        </div>
      </div>
    </div>
  );
}

function nodeTypeLabel(type: string) {
  switch (type) {
    case "exam":
      return "考试";
    case "subject":
      return "科目";
    case "chapter":
      return "章节";
    case "knowledge_point":
      return "知识点";
    default:
      return type;
  }
}

export default function ProfilePage() {
  const storedSession = readStoredSession();
  const [me, setMe] = useState<MePayload | null>(null);
  const [profile, setProfile] = useState<ProfilePayload | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      const examId = storedSession?.activeExam?.id ?? "";

      try {
        const [meResponse, profileResponse] = await Promise.all([
          authFetch(`${API_BASE}/me`, {
            headers: authHeaders(),
            cache: "no-store",
          }),
          authFetch(
            `${API_BASE}/learner/profile${examId ? `?exam_id=${encodeURIComponent(examId)}` : ""}`,
            {
              headers: authHeaders(),
              cache: "no-store",
            },
          ),
        ]);

        if (!cancelled && meResponse.ok) {
          const mePayload = (await meResponse.json()) as { data?: MePayload };
          setMe(mePayload.data ?? null);
        }

        if (!cancelled && profileResponse.ok) {
          const profilePayload = (await profileResponse.json()) as {
            data?: ProfilePayload;
          };
          setProfile(profilePayload.data ?? null);
        }
      } catch {
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    void load();
    return () => {
      cancelled = true;
    };
  }, [storedSession?.activeExam?.id]);

  const displayName = useMemo(() => {
    const email = me?.user.email ?? storedSession?.user?.email ?? "learner";
    return email.split("@")[0] || "learner";
  }, [me?.user.email, storedSession?.user?.email]);

  const activeExamName =
    me?.active_exam_enrollment?.exam_name ??
    storedSession?.activeExam?.name ??
    null;

  const archive = profile?.archive;
  const records = profile?.records;
  const portrait = profile?.portrait;

  return (
    <div className="mx-auto w-full max-w-[var(--container-max-width)] px-[var(--margin-mobile)] py-6 md:px-[var(--margin-desktop)]">
      <div className="mb-8 flex flex-col gap-3">
        <h1 className="font-heading text-3xl font-bold tracking-tight text-[var(--text-main)]">
          用户中心
        </h1>
      </div>

      <Tabs defaultValue="archive" className="gap-6">
        <TabsList className="h-auto rounded-full bg-[var(--surface-container)] p-1">
          <TabsTrigger value="archive" className="rounded-full px-5 py-2">
            档案
          </TabsTrigger>
          <TabsTrigger value="records" className="rounded-full px-5 py-2">
            记录
          </TabsTrigger>
          <TabsTrigger value="portrait" className="rounded-full px-5 py-2">
            画像
          </TabsTrigger>
        </TabsList>

        <TabsContent value="archive" className="mt-0">
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
            <Card className="rounded-2xl border-0 bg-[var(--surface-container-lowest)] shadow-sm lg:col-span-8">
              <CardContent className="flex flex-col gap-8 py-8">
                <div className="flex flex-col gap-6 md:flex-row md:items-start md:justify-between">
                  <div className="flex items-center gap-5">
                    <div className="flex size-24 items-center justify-center rounded-full bg-[var(--secondary)]/10 text-[var(--secondary)]">
                      <UserRound className="size-10" />
                    </div>
                    <div className="space-y-2">
                      <div className="flex flex-wrap items-center gap-2">
                        <h2 className="font-heading text-2xl font-semibold text-[var(--text-main)]">
                          {displayName}
                        </h2>
                        <Badge className="rounded-full bg-[var(--secondary)]/10 text-[var(--secondary)] hover:bg-[var(--secondary)]/10">
                          Learner
                        </Badge>
                      </div>
                      <p className="text-sm text-[var(--text-muted)]">
                        {me?.user.email ??
                          storedSession?.user?.email ??
                          "未获取到邮箱"}
                      </p>
                      <div className="flex flex-wrap gap-2">
                        <Badge
                          variant="outline"
                          className="rounded-full border-[var(--outline-variant)] bg-[var(--surface)]"
                        >
                          {activeExamName
                            ? `激活考试：${activeExamName}`
                            : "尚未激活考试"}
                        </Badge>
                        {archive?.streak_status ? (
                          <Badge
                            variant="outline"
                            className="rounded-full border-[var(--outline-variant)] bg-[var(--surface)]"
                          >
                            连击状态：{archive.streak_status}
                          </Badge>
                        ) : null}
                      </div>
                    </div>
                  </div>
                  <Link href="/practice/setup">
                    <Button className="rounded-full">开始练习</Button>
                  </Link>
                </div>

                <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
                  {[
                    {
                      label: "累计练习",
                      value: archive?.total_sessions ?? 0,
                      icon: BookOpen,
                    },
                    {
                      label: "累计作答",
                      value: archive?.total_answered ?? 0,
                      icon: Target,
                    },
                    {
                      label: "累计 XP",
                      value: archive?.total_xp ?? 0,
                      icon: Sparkles,
                    },
                    {
                      label: "金币余额",
                      value: archive?.coins_balance ?? 0,
                      icon: Coins,
                    },
                  ].map((item) => (
                    <div
                      key={item.label}
                      className="rounded-2xl border border-[var(--outline-variant)] bg-[var(--surface)] p-5"
                    >
                      <div className="mb-3 flex size-10 items-center justify-center rounded-full bg-[var(--secondary)]/10 text-[var(--secondary)]">
                        <item.icon className="size-5" />
                      </div>
                      <p className="text-2xl font-semibold text-[var(--text-main)]">
                        {item.value}
                      </p>
                      <p className="mt-1 text-sm text-[var(--text-muted)]">
                        {item.label}
                      </p>
                    </div>
                  ))}
                </div>

                <div className="space-y-3">
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-[var(--text-muted)]">累计正确率</span>
                    <span className="font-medium text-[var(--text-main)]">
                      {loading ? "加载中..." : `${archive?.accuracy ?? 0}%`}
                    </span>
                  </div>
                  <Progress
                    value={archive?.accuracy ?? 0}
                    className="h-3 bg-[var(--surface-container-high)] [&>[data-slot=progress-indicator]]:bg-[var(--secondary)]"
                  />
                </div>
              </CardContent>
            </Card>

            <Card className="rounded-2xl border-0 bg-[var(--primary-container)] text-white shadow-sm lg:col-span-4">
              <CardContent className="grid h-full gap-6 py-8">
                <div className="flex items-center gap-3">
                  <Award className="size-5 text-[var(--secondary-fixed)]" />
                  <h3 className="font-heading text-xl font-semibold">
                    档案概览
                  </h3>
                </div>
                <div className="grid gap-4">
                  <div className="flex items-center justify-between rounded-2xl bg-white/10 px-4 py-3">
                    <span className="text-sm text-white/75">学习状态</span>
                    <span className="font-medium">
                      {activeExamName ? "进行中" : "未开始"}
                    </span>
                  </div>
                  <div className="flex items-center justify-between rounded-2xl bg-white/10 px-4 py-3">
                    <span className="text-sm text-white/75">当前连击</span>
                    <span className="font-medium">
                      {archive?.current_streak ?? 0} 天
                    </span>
                  </div>
                  <div className="flex items-center justify-between rounded-2xl bg-white/10 px-4 py-3">
                    <span className="text-sm text-white/75">最佳连击</span>
                    <span className="font-medium">
                      {archive?.best_streak ?? 0} 天
                    </span>
                  </div>
                  <div className="flex items-center justify-between rounded-2xl bg-white/10 px-4 py-3">
                    <span className="text-sm text-white/75">最近学习</span>
                    <span className="font-medium">
                      {formatDate(archive?.last_studied_at)}
                    </span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="records" className="mt-0">
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
            <div className="grid gap-4 sm:grid-cols-3 lg:col-span-12">
              {[
                {
                  label: "近 30 天练习",
                  value: records?.monthly_sessions ?? 0,
                  icon: CalendarDays,
                },
                {
                  label: "近 30 天作答",
                  value: records?.monthly_answered ?? 0,
                  icon: BookOpen,
                },
                {
                  label: "近 30 天 XP",
                  value: records?.monthly_xp ?? 0,
                  icon: Sparkles,
                },
              ].map((item) => (
                <Card
                  key={item.label}
                  className="rounded-2xl border-0 bg-[var(--surface-container-lowest)] shadow-sm"
                >
                  <CardContent className="flex items-center gap-4 py-6">
                    <div className="flex size-11 items-center justify-center rounded-full bg-[var(--secondary)]/10 text-[var(--secondary)]">
                      <item.icon className="size-5" />
                    </div>
                    <div>
                      <p className="text-2xl font-semibold text-[var(--text-main)]">
                        {item.value}
                      </p>
                      <p className="text-sm text-[var(--text-muted)]">
                        {item.label}
                      </p>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>

            <Card className="rounded-2xl border-0 bg-[var(--surface-container-lowest)] shadow-sm lg:col-span-8">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <CalendarDays className="size-5 text-[var(--secondary)]" />
                  最近练习记录
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                {!records || records.recent_sessions.length === 0 ? (
                  <div className="rounded-2xl border border-dashed border-[var(--outline-variant)] bg-[var(--surface)] px-5 py-10 text-center">
                    <p className="font-medium text-[var(--text-main)]">
                      还没有可展示的练习记录
                    </p>
                    <p className="mt-2 text-sm text-[var(--text-muted)]">
                      完成一次练习后，这里会展示你的时间线、正确率与学习轨迹。
                    </p>
                  </div>
                ) : (
                  records.recent_sessions.map((session, index) => (
                    <div key={session.id}>
                      <div className="flex flex-col gap-3 rounded-2xl bg-[var(--surface)] p-4 md:flex-row md:items-center md:justify-between">
                        <div className="space-y-2">
                          <div className="flex flex-wrap items-center gap-2">
                            <Badge
                              className={cn(
                                "rounded-full",
                                sessionStatusTone(session.status),
                              )}
                            >
                              {sessionStatusLabel(session.status)}
                            </Badge>
                            <span className="text-sm text-[var(--text-muted)]">
                              {formatDate(session.created_at)}
                            </span>
                          </div>
                          <div className="flex flex-wrap gap-4 text-sm text-[var(--text-muted)]">
                            <span>
                              答题 {session.answered_count}/
                              {session.total_count}
                            </span>
                            <span>正确率 {session.accuracy}%</span>
                            <span>XP +{session.xp_earned}</span>
                            <span>金币 +{session.coins_earned}</span>
                            <span>用时 {session.duration_minutes} 分钟</span>
                          </div>
                        </div>
                      </div>
                      {index < records.recent_sessions.length - 1 ? (
                        <Separator className="my-2 bg-[var(--outline-variant)]" />
                      ) : null}
                    </div>
                  ))
                )}
              </CardContent>
            </Card>

            <Card className="rounded-2xl border-0 bg-[var(--surface-container-lowest)] shadow-sm lg:col-span-4">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <BarChart3 className="size-5 text-[var(--secondary)]" />
                  记录摘要
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="rounded-2xl bg-[var(--surface)] p-4 text-sm text-[var(--text-muted)]">
                  这里的数据来自已保存的练习会话与答题结果，不再使用静态占位内容。
                </div>
                <div className="grid gap-3">
                  <div className="flex items-center justify-between rounded-2xl border border-[var(--outline-variant)] bg-[var(--surface)] px-4 py-3">
                    <span className="text-sm text-[var(--text-muted)]">
                      累计会话
                    </span>
                    <span className="font-medium text-[var(--text-main)]">
                      {archive?.total_sessions ?? 0}
                    </span>
                  </div>
                  <div className="flex items-center justify-between rounded-2xl border border-[var(--outline-variant)] bg-[var(--surface)] px-4 py-3">
                    <span className="text-sm text-[var(--text-muted)]">
                      累计正确率
                    </span>
                    <span className="font-medium text-[var(--text-main)]">
                      {archive?.accuracy ?? 0}%
                    </span>
                  </div>
                  <div className="flex items-center justify-between rounded-2xl border border-[var(--outline-variant)] bg-[var(--surface)] px-4 py-3">
                    <span className="text-sm text-[var(--text-muted)]">
                      最近学习
                    </span>
                    <span className="font-medium text-[var(--text-main)]">
                      {formatDate(archive?.last_studied_at)}
                    </span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="portrait" className="mt-0">
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
            <Card className="rounded-2xl border-0 bg-[var(--surface-container-lowest)] shadow-sm lg:col-span-12">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <BrainCircuit className="size-5 text-[var(--secondary)]" />
                  综合学习画像
                </CardTitle>
              </CardHeader>
              <CardContent className="grid gap-4 md:grid-cols-3">
                {[
                  {
                    label: "总答题覆盖",
                    value: portrait?.total_attempts ?? 0,
                    icon: Target,
                  },
                  {
                    label: "整体正确率",
                    value: `${portrait?.overall_accuracy ?? 0}%`,
                    icon: TrendingUp,
                  },
                  {
                    label: "激活考试",
                    value: activeExamName ?? "未设置",
                    icon: Compass,
                  },
                ].map((item) => (
                  <div
                    key={item.label}
                    className="rounded-2xl border border-[var(--outline-variant)] bg-[var(--surface)] p-5"
                  >
                    <div className="mb-3 flex size-10 items-center justify-center rounded-full bg-[var(--secondary)]/10 text-[var(--secondary)]">
                      <item.icon className="size-5" />
                    </div>
                    <p className="text-2xl font-semibold text-[var(--text-main)]">
                      {item.value}
                    </p>
                    <p className="mt-1 text-sm text-[var(--text-muted)]">
                      {item.label}
                    </p>
                  </div>
                ))}
              </CardContent>
            </Card>

            <Card className="rounded-2xl border-0 bg-[var(--surface-container-lowest)] shadow-sm lg:col-span-12">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Compass className="size-5 text-[var(--secondary)]" />
                  用户知识图谱
                </CardTitle>
              </CardHeader>
              <CardContent>
                <LearnerKnowledgeGraph graph={portrait?.knowledge_graph} />
              </CardContent>
            </Card>

            <div className="grid gap-6 lg:col-span-6">
              <PointList
                title="薄弱知识点"
                points={portrait?.weak_points ?? []}
                emptyCopy="当前还没有足够的真实答题数据来识别薄弱知识点。"
                tone="weak"
              />
            </div>

            <div className="grid gap-6 lg:col-span-6">
              <PointList
                title="相对优势知识点"
                points={portrait?.strong_points ?? []}
                emptyCopy="当前还没有足够的真实答题数据来识别优势知识点。"
                tone="strong"
              />
            </div>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
