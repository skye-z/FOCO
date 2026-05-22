import { expect, type Locator, type Page } from "@playwright/test"

export async function clearBrowserState(page: Page) {
  await page.context().clearCookies()
  await page.context().addInitScript(() => {
    if (location.pathname === "/") {
      window.localStorage.removeItem("foco.auth.session")
      window.localStorage.removeItem("foco.auth.remembered_email")
      window.sessionStorage.clear()
      document.cookie = "foco_access_token=; Path=/; Max-Age=0; SameSite=Lax"
    }
  })
}

export async function login(page: Page, url: string, email: string, password: string) {
  await page.goto(url)
  await page.getByRole("textbox", { name: "邮箱" }).fill(email)
  await page.getByRole("textbox", { name: "密码" }).fill(password)
  await page.getByRole("button", { name: /^登录$/ }).click()
}

export async function logoutIfVisible(page: Page) {
  const logout = page.getByRole("button", { name: /退出登录/ }).or(page.getByLabel("退出登录"))
  if (await logout.first().isVisible().catch(() => false)) {
    await logout.first().click()
  }
}

export async function clickFirstVisible(locator: Locator, description: string) {
  const count = await locator.count()
  for (let index = 0; index < count; index += 1) {
    const item = locator.nth(index)
    if (await item.isVisible().catch(() => false)) {
      await item.click()
      return
    }
  }
  throw new Error(`No visible element found for ${description}`)
}

export async function selectByVisibleText(page: Page, triggerText: string | RegExp, optionText: string | RegExp) {
  await clickFirstVisible(page.getByText(triggerText).or(page.getByRole("combobox", { name: triggerText as string })), `select ${String(triggerText)}`)
  await clickFirstVisible(page.getByRole("option", { name: optionText }), `option ${String(optionText)}`)
}

export async function expectPageHealthy(page: Page) {
  await expect(page.locator("body")).not.toContainText(/无法加载|加载失败|请求失败|登录失败|未完成，请稍后重试/)
}

export async function waitForEither(page: Page, texts: Array<string | RegExp>) {
  await expect(
    texts
      .map((text) => page.getByText(text))
      .reduce((acc, locator) => acc.or(locator))
      .first(),
  ).toBeVisible()
}

export async function answerAllDiagnosticQuestions(page: Page) {
  await expect(page.getByText("先完成一次入门诊断")).toBeVisible()
  const questionTitles = page.getByText(/^第 \d+ 题$/)
  const count = await questionTitles.count()
  expect(count, "diagnostic question count").toBeGreaterThan(0)

  for (let index = 0; index < count; index += 1) {
    const questionCard = questionTitles
      .nth(index)
      .locator("xpath=ancestor::*[contains(@class,'rounded-2xl')][1]")
    await questionCard.getByRole("button").first().click()
  }

  await page.getByRole("button", { name: "提交测验" }).click()
  await waitForEither(page, ["引导测验完成", "回到首页"])
}

export async function completePracticeSession(page: Page) {
  await expect(page.getByText(/第 \d+ \/ \d+ 题/)).toBeVisible()

  for (let guard = 0; guard < 80; guard += 1) {
    if (await page.getByRole("button", { name: "查看结果" }).isVisible().catch(() => false)) {
      await page.getByRole("button", { name: "查看结果" }).click()
      await expect(page.getByText("练习完成！")).toBeVisible()
      return
    }

    const submit = page.getByRole("button", { name: "提交答案" })
    if (await submit.isVisible().catch(() => false)) {
      const optionButtons = page.locator("main button").filter({
        hasNotText: /提交答案|下一题|查看结果/,
      })
      await clickFirstVisible(optionButtons, "practice answer option")
      await submit.click()
      await waitForEither(page, ["回答正确！", "回答错误"])
    }

    const next = page.getByRole("button", { name: /下一题/ })
    if (await next.isVisible().catch(() => false)) {
      await next.click()
    }
  }

  throw new Error("Practice session did not complete before guard limit.")
}

