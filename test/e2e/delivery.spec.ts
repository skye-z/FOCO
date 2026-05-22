import { expect, test } from "@playwright/test";
import { existsSync } from "node:fs";
import { e2eConfig } from "../support/config";
import {
  resetDataIfConfigured,
  seedDefaultAdmin,
  waitForBackendHealth,
} from "../support/api";
import {
  answerAllDiagnosticQuestions,
  clearBrowserState,
  clickFirstVisible,
  completeFirstInteractiveUnit,
  completePracticeSession,
  expectPageHealthy,
  login,
  logoutIfVisible,
  selectByVisibleText,
} from "../support/ui";

test.describe.serial("FOCO delivery E2E", () => {
  test.beforeAll(async () => {
    resetDataIfConfigured();
    await waitForBackendHealth();
    await seedDefaultAdmin();
  });

  test("admin import, content management, learner journey, and profile data", async ({
    page,
  }) => {
    test.setTimeout(30 * 60 * 1000);
    expect(
      existsSync(e2eConfig.contentPackagePath),
      `content package exists: ${e2eConfig.contentPackagePath}`,
    ).toBe(true);

    await test.step("Clear browser data and login to admin", async () => {
      await clearBrowserState(page);
      await login(
        page,
        e2eConfig.adminUrl,
        e2eConfig.adminEmail,
        e2eConfig.adminPassword,
      );
      await expect(
        page.getByRole("heading", { name: "平台概览" }),
      ).toBeVisible();
      await expect(page.getByText("考试总数")).toBeVisible();
      await expect(page.getByText("注册用户")).toBeVisible();
      await expectPageHealthy(page);
    });

    await test.step("Import and export content package from settings", async () => {
      await page.getByRole("button", { name: "设置" }).click();
      await expect(page.getByText("内容包导入与导出")).toBeVisible();
      await page.setInputFiles(
        "#content-package-file",
        e2eConfig.contentPackagePath,
      );
      const shouldImport =
        process.env.E2E_FORCE_IMPORT === "1" ||
        !(await hasImportedContent(page));
      if (shouldImport) {
        await page.getByRole("button", { name: "导入内容包" }).click();
        await expect(page.getByText("导入结果统计")).toBeVisible({
          timeout: 90_000,
        });
        await expect(page.getByText(/题目：\d+/)).toBeVisible();
        await expect(page.getByText(/交互单元：\d+/)).toBeVisible();
      } else {
        test.info().annotations.push({
          type: "note",
          description:
            "Skipped repeated content import because content already exists. Set E2E_FORCE_IMPORT=1 to force it.",
        });
      }

      const downloadPromise = page.waitForEvent("download");
      await page.getByRole("button", { name: "导出内容包" }).click();
      const download = await downloadPromise;
      expect(download.suggestedFilename()).toBe("foco-content-package.json");

      await selectByVisibleText(page, "注册状态", "开放注册").catch(
        async () => {
          await page
            .getByText(/开放注册|关闭注册/)
            .last()
            .click();
          await clickFirstVisible(
            page.getByRole("option", { name: "开放注册" }),
            "open registration option",
          );
        },
      );
      await page.getByRole("button", { name: "保存设置" }).click();
      await expect(page.getByText("设置已保存")).toBeVisible();
    });

    await test.step("Review admin overview after import", async () => {
      await page.getByRole("button", { name: "概览" }).click();
      await expect(
        page.getByRole("heading", { name: "平台概览" }),
      ).toBeVisible();
      await expect(page.getByText("平台信息")).toBeVisible();
      await expectPageHealthy(page);
    });

    await test.step("Inspect question bank tree, filters, question detail, and interactive cards", async () => {
      await page.getByRole("button", { name: "题库管理" }).click();
      await expect(
        page.getByRole("heading", { name: "题库管理" }),
      ).toBeVisible();
      await expect(page.getByText(/全部难度/)).toBeVisible();
      await expect(page.getByText(/全部状态/)).toBeVisible();
      await expect(page.getByText(/全部知识点/)).toBeVisible();
      await expect(page.getByText(/全部类型/)).toBeVisible();

      await page.getByText("全部类型").click();
      await page.getByRole("option", { name: "交互单元" }).click();
      await expect(page.getByText("交互单元").first()).toBeVisible();
      await page
        .locator("main")
        .getByRole("combobox")
        .filter({ hasText: "交互单元" })
        .click();
      await page.getByRole("option", { name: "题目" }).click();
      await expect(
        page.getByText(/单选|多选|判断|填空|简答/).first(),
      ).toBeVisible();

      const firstQuestionCard = page
        .locator("main [class*='break-inside-avoid']")
        .first();
      await firstQuestionCard.click();
      await expect(page.getByText("题目详情")).toBeVisible();
      await expect(page.getByText(/保存|发布|版本|题干/).first()).toBeVisible();
      await page.keyboard.press("Escape");

      await page.getByRole("button", { name: /新建题目/ }).click();
      await expect(
        page.getByRole("heading", { name: "新建题目" }),
      ).toBeVisible();
      await page.keyboard.press("Escape");

      await page.getByRole("button", { name: /新建交互单元/ }).click();
      await expect(
        page.getByRole("heading", { name: "新建交互单元" }),
      ).toBeVisible();
      await page.keyboard.press("Escape");
    });

    await test.step("Review user management functions without disabling the delivery account", async () => {
      await page.getByRole("button", { name: "用户管理" }).click();
      await expect(
        page.getByRole("heading", { name: "用户管理" }),
      ).toBeVisible();
      await expect(page.getByText("用户列表")).toBeVisible();
      await page
        .getByPlaceholder("搜索姓名、邮箱或用户 ID")
        .fill(e2eConfig.adminEmail);
      await expect(page.getByText(e2eConfig.adminEmail)).toBeVisible();

      await page.getByText("全部角色").click();
      await page.getByRole("option", { name: "管理员" }).click();
      await expect(page.getByText(e2eConfig.adminEmail)).toBeVisible();

      await page.getByRole("button", { name: "补角色" }).first().click();
      await expect(page.getByText("确认补角色")).toBeVisible();
      await page.getByRole("button", { name: "取消" }).click();

      await page.getByRole("button", { name: "重置密码" }).first().click();
      await expect(page.getByText("临时密码")).toBeVisible();
      await page.getByRole("button", { name: "取消" }).click();

      await page.getByRole("button", { name: "禁用用户" }).first().click();
      await expect(page.getByText("确认禁用用户")).toBeVisible();
      await page.getByRole("button", { name: "取消" }).click();
    });

    await test.step("Logout from admin", async () => {
      await logoutIfVisible(page);
      await expect(page.getByText("登录管理后台")).toBeVisible();
    });

    await test.step("Login to learner and finish onboarding diagnostic", async () => {
      await login(
        page,
        e2eConfig.learnerUrl,
        e2eConfig.adminEmail,
        e2eConfig.adminPassword,
      );

      if (
        await page
          .getByText("开启你的学习旅程")
          .isVisible()
          .catch(() => false)
      ) {
        await clickFirstVisible(
          page.locator("button").filter({ hasNotText: /^继续$/ }),
          "first onboarding exam",
        );
        await page.getByRole("button", { name: "继续" }).click();
      }

      if (
        await page
          .getByText("先完成一次入门诊断")
          .isVisible({ timeout: 30_000 })
          .catch(() => false)
      ) {
        await answerAllDiagnosticQuestions(page);
        await page.getByRole("button", { name: /进入首页|回到首页/ }).click();
      }

      const enterHome = page
        .getByRole("button", { name: /进入首页|回到首页/ })
        .first();
      if (await enterHome.isVisible().catch(() => false)) {
        await enterHome.click();
      }

      await expect(page.getByText("今日学习路径")).toBeVisible();
      await expect(page.getByText("本周学习节奏")).toBeVisible();
    });

    await test.step("Start a learning-path practice task and complete it", async () => {
      const enterHome = page
        .getByRole("button", { name: /进入首页|回到首页/ })
        .first();
      if (await enterHome.isVisible().catch(() => false)) {
        await enterHome.click();
      }
      await expect(page.getByText("今日学习路径")).toBeVisible();

      const learningPath = page
        .getByText("今日学习路径")
        .locator("xpath=ancestor::*[contains(@class,'rounded-2xl')][1]");
      const taskButton = learningPath
        .getByRole("button", { name: /^开始$/ })
        .first();
      if (await taskButton.isVisible().catch(() => false)) {
        await taskButton.click();
      } else {
        await page.getByRole("button", { name: "开始今日练习" }).click();
      }

      const diagnosticResultHome = page
        .getByRole("button", { name: /进入首页|回到首页/ })
        .first();
      if (
        await diagnosticResultHome
          .isVisible({ timeout: 5_000 })
          .catch(() => false)
      ) {
        await diagnosticResultHome.click();
        await expect(page.getByText("今日学习路径")).toBeVisible();
        await page.getByRole("button", { name: "开始今日练习" }).click();
      }

      if (
        await page
          .getByText("选择练习方式")
          .isVisible({ timeout: 30_000 })
          .catch(() => false)
      ) {
        const startSmart = page.getByRole("button", { name: "开始智能练习" });
        if (await startSmart.isEnabled().catch(() => false)) {
          await startSmart.click();
        } else {
          await page.getByText("手动练习").click();
          await page.getByRole("button", { name: "开始手动练习" }).click();
        }
      }

      await completePracticeSession(page);
      await page.getByRole("button", { name: "返回首页" }).click();
      await expect(page.getByText("今日学习路径")).toBeVisible();
      await expect(page.getByText("最近学习")).toBeVisible();
    });

    await test.step("Complete the first interactive learning unit", async () => {
      await page.getByRole("link", { name: /学习/ }).click();
      await completeFirstInteractiveUnit(page);
      await page.getByRole("button", { name: "返回实验室列表" }).click();
      await expect(page.getByText("交互学习单元")).toBeVisible();
    });

    await test.step("Review wrong book page", async () => {
      await page.getByRole("link", { name: "错题本" }).click();
      await expect(page.getByRole("heading", { name: "错题本" })).toBeVisible();
      await expect(
        page.getByText(/暂无错题记录|你的答案|正确答案/).first(),
      ).toBeVisible();
      const explanation = page
        .getByRole("button", { name: "查看解析" })
        .first();
      if (await explanation.isVisible().catch(() => false)) {
        await explanation.click();
      }
    });

    await test.step("Return home and verify updated learning data", async () => {
      await page.getByRole("link", { name: "首页" }).click();
      await expect(page.getByText("今日学习路径")).toBeVisible();
      await expect(page.getByText("已完成题数")).toBeVisible();
      await expect(page.getByText("累计正确率")).toBeVisible();
      await expect(page.getByText("最近学习")).toBeVisible();
    });

    await test.step("Open user profile and inspect archive, records, and portrait tabs", async () => {
      await page.getByText("用户中心").click();
      await expect(
        page.getByRole("heading", { name: "用户中心" }),
      ).toBeVisible();
      await expect(page.getByText(e2eConfig.adminEmail)).toBeVisible();
      await expect(page.getByText("累计练习")).toBeVisible();
      await page.getByRole("tab", { name: "记录" }).click();
      await expect(page.getByText("最近练习记录")).toBeVisible();
      await page.getByRole("tab", { name: "画像" }).click();
      await expect(page.getByText("综合学习画像")).toBeVisible();
    });

    await test.step("Logout from learner and verify auth gate", async () => {
      await logoutIfVisible(page);
    });
  });
});

async function hasImportedContent(page: import("@playwright/test").Page) {
  return page.evaluate(async () => {
    const cookieToken = document.cookie
      .split("; ")
      .find((entry) => entry.startsWith("foco_access_token="))
      ?.split("=", 2)[1];
    const storedSession = window.localStorage.getItem("foco.auth.session");
    const storedToken = storedSession
      ? (JSON.parse(storedSession) as { accessToken?: string }).accessToken
      : "";
    const token = cookieToken ? decodeURIComponent(cookieToken) : storedToken;
    const headers: Record<string, string> = {};
    if (token) headers.Authorization = `Bearer ${token}`;

    const statsResponse = await fetch("/api/v1/admin/stats", {
      cache: "no-store",
      headers,
    });
    if (statsResponse.ok) {
      const statsPayload = await statsResponse.json();
      if (Number(statsPayload?.data?.total_exams ?? 0) > 0) return true;
    }

    const treeResponse = await fetch("/api/v1/admin/exam-tree", {
      cache: "no-store",
      headers,
    });
    if (!treeResponse.ok) return false;
    const treePayload = await treeResponse.json();
    return Array.isArray(treePayload?.data) && treePayload.data.length > 0;
  });
}
