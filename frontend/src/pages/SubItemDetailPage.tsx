import { useState, useMemo, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Breadcrumb,
  Card,
  Descriptions,
  Tag,
  Progress,
  Timeline,
  Button,
  Modal,
  Form,
  InputNumber,
  Input,
  Tooltip,
  Skeleton,
  Empty,
  message,
} from 'antd'
import { CheckOutlined, CloseOutlined, EditOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import { useAuthStore } from '@/store/auth'
import { useTeamStore } from '@/store/team'
import { getSubItemApi } from '@/api/subItems'
import { getMainItemApi } from '@/api/mainItems'
import { listProgressApi, appendProgressApi, correctCompletionApi } from '@/api/progress'
import { listMembersApi } from '@/api/teams'
import type { SubItem, MainItem, ProgressRecord, TeamMemberResp } from '@/types'

const { TextArea } = Input

const PRIORITY_COLOR_MAP: Record<string, string> = {
  P1: 'orange',
  P2: 'blue',
  P3: 'default',
}

const STATUS_TAG_COLOR_MAP: Record<string, string> = {
  '未开始': 'default',
  '进行中': 'processing',
  '待评审': 'warning',
  '已完成': 'success',
  '已关闭': 'default',
  '阻塞中': 'error',
  '延期': 'orange',
  '归档': 'default',
}

function progressStrokeColor(status: string): string {
  if (status === '已完成') return '#52c41a'
  if (status === '阻塞中') return '#ff4d4f'
  if (status === '延期') return '#fa8c16'
  return '#1677ff'
}

interface AppendFormValues {
  completion: number
  achievement?: string
  blocker?: string
  lesson?: string
}

export default function SubItemDetailPage() {
  const { mainItemId, subItemId } = useParams<{ mainItemId: string; subItemId: string }>()
  const { user } = useAuthStore()
  const { currentTeamId, teams } = useTeamStore()
  const queryClient = useQueryClient()
  const mId = Number(mainItemId)
  const sId = Number(subItemId)

  const [appendModalOpen, setAppendModalOpen] = useState(false)
  const [editingRecordId, setEditingRecordId] = useState<number | null>(null)
  const [editValue, setEditValue] = useState<number>(0)
  const [appendForm] = Form.useForm<AppendFormValues>()
  const [completionError, setCompletionError] = useState<string | null>(null)

  const currentTeam = useMemo(
    () => teams.find((t) => t.id === currentTeamId),
    [teams, currentTeamId],
  )
  const isPM = useMemo(
    () => !!user && !!currentTeam && user.id === currentTeam.pm_id,
    [user, currentTeam],
  )

  // Fetch sub-item detail
  const { data: subItem, isLoading: loadingSub } = useQuery({
    queryKey: ['subItem', currentTeamId, sId],
    queryFn: () => getSubItemApi(currentTeamId!, sId),
    enabled: !!currentTeamId && !!sId,
  })

  // Fetch main item for breadcrumb link
  const { data: mainItem } = useQuery({
    queryKey: ['mainItem', currentTeamId, mId],
    queryFn: () => getMainItemApi(currentTeamId!, mId),
    enabled: !!currentTeamId && !!mId,
  })

  // Fetch members for display names
  const { data: membersData } = useQuery({
    queryKey: ['teamMembers', currentTeamId],
    queryFn: () => listMembersApi(currentTeamId!),
    enabled: !!currentTeamId,
  })

  const memberNameMap = useMemo(() => {
    const map = new Map<number, string>()
    for (const m of membersData ?? []) {
      map.set(m.userId, m.displayName)
    }
    return map
  }, [membersData])

  // Fetch progress records
  const { data: progressRecords = [], isLoading: loadingProgress } = useQuery({
    queryKey: ['progress', currentTeamId, sId],
    queryFn: () => listProgressApi(currentTeamId!, sId),
    enabled: !!currentTeamId && !!sId,
  })

  const sortedRecords = useMemo(
    () => [...progressRecords].sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime()),
    [progressRecords],
  )

  const maxCompletion = useMemo(() => {
    if (sortedRecords.length === 0) return 0
    return Math.max(...sortedRecords.map((r) => r.completion))
  }, [sortedRecords])

  const isAssignee = useMemo(
    () => !!user && !!subItem && user.id === subItem.assignee_id,
    [user, subItem],
  )

  const canAppend = isPM || isAssignee

  const isOverdue = useCallback((item: SubItem): boolean => {
    if (!item.expected_end_date) return false
    if (item.status === '已完成' || item.status === '已关闭') return false
    return dayjs().isAfter(dayjs(item.expected_end_date))
  }, [])

  // --- Handlers ---

  const handleOpenAppend = useCallback(() => {
    appendForm.resetFields()
    setCompletionError(null)
    setAppendModalOpen(true)
  }, [appendForm])

  const handleAppendSubmit = useCallback(async () => {
    try {
      const values = await appendForm.validateFields()
      if (values.completion < maxCompletion) {
        setCompletionError(`完成度不能低于上一次记录的最大值 ${maxCompletion}%`)
        return
      }
      setCompletionError(null)
      await appendProgressApi(currentTeamId!, sId, {
        completion: values.completion,
        achievement: values.achievement,
        blocker: values.blocker,
        lesson: values.lesson,
      })
      message.success('追加进度成功')
      setAppendModalOpen(false)
      queryClient.invalidateQueries({ queryKey: ['progress', currentTeamId, sId] })
      queryClient.invalidateQueries({ queryKey: ['subItem', currentTeamId, sId] })
    } catch {
      // validation errors handled by form
    }
  }, [appendForm, currentTeamId, sId, maxCompletion, queryClient])

  const handleStartPMEdit = useCallback((record: ProgressRecord) => {
    setEditingRecordId(record.id)
    setEditValue(record.completion)
  }, [])

  const handleConfirmPMEdit = useCallback(async (recordId: number) => {
    try {
      await correctCompletionApi(currentTeamId!, recordId, { completion: editValue })
      message.success('修正完成度成功')
      setEditingRecordId(null)
      queryClient.invalidateQueries({ queryKey: ['progress', currentTeamId, sId] })
      queryClient.invalidateQueries({ queryKey: ['subItem', currentTeamId, sId] })
    } catch {
      message.error('修正失败')
    }
  }, [currentTeamId, editValue, sId, queryClient])

  const handleCancelPMEdit = useCallback(() => {
    setEditingRecordId(null)
  }, [])

  // --- Render ---

  if (loadingSub) {
    return (
      <div data-testid="sub-item-detail-page">
        <div data-testid="detail-skeleton">
          <Skeleton active paragraph={{ rows: 6 }} />
        </div>
      </div>
    )
  }

  if (!subItem) return null

  const overdue = isOverdue(subItem)

  return (
    <div data-testid="sub-item-detail-page">
      {/* Breadcrumb */}
      <Breadcrumb
        data-testid="breadcrumb"
        items={[
          { title: <Link to="/items">事项视图</Link> },
          { title: <Link to={`/items/${mId}`} data-testid="breadcrumb-main-link">{mainItem?.title ?? '...'}</Link> },
          { title: subItem.title },
        ]}
        style={{ marginBottom: 16 }}
      />

      {/* Info Card */}
      <Card data-testid="info-card" style={{ marginBottom: 24 }}>
        <Descriptions column={3} bordered size="middle">
          <Descriptions.Item label="编号">
            <span style={{ fontFamily: 'monospace' }}>SI-{String(subItem.id).padStart(4, '0')}</span>
          </Descriptions.Item>
          <Descriptions.Item label="所属主事项">
            <Link to={`/items/${mId}`} data-testid="main-item-link">
              {mainItem?.title ?? '...'}
            </Link>
          </Descriptions.Item>
          <Descriptions.Item label="优先级">
            <Tag color={PRIORITY_COLOR_MAP[subItem.priority]}>{subItem.priority}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="负责人">
            {subItem.assignee_id ? (memberNameMap.get(subItem.assignee_id) || '-') : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="状态">
            <Tag color={STATUS_TAG_COLOR_MAP[subItem.status]}>{subItem.status}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="预期完成时间">
            {subItem.expected_end_date ? (
              overdue ? (
                <Tooltip title="已超期">
                  <span data-testid="overdue-date" style={{ color: '#ff4d4f' }}>
                    {dayjs(subItem.expected_end_date).format('YYYY-MM-DD')}
                  </span>
                </Tooltip>
              ) : (
                <span>{dayjs(subItem.expected_end_date).format('YYYY-MM-DD')}</span>
              )
            ) : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="当前完成度">
            <Progress type="line" percent={subItem.completion} size="small" style={{ width: 150 }} />
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {/* Progress Summary Bar */}
      <Card title="完成度概览" style={{ marginBottom: 24 }}>
        <div data-testid="progress-summary">
          <Progress
            type="line"
            percent={subItem.completion}
            strokeColor={progressStrokeColor(subItem.status)}
            style={{ width: '100%' }}
          />
        </div>
      </Card>

      {/* Progress Timeline */}
      <Card
        title="进度记录"
        extra={
          canAppend && (
            <Button
              type="primary"
              data-testid="append-progress-btn"
              onClick={handleOpenAppend}
            >
              追加进度
            </Button>
          )
        }
      >
        {loadingProgress ? (
          <Skeleton active paragraph={{ rows: 3 }} />
        ) : sortedRecords.length === 0 ? (
          <div data-testid="timeline-empty">
            <Empty description="暂无进度记录" image={Empty.PRESENTED_IMAGE_SIMPLE} />
          </div>
        ) : (
          <div data-testid="progress-timeline">
            <Timeline
              mode="left"
              items={sortedRecords.map((record) => {
                const timeLabel = dayjs(record.created_at).format('MM-DD HH:mm')
                const authorName = memberNameMap.get(record.author_id) || `User#${record.author_id}`

                return {
                  key: record.id,
                  label: (
                    <span>
                      {timeLabel} {authorName}
                    </span>
                  ),
                  children: (
                    <div data-testid={`timeline-item-${record.id}`}>
                      {/* Completion row with inline PM edit */}
                      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
                        <span style={{ fontWeight: 500 }}>完成度:</span>
                        {editingRecordId === record.id ? (
                          <>
                            <InputNumber
                              data-testid={`pm-edit-input-${record.id}`}
                              min={0}
                              max={100}
                              value={editValue}
                              onChange={(v) => setEditValue(v ?? 0)}
                              size="small"
                              style={{ width: 80 }}
                            />
                            <CheckOutlined
                              data-testid={`pm-edit-confirm-${record.id}`}
                              style={{ color: '#52c41a', cursor: 'pointer' }}
                              onClick={() => handleConfirmPMEdit(record.id)}
                            />
                            <CloseOutlined
                              data-testid={`pm-edit-cancel-${record.id}`}
                              style={{ color: '#ff4d4f', cursor: 'pointer' }}
                              onClick={handleCancelPMEdit}
                            />
                          </>
                        ) : (
                          <>
                            <div data-testid={`timeline-progress-${record.id}`} style={{ width: 200 }}>
                              <Progress type="line" percent={record.completion} size="small" />
                            </div>
                            {isPM && (
                              <EditOutlined
                                data-testid={`pm-edit-trigger-${record.id}`}
                                style={{ color: '#1677ff', cursor: 'pointer' }}
                                onClick={() => handleStartPMEdit(record)}
                              />
                            )}
                          </>
                        )}
                        {record.is_pm_correct && (
                          <Tag color="orange" style={{ marginLeft: 4 }}>PM已修正</Tag>
                        )}
                      </div>
                      {/* Text blocks */}
                      {record.achievement && (
                        <div style={{ marginBottom: 4 }}>
                          <span style={{ fontWeight: 500, color: '#52c41a' }}>成果: </span>
                          <span>{record.achievement}</span>
                        </div>
                      )}
                      {record.blocker && (
                        <div style={{ marginBottom: 4 }}>
                          <span style={{ fontWeight: 500, color: '#ff4d4f' }}>卡点: </span>
                          <span>{record.blocker}</span>
                        </div>
                      )}
                      {record.lesson && (
                        <div style={{ marginBottom: 4 }}>
                          <span style={{ fontWeight: 500, color: '#1677ff' }}>经验: </span>
                          <span>{record.lesson}</span>
                        </div>
                      )}
                    </div>
                  ),
                }
              })}
            />
          </div>
        )}
      </Card>

      {/* Append Progress Modal */}
      <Modal
        title="追加进度"
        open={appendModalOpen}
        onCancel={() => setAppendModalOpen(false)}
        onOk={handleAppendSubmit}
        width={480}
        data-testid="append-modal"
        okText="确认"
        cancelText="取消"
        okButtonProps={{ 'data-testid': 'append-modal-submit' }}
      >
        <Form form={appendForm} layout="vertical">
          <Form.Item
            name="completion"
            label="完成度"
            rules={[{ required: true, message: '请输入完成度' }]}
            validateStatus={completionError ? 'error' : undefined}
            help={
              completionError ? (
                <span data-testid="completion-error">{completionError}</span>
              ) : (
                <span data-testid="completion-helper-text">当前最大完成度: {maxCompletion}%</span>
              )
            }
          >
            <InputNumber
              data-testid="append-form-completion"
              min={0}
              max={100}
              style={{ width: '100%' }}
              placeholder="请输入完成度 (0-100)"
            />
          </Form.Item>
          <Form.Item name="achievement" label="成果">
            <TextArea data-testid="append-form-achievement" rows={3} placeholder="请输入成果" />
          </Form.Item>
          <Form.Item name="blocker" label="卡点">
            <TextArea data-testid="append-form-blocker" rows={3} placeholder="请输入卡点" />
          </Form.Item>
          <Form.Item name="lesson" label="经验">
            <TextArea data-testid="append-form-lesson" rows={3} placeholder="请输入经验" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