export async function completeFirstInteractiveUnit(page: Page) {
  await expect(page.getByText("交互学习单元")).toBeVisible()
  await clickFirstVisible(page.locator("a[href^='/labs/']").first(), "first interactive unit")
  await expect(page.getByRole("button", { name: "开始学习" })).toBeVisible()
  await page.getByRole("button", { name: "开始学习" }).click()

  for (let guard = 0; guard < 30; guard += 1) {
    if (await page.getByText("单元完成!").isVisible().catch(() => false)) {
      return
    }

    await prepareCurrentInteractiveStep(page)

    const submit = page
      .getByRole("button", { name: /提交答案|提交标注|提交实验结果|提交公式/ })
      .first()
    if (await submit.isVisible().catch(() => false)) {
      await expect(submit).toBeEnabled()
      await submit.click()
      await waitForEither(page, ["回答正确!", "回答不正确"])
    }

    const finish = page.getByRole("button", { name: "完成单元" })
    if (await finish.isVisible().catch(() => false)) {
      await finish.click()
      await expect(page.getByText("单元完成!")).toBeVisible()
      return
    }

    const next = page.getByRole("button", { name: /下一题/ })
    if (await next.isVisible().catch(() => false)) {
      await next.click()
    }
  }

  throw new Error("Interactive unit did not complete before guard limit.")
}

async function prepareCurrentInteractiveStep(page: Page) {
  const observationAnswer = page.getByPlaceholder("填写计算结果或影响方向")
  if (await observationAnswer.isVisible().catch(() => false)) {
    const pageText = await page.locator("body").innerText()
    let answer = "-6"
    if (pageText.includes("100万减值")) answer = "-75万"
    else if (pageText.includes("折现率从8%上升到12%")) answer = "下降"
    else if (pageText.includes("违规严重程度=8")) answer = "-6"
    await observationAnswer.fill(answer)
  }

  const submitMarking = page.getByRole("button", { name: "提交标注" })
  if (await submitMarking.isVisible().catch(() => false)) {
    const targetIndexes = await page.evaluate(() => {
      const submit = Array.from(document.querySelectorAll("button")).find((button) =>
        button.textContent?.includes("提交标注"),
      )
      const root = submit?.closest(".space-y-4") ?? document.querySelector("main")
      const spans = Array.from(
        root?.querySelectorAll<HTMLElement>("span[class*='cursor-pointer']") ?? [],
      )
      const source = spans.map((span) => span.textContent ?? "").join("")
      const targets = [
        "接受了目标公司提供的免费豪华酒店住宿",
        "未在报告中披露此信息",
        "原材料成本持续上涨",
        "对部分过时产品进行了减值处理",
        "减值损失计入利润表",
      ]
      const indexes: number[] = []

      for (const target of targets) {
        const start = source.indexOf(target)
        if (start < 0) continue
        for (let index = start; index < start + Array.from(target).length; index += 1) {
          indexes.push(index)
        }
      }

      return indexes.length > 0 ? indexes : spans.slice(0, 8).map((_, index) => index)
    })
    expect(targetIndexes.length, "highlight selectable character count").toBeGreaterThan(0)
    const selectableText = page.locator("main span[class*='cursor-pointer']")
    for (const index of targetIndexes) {
      await selectableText.nth(index).click()
    }
    return
  }

  const formulaButtons = page
    .locator("div")
    .filter({ hasText: "提交公式" })
    .getByRole("button")
    .filter({ hasNotText: /提交公式|返回|下一题|完成单元/ })
  if (await formulaButtons.first().isVisible().catch(() => false)) {
    await formulaButtons.first().click()
    return
  }

  const choiceButtons = page
    .locator("div")
    .filter({ hasText: /提交答案|提交实验结果/ })
    .getByRole("button")
    .filter({ hasNotText: /提交答案|提交实验结果|返回|下一题|完成单元/ })
  if (await choiceButtons.first().isVisible().catch(() => false)) {
    await choiceButtons.first().click()
  }
}
