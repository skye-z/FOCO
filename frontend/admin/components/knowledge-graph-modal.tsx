"use client"

import * as React from "react"
import cytoscape from "cytoscape"
import { LoaderCircle } from "lucide-react"
import { toast } from "sonner"

import { authFetch, readBrowserAccessToken } from "@/lib/auth-session"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"

type GraphNode = {
  id: string
  type: string
  ref_id: string
  label: string
  description?: string
  group?: string
}

type GraphEdge = {
  id: string
  source: string
  target: string
  type: string
  label?: string
}

type GraphPayload = {
  nodes: GraphNode[]
  edges: GraphEdge[]
}

type ExamTreeNode = {
  id: string
  name: string
  type: "exam" | "subject" | "chapter"
}

export function KnowledgeGraphModal({
  open,
  onClose,
}: {
  open: boolean
  onClose: () => void
}) {
  const containerRef = React.useRef<HTMLDivElement | null>(null)
  const graphRef = React.useRef<cytoscape.Core | null>(null)
  const [containerReady, setContainerReady] = React.useState(0)
  const [loading, setLoading] = React.useState(false)

  const setContainerNode = React.useCallback((node: HTMLDivElement | null) => {
    containerRef.current = node
    if (node) {
      setContainerReady((value) => value + 1)
    }
  }, [])

  React.useEffect(() => {
    if (!open || !containerRef.current) {
      return
    }

    let disposed = false

    async function loadGraph() {
      const token = readBrowserAccessToken()
      if (!token) return

      setLoading(true)
      try {
        const response = await authFetch("/api/v1/admin/knowledge-graph", {
          headers: { Authorization: `Bearer ${token}` },
          cache: "no-store",
        })
        if (!response.ok) throw new Error("graph load failed")
        const payload = (await response.json()).data as GraphPayload
        if (disposed || !containerRef.current) return

        graphRef.current?.destroy()
        const filteredNodes = payload.nodes
        const nodeIdSet = new Set(filteredNodes.map((node) => node.id))
        const filteredEdges = payload.edges.filter((edge) => nodeIdSet.has(edge.source) && nodeIdSet.has(edge.target))

        graphRef.current = cytoscape({
          container: containerRef.current,
          elements: {
            nodes: filteredNodes.map((node) => ({
              data: {
                id: node.id,
                label: node.label || "未命名节点",
                type: node.type,
                description: node.description || "",
              },
            })),
            edges: filteredEdges.map((edge) => ({
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
              selector: 'node[type = "exam"]',
              style: {
                label: "data(label)",
                shape: "triangle",
                width: 36,
                height: 36,
                "background-color": "#f59e0b",
                color: "#94a3b8",
                "font-size": 12,
                "font-weight": 700,
                "text-wrap": "wrap",
                "text-max-width": "180px",
                "text-valign": "bottom",
                "text-margin-y": 10,
              },
            },
            {
              selector: 'node[type = "knowledge_point"]',
              style: {
                label: "data(label)",
                shape: "star",
                width: 34,
                height: 34,
                "background-color": "#2563eb",
                color: "#94a3b8",
                "font-size": 11,
                "text-wrap": "wrap",
                "text-max-width": "140px",
                "text-valign": "bottom",
                "text-margin-y": 8,
              },
            },
            {
              selector: 'node[type = "subject"]',
              style: {
                label: "data(label)",
                shape: "diamond",
                width: 34,
                height: 34,
                "background-color": "#7c3aed",
                color: "#94a3b8",
                "font-size": 11,
                "font-weight": 600,
                "text-wrap": "wrap",
                "text-max-width": "160px",
                "text-valign": "bottom",
                "text-margin-y": 8,
              },
            },
            {
              selector: 'node[type = "chapter"]',
              style: {
                label: "data(label)",
                shape: "ellipse",
                width: 28,
                height: 28,
                "background-color": "#0f766e",
                color: "#94a3b8",
                "font-size": 10,
                "font-weight": 600,
                "text-wrap": "wrap",
                "text-max-width": "150px",
                "text-valign": "bottom",
                "text-margin-y": 8,
              },
            },
            {
              selector: 'node[type = "question"]',
              style: {
                label: "data(label)",
                shape: "round-rectangle",
                width: 24,
                height: 24,
                "background-color": "#16a34a",
                color: "#94a3b8",
                "font-size": 10,
                "text-wrap": "wrap",
                "text-max-width": "180px",
                "text-valign": "bottom",
                "text-margin-y": 8,
              },
            },
            {
              selector: "node.hovered, node:selected",
              style: {
                color: "#0f172a",
                "z-index": 999,
                "overlay-opacity": 0,
              },
            },
            {
              selector: "edge",
              style: {
                width: 1.5,
                "line-color": "#94a3b8",
                "target-arrow-color": "#94a3b8",
                "target-arrow-shape": "triangle",
                "curve-style": "bezier",
                opacity: 0.7,
              },
            },
            {
              selector: 'edge[type = "question_tag"]',
              style: {
                "line-style": "dashed",
                "target-arrow-shape": "none",
              },
            },
          ],
          layout: {
            name: "cose",
            animate: false,
            fit: true,
            padding: 72,
            idealEdgeLength: 180,
            nodeRepulsion: 120000,
            nodeOverlap: 24,
            componentSpacing: 120,
            gravity: 0.4,
          },
        })

        requestAnimationFrame(() => {
          graphRef.current?.resize()
          graphRef.current?.fit(undefined, 48)
          graphRef.current?.center()
        })

        graphRef.current.on("mouseover", "node", (event) => {
          event.target.addClass("hovered")
        })

        graphRef.current.on("mouseout", "node", (event) => {
          event.target.removeClass("hovered")
        })

        graphRef.current.on("tap", "node", (event) => {
          const node = event.target.data()
          toast.info(node.label, {
            description: node.description || `节点类型：${node.type}`,
          })
        })
      } catch {
        toast.error("知识图谱加载失败", { description: "请稍后重试。" })
      } finally {
        if (!disposed) {
          setLoading(false)
        }
      }
    }

    void loadGraph()
    return () => {
      disposed = true
    }
  }, [containerReady, open])

  React.useEffect(() => {
    if (!open) {
      graphRef.current?.destroy()
      graphRef.current = null
    }
  }, [open])

  React.useEffect(() => {
    if (!open) return

    function handleResize() {
      graphRef.current?.resize()
      graphRef.current?.fit(undefined, 48)
    }

    window.addEventListener("resize", handleResize)
    return () => window.removeEventListener("resize", handleResize)
  }, [open])

  return (
    <Dialog open={open} onOpenChange={(nextOpen) => !nextOpen && onClose()}>
      <DialogContent className="!top-0 !left-0 !h-screen !max-h-screen !w-screen !max-w-none !translate-x-0 !translate-y-0 rounded-none p-6">
        <DialogHeader>
          <DialogTitle>知识图谱</DialogTitle>
        </DialogHeader>
        <div className="flex h-full min-h-0 flex-col">
          <div className="relative flex-1 min-h-0 overflow-hidden rounded-xl border bg-muted/10">
            {loading ? (
              <div className="absolute inset-0 z-10 flex items-center justify-center bg-background/80">
                <LoaderCircle className="size-6 animate-spin text-muted-foreground" />
              </div>
            ) : null}
            <div ref={setContainerNode} className="h-full w-full" />
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
