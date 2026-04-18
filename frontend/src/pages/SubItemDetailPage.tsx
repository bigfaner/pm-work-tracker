import { useState, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTeamStore } from '@/store/team'
import { getMainItemApi } from '@/api/mainItems'
import { getSubItemApi } from '@/api/subItems'
import { listProgressApi, appendProgressApi } from '@/api/progress'
import { listMembersApi } from '@/api/teams'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogBody,
  DialogFooter,
} from '@/components/ui/dialog'
import { Card, CardHeader, CardContent } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import StatusBadge from '@/components/shared/StatusBadge'
import PriorityBadge from '@/components/shared/PriorityBadge'
import UserAvatar from '@/components/shared/UserAvatar'

// --- Main Component ---

export default function SubItemDetailPage() {
  const { mainItemId, subItemId } = useParams<{ mainItemId: string; subItemId: string }>()
  const teamId = useTeamStore((s) => s.currentTeamId)
  const qc = useQueryClient()
  const mId = Number(mainItemId)
  const sId = Number(subItemId)

  // State
  const [appendOpen, setAppendOpen] = useState(false)
  const [appendForm, setAppendForm] = useState({ completion: '', achievement: '', blocker: '' })

  // --- Data fetching ---

  const { data: subItem, isLoading: subLoading } = useQuery({
    queryKey: ['subItem', teamId, sId],
    queryFn: () => getSubItemApi(teamId!, sId),
    enabled: !!teamId && !!sId,
  })

  const { data: mainItem } = useQuery({
    queryKey: ['mainItem', teamId, mId],
    queryFn: () => getMainItemApi(teamId!, mId),
    enabled: !!teamId && !!mId,
  })

  const { data: progressRecords } = useQuery({
    queryKey: ['progress', teamId, sId],
    queryFn: () => listProgressApi(teamId!, sId),
    enabled: !!teamId && !!sId,
  })

  const { data: members } = useQuery({
    queryKey: ['members', teamId],
    queryFn: () => listMembersApi(teamId!),
    enabled: !!teamId,
  })

  const memberName = useCallback(
    (assigneeId: number | null) => {
      if (!assigneeId || !members) return '-'
      const m = members.find((m) => m.userId === assigneeId)
      return m ? m.displayName : 'Unknown'
    },
    [members],
  )

  // Last completion value for validation
  const lastCompletion = progressRecords && progressRecords.length > 0
    ? progressRecords[progressRecords.length - 1].completion
    : 0

  // --- Mutations ---

  const appendMutation = useMutation({
    mutationFn: (req: { completion: number; achievement?: string; blocker?: string }) =>
      appendProgressApi(teamId!, sId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['progress', teamId, sId] })
      qc.invalidateQueries({ queryKey: ['subItem', teamId, sId] })
      setAppendOpen(false)
      setAppendForm({ completion: '', achievement: '', blocker: '' })
    },
  })

  // --- Handlers ---

  const handleAppend = useCallback(() => {
    const val = Number(appendForm.completion)
    if (isNaN(val) || val < 0 || val > 100) return
    if (val < lastCompletion) return
    appendMutation.mutate({
      completion: val,
      ...(appendForm.achievement && { achievement: appendForm.achievement }),
      ...(appendForm.blocker && { blocker: appendForm.blocker }),
    })
  }, [appendForm, lastCompletion, appendMutation])

  // Sorted progress records: reverse chronological
  const sortedRecords = [...(progressRecords || [])].reverse()

  // Sub item code
  const subCode = subItem ? `SI-${String(mId).padStart(3, '0')}-${String(sId).slice(-2)}` : ''

  // --- Render ---

  if (!teamId) return <div className="p-6 text-tertiary">请先选择团队</div>

  const isLoading = subLoading

  return (
    <div data-testid="sub-item-detail-page">
      {isLoading ? (
        <div className="py-8 text-center text-tertiary text-sm">加载中...</div>
      ) : !subItem ? (
        <div className="py-8 text-center text-tertiary text-sm">子事项不存在</div>
      ) : (
        <>
          {/* Breadcrumb */}
          <Breadcrumb className="mb-4">
            <BreadcrumbItem href="/items">事项清单</BreadcrumbItem>
            <BreadcrumbSeparator />
            <BreadcrumbItem href={`/items/${mId}`}>
              {mainItem?.title || `主事项 #${mId}`}
            </BreadcrumbItem>
            <BreadcrumbSeparator />
            <BreadcrumbItem isCurrent>{subItem.title}</BreadcrumbItem>
          </Breadcrumb>

          {/* Title Bar */}
          <div className="flex items-center justify-between mb-5">
            <h1 className="text-xl font-semibold text-primary m-0">{subItem.title}</h1>
            <Button onClick={() => setAppendOpen(true)}>追加进度</Button>
          </div>

          {/* Info Card (4-col grid) */}
          <Card className="mb-5">
            <CardContent>
              <div className="grid grid-cols-4 gap-4">
                <div>
                  <div className="text-xs text-tertiary mb-1">编号</div>
                  <span className="font-mono text-xs bg-bg-alt px-1.5 py-0.5 rounded">{subCode}</span>
                </div>
                <div>
                  <div className="text-xs text-tertiary mb-1">所属主事项</div>
                  <Link to={`/items/${mId}`} className="text-[13px] font-medium text-primary hover:text-primary-600">
                    {mainItem?.title || `主事项 #${mId}`}
                  </Link>
                </div>
                <div>
                  <div className="text-xs text-tertiary mb-1">优先级</div>
                  <PriorityBadge priority={subItem.priority} />
                </div>
                <div>
                  <div className="text-xs text-tertiary mb-1">负责人</div>
                  <div className="flex items-center gap-2">
                    <UserAvatar name={memberName(subItem.assignee_id)} size="sm" />
                    <span className="text-[13px] font-medium">{memberName(subItem.assignee_id)}</span>
                  </div>
                </div>
                <div>
                  <div className="text-xs text-tertiary mb-1">状态</div>
                  <StatusBadge status={subItem.status} />
                </div>
                <div>
                  <div className="text-xs text-tertiary mb-1">预期完成时间</div>
                  <span className="text-[13px]">{subItem.expected_end_date ? subItem.expected_end_date.slice(5) : '-'}</span>
                </div>
                <div>
                  <div className="text-xs text-tertiary mb-1">当前完成度</div>
                  <span className="text-[13px] font-semibold">{subItem.completion}%</span>
                </div>
                <div className="col-span-4">
                  <div className="text-xs text-tertiary mb-1">描述</div>
                  <span className="text-[13px] text-secondary leading-relaxed">{subItem.description || '暂无描述'}</span>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Progress Bar */}
          <Card className="mb-5">
            <CardContent>
              <div className="flex items-center gap-3">
                <span className="text-[13px] text-secondary whitespace-nowrap">总进度</span>
                <Progress value={subItem.completion} className="flex-1" />
                <span className="text-[13px] font-semibold">{subItem.completion}%</span>
              </div>
            </CardContent>
          </Card>

          {/* Progress Timeline */}
          <Card>
            <CardHeader>
              <h3 className="text-sm font-semibold text-primary m-0">进度记录</h3>
            </CardHeader>
            <CardContent>
              {sortedRecords.length === 0 ? (
                <div className="text-tertiary text-sm">暂无进度记录</div>
              ) : (
                <div className="relative pl-6 border-l-2 border-border space-y-5">
                  {sortedRecords.map((record) => {
                    const date = new Date(record.created_at)
                    const dateStr = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
                    return (
                      <div key={record.id} className="relative">
                        {/* Timeline dot */}
                        <div className="absolute -left-[25px] top-1 w-2.5 h-2.5 rounded-full bg-primary-500 border-2 border-white" />
                        <div className="mb-1">
                          <span className="text-xs text-tertiary">{dateStr}</span>
                          <span className="text-xs text-tertiary ml-2">{record.completion}%</span>
                        </div>
                        <div className="text-[13px] text-secondary">
                          {record.achievement && (
                            <div className="mt-1">
                              <strong className="text-emerald-600">成果：</strong>{record.achievement}
                            </div>
                          )}
                          {record.blocker && (
                            <div className="mt-1">
                              <strong className="text-red-600">卡点：</strong>{record.blocker}
                            </div>
                          )}
                        </div>
                      </div>
                    )
                  })}
                </div>
              )}
            </CardContent>
          </Card>

          {/* Append Progress Dialog */}
          <Dialog open={appendOpen} onOpenChange={setAppendOpen}>
            <DialogContent size="md">
              <DialogHeader>
                <DialogTitle>追加进度</DialogTitle>
              </DialogHeader>
              <DialogBody>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-primary mb-1">
                    完成度 <span className="text-error">*</span>
                  </label>
                  <Input
                    type="number"
                    min={0}
                    max={100}
                    placeholder="请输入 0-100 的整数"
                    value={appendForm.completion}
                    onChange={(e) => setAppendForm((f) => ({ ...f, completion: e.target.value }))}
                  />
                  <div className="text-xs text-tertiary mt-1">
                    不能低于上一条记录的完成度（当前：{lastCompletion}%）
                  </div>
                </div>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-primary mb-1">成果</label>
                  <textarea
                    className="flex w-full rounded-md border border-border-dark bg-white px-3 py-2 text-[13px] text-primary shadow-sm placeholder:text-tertiary focus:outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-200"
                    rows={3}
                    placeholder="描述本次取得的成果"
                    value={appendForm.achievement}
                    onChange={(e) => setAppendForm((f) => ({ ...f, achievement: e.target.value }))}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-primary mb-1">卡点</label>
                  <textarea
                    className="flex w-full rounded-md border border-border-dark bg-white px-3 py-2 text-[13px] text-primary shadow-sm placeholder:text-tertiary focus:outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-200"
                    rows={3}
                    placeholder="描述遇到的问题或阻碍"
                    value={appendForm.blocker}
                    onChange={(e) => setAppendForm((f) => ({ ...f, blocker: e.target.value }))}
                  />
                </div>
              </DialogBody>
              <DialogFooter>
                <Button variant="secondary" onClick={() => setAppendOpen(false)}>取消</Button>
                <Button
                  onClick={handleAppend}
                  disabled={!appendForm.completion || appendMutation.isPending}
                >
                  提交
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </>
      )}
    </div>
  )
}
