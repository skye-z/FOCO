"use client";

import * as React from "react";
import { Bot, LoaderCircle, Package, ShieldCheck } from "lucide-react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";

import {
  authFetch,
  clearStoredSession,
  readBrowserAccessToken,
} from "@/lib/auth-session";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";

type SettingsPayload = {
  llm: {
    provider: string;
    base_url: string;
    model: string;
    enabled: boolean;
    configured: boolean;
  };
  registration_open: boolean;
};

type ImportReport = {
  exams_imported: number;
  subjects_imported: number;
  chapters_imported: number;
  knowledge_points_imported: number;
  knowledge_point_edges_imported: number;
  questions_imported: number;
  question_versions_imported: number;
  question_version_knowledge_points_imported: number;
  interactive_units_imported: number;
  interactive_unit_steps_imported: number;
  validation_errors: string[];
};

type APIEnvelope<T> = {
  data?: T;
  error?: string | null;
};

async function readAPIEnvelope<T>(res: Response): Promise<APIEnvelope<T>> {
  const text = await res.text();
  if (!text) {
    return {};
  }
  try {
    return JSON.parse(text) as APIEnvelope<T>;
  } catch {
    return { error: text };
  }
}

export default function SettingsPage() {
  const router = useRouter();
  const [loading, setLoading] = React.useState(true);
  const [saving, setSaving] = React.useState(false);
  const [provider, setProvider] = React.useState("openai");
  const [baseUrl, setBaseUrl] = React.useState("");
  const [model, setModel] = React.useState("");
  const [apiKey, setApiKey] = React.useState("");
  const [enabled, setEnabled] = React.useState(false);
  const [configured, setConfigured] = React.useState(false);
  const [registrationOpen, setRegistrationOpen] = React.useState(true);
  const [importFileName, setImportFileName] = React.useState("");
  const [importContent, setImportContent] = React.useState("");
  const [importReport, setImportReport] = React.useState<ImportReport | null>(
    null,
  );
  const [importing, setImporting] = React.useState(false);

  const loadSettings = React.useCallback(async () => {
    const token = readBrowserAccessToken();
    if (!token) {
      clearStoredSession();
      router.replace("/");
      return;
    }

    try {
      const res = await authFetch("/api/v1/admin/settings", {
        headers: { Authorization: `Bearer ${token}` },
        cache: "no-store",
      });
      if (!res.ok) throw new Error("load settings failed");
      const payload = (await res.json()).data as SettingsPayload;
      setProvider(payload.llm.provider || "openai");
      setBaseUrl(payload.llm.base_url || "");
      setModel(payload.llm.model || "");
      setEnabled(Boolean(payload.llm.enabled));
      setConfigured(Boolean(payload.llm.configured));
      setRegistrationOpen(Boolean(payload.registration_open));
    } catch {
      toast.error("设置加载失败", { description: "请刷新页面后重试。" });
    } finally {
      setLoading(false);
    }
  }, [router]);

  React.useEffect(() => {
    void loadSettings();
  }, [loadSettings]);

  async function handleSave() {
    const token = readBrowserAccessToken();
    if (!token) {
      clearStoredSession();
      router.replace("/");
      return;
    }

    setSaving(true);
    try {
      const res = await authFetch("/api/v1/admin/settings", {
        method: "PATCH",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          llm: {
            provider,
            base_url: baseUrl,
            api_key: apiKey,
            model,
            enabled,
          },
          registration_open: registrationOpen,
        }),
      });
      if (!res.ok) {
        throw new Error(await res.text());
      }
      toast.success("设置已保存", {
        description: "LLM 配置和注册开关已更新。",
      });
      setApiKey("");
      await loadSettings();
    } catch (error) {
      toast.error("设置保存失败", {
        description:
          error instanceof Error && error.message
            ? error.message
            : "请稍后重试。",
      });
    } finally {
      setSaving(false);
    }
  }

  async function handleExport() {
    const token = readBrowserAccessToken();
    if (!token) {
      clearStoredSession();
      router.replace("/");
      return;
    }

    try {
      const res = await authFetch("/api/v1/admin/content-package/export", {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) {
        throw new Error(await res.text());
      }
      const blob = await res.blob();
      const url = window.URL.createObjectURL(blob);
      const anchor = document.createElement("a");
      anchor.href = url;
      anchor.download = "foco-content-package.json";
      anchor.click();
      window.URL.revokeObjectURL(url);
      toast.success("导出成功", { description: "内容包已开始下载。" });
    } catch (error) {
      toast.error("导出失败", {
        description:
          error instanceof Error && error.message
            ? error.message
            : "请稍后重试。",
      });
    }
  }

  async function handleImport() {
    const token = readBrowserAccessToken();
    if (!token) {
      clearStoredSession();
      router.replace("/");
      return;
    }
    if (!importContent) {
      toast.error("导入失败", {
        description: "请先选择一个内容包 JSON 文件。",
      });
      return;
    }

    try {
      setImporting(true);
      const res = await authFetch("/api/v1/admin/content-package/import", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: importContent,
      });
      const envelope = await readAPIEnvelope<ImportReport>(res);
      const payload = envelope.data;
      setImportReport(payload ?? null);
      if (!res.ok) {
        if (payload?.validation_errors?.length) {
          toast.error("导入校验失败", {
            description: "请先修复内容包结构错误。",
          });
          return;
        }
        throw new Error(envelope.error || "导入失败");
      }
      if (!payload) {
        throw new Error("导入接口未返回报告。");
      }
      toast.success("导入成功", {
        description: importFileName
          ? `${importFileName} 已导入。`
          : "内容包已导入。",
      });
    } catch (error) {
      toast.error("导入失败", {
        description:
          error instanceof Error && error.message
            ? error.message
            : "请检查 JSON 文件后重试。",
      });
    } finally {
      setImporting(false);
    }
  }

  async function handleFileChange(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    if (!file) return;
    setImportReport(null);
    setImportFileName(file.name);
    setImportContent(await file.text());
  }

  if (loading) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <LoaderCircle className="size-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <main className="mx-auto max-w-6xl px-6 py-8">
      <div className="mb-8">
        <h1 className="text-2xl font-bold tracking-tight">设置</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          管理内容包、LLM 配置和注册开放策略。
        </p>
      </div>

      <div className="grid gap-6">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Package className="size-5 text-primary" />
              内容包导入与导出
            </CardTitle>
            <CardDescription>
              支持导出当前内容包，以及导入 JSON 内容包。
            </CardDescription>
          </CardHeader>
          <CardContent className="grid gap-4 md:grid-cols-[minmax(0,1fr)_auto_auto] md:items-end">
            <div className="space-y-2">
              <Label htmlFor="content-package-file">选择内容包文件</Label>
              <Input
                id="content-package-file"
                type="file"
                accept="application/json"
                onChange={handleFileChange}
              />
            </div>
            <Button type="button" variant="outline" onClick={handleImport} disabled={importing || !importContent}>
              {importing ? "导入中..." : "导入内容包"}
            </Button>
            <Button type="button" variant="outline" onClick={handleExport}>
              导出内容包
            </Button>
          </CardContent>
          {importing ? (
            <CardContent className="pt-0">
              <div className="flex items-center gap-3 rounded-xl border bg-muted/20 p-4">
                <LoaderCircle className="size-5 animate-spin text-primary" />
                <div>
                  <p className="text-sm font-medium">正在导入内容包...</p>
                  <p className="text-xs text-muted-foreground">请勿关闭页面，导入完成后将自动显示结果。</p>
                </div>
              </div>
            </CardContent>
          ) : importReport ? (
            <CardContent className="pt-0">
              <div className="rounded-xl border bg-muted/20 p-4">
                <h3 className="mb-3 text-sm font-semibold text-foreground">
                  导入结果统计
                </h3>
                <div className="grid gap-2 text-sm text-muted-foreground md:grid-cols-2">
                  <p>考试：{importReport.exams_imported}</p>
                  <p>科目：{importReport.subjects_imported}</p>
                  <p>章节：{importReport.chapters_imported}</p>
                  <p>知识点：{importReport.knowledge_points_imported}</p>
                  <p>
                    知识点关系：{importReport.knowledge_point_edges_imported}
                  </p>
                  <p>题目：{importReport.questions_imported}</p>
                  <p>题目版本：{importReport.question_versions_imported}</p>
                  <p>
                    题目知识点关联：
                    {importReport.question_version_knowledge_points_imported}
                  </p>
                  <p>交互单元：{importReport.interactive_units_imported}</p>
                  <p>交互单元步骤：{importReport.interactive_unit_steps_imported}</p>
                </div>

                {importReport.validation_errors?.length ? (
                  <div className="mt-4 rounded-lg border border-destructive/30 bg-destructive/5 px-4 py-3">
                    <h4 className="mb-2 text-sm font-semibold text-destructive">
                      结构校验错误
                    </h4>
                    <ul className="list-disc space-y-1 pl-5 text-sm text-destructive">
                      {importReport.validation_errors.map((item) => (
                        <li key={item}>{item}</li>
                      ))}
                    </ul>
                  </div>
                ) : null}
              </div>
            </CardContent>
          ) : null}
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Bot className="size-5 text-primary" />
              LLM API 配置
            </CardTitle>
            <CardDescription>
              配置会持久化到数据库。API Key 不回显明文，重新填写会覆盖旧值。
            </CardDescription>
          </CardHeader>
          <CardContent className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label>提供商</Label>
              <Select
                value={provider}
                onValueChange={(value) => setProvider(value ?? "openai")}
              >
                <SelectTrigger>{provider || "选择提供商"}</SelectTrigger>
                <SelectContent>
                  <SelectItem value="openai">OpenAI</SelectItem>
                  <SelectItem value="openrouter">OpenRouter</SelectItem>
                  <SelectItem value="anthropic">Anthropic</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>模型</Label>
              <Input
                value={model}
                onChange={(event) => setModel(event.target.value)}
                placeholder="如：gpt-4.1-mini"
              />
            </div>
            <div className="space-y-2 md:col-span-2">
              <Label>Base URL</Label>
              <Input
                value={baseUrl}
                onChange={(event) => setBaseUrl(event.target.value)}
                placeholder="https://api.openai.com/v1"
              />
            </div>
            <div className="space-y-2 md:col-span-2">
              <Label>API Key</Label>
              <Input
                value={apiKey}
                onChange={(event) => setApiKey(event.target.value)}
                placeholder={
                  configured ? "已配置，重新填写将覆盖旧值" : "输入新的 API Key"
                }
              />
            </div>
            <div className="space-y-2">
              <Label>启用状态</Label>
              <Select
                value={enabled ? "enabled" : "disabled"}
                onValueChange={(value) => setEnabled(value === "enabled")}
              >
                <SelectTrigger>{enabled ? "已启用" : "未启用"}</SelectTrigger>
                <SelectContent>
                  <SelectItem value="enabled">已启用</SelectItem>
                  <SelectItem value="disabled">未启用</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <ShieldCheck className="size-5 text-primary" />
              注册开放策略
            </CardTitle>
            <CardDescription>
              控制 learner 端是否允许新用户注册。
            </CardDescription>
          </CardHeader>
          <CardContent className="flex flex-wrap items-end gap-4">
            <div className="w-full max-w-xs space-y-2">
              <Label>注册状态</Label>
              <Select
                value={registrationOpen ? "open" : "closed"}
                onValueChange={(value) => setRegistrationOpen(value === "open")}
              >
                <SelectTrigger>
                  {registrationOpen ? "开放注册" : "关闭注册"}
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="open">开放注册</SelectItem>
                  <SelectItem value="closed">关闭注册</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <p className="text-sm text-muted-foreground">
              关闭后，learner 登录仍可用，但注册入口会隐藏并显示中文提示。
            </p>
          </CardContent>
        </Card>

        <div className="flex justify-end">
          <Button type="button" onClick={handleSave} disabled={saving}>
            {saving ? "保存中..." : "保存设置"}
          </Button>
        </div>
      </div>
    </main>
  );
}
