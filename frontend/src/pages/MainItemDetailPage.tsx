import { useState, useMemo, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Breadcrumb,
  Card,
  Descriptions,
  Tag,
  Progress,
  Select,
  Button,
  Table,
  Tooltip,
  Skeleton,
  Empty,
  Modal,
  Form,
  Input,
  DatePicker,
  message,
} from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import { useAuthStore } from '@/store/auth'
import { useTeamStore } from '@/store/team'
import { getMainItemApi } from '@/api/mainItems'
import {
  listSubItemsApi,
  createSubItemApi,
  updateSubItemApi,
  changeSubItemStatusApi,
  assignSubItemApi,
} from '@/api/subItems'
import { listMembersApi } from '@/api/teams'
import type { MainItem, SubItem, SubItemFilter, TeamMemberResp } from '@/types'

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

const STATUS_OPTIONS = ['未开始', '进行中', '待评审', '已完成', '已关闭', '阻塞中', '延期']

function isOverdue(item: MainItem): boolean {
  if (!item.expected_end_date) return false
  if (item.status === '已完成' || item.status === '已关闭') return false
  return dayjs().isAfter(dayjs(item.expected_end_date))
}

function overdueDays(item: MainItem): number {
  if (!item.expected_end_date) return 0
  return dayjs().diff(dayjs(item.expected_end_date), 'day')
}

interface SubItemFormValues {
  title: string
  description?: string
  priority: string
  assigneeId: number
  startDate?: dayjs.Dayjs
  expectedEndDate?: dayjs.Dayjs
}

export default function MainItemDetailPage() {
  const { mainItemId } = useParams<{ mainItemId: string }>()
  const { user } = useAuthStore()
  const { currentTeamId, teams } = useTeamStore()
  const queryClient = useQueryClient()
  const itemId = Number(mainItemId)

  const [subFilters, setSubFilters] = useState<SubItemFilter>({})
  const [modalOpen, setModalOpen] = useState(false)
  const [editingSub, setEditingSub] = useState<SubItem | null>(null)
  const [statusModalOpen, setStatusModalOpen] = useState(false)
  const [statusTarget, setStatusTarget] = useState<SubItem | null>(null)
  const [assignModalOpen, setAssignModalOpen] = useState(false)
  const [assignTarget, setAssignTarget] = useState<SubItem | null>(null)
  const [form] = Form.useForm<SubItemFormValues>()
  const [statusForm] = Form.useForm<{ status: string }>()
  const [assignForm] = Form.useForm<{ assigneeId: number }>()

  const currentTeam = useMemo(
    () => teams.find((t) => t.id === currentTeamId),
    [teams, currentTeamId],
  )
  const isPM = useMemo(
    () => !!user && !!currentTeam && user.id === currentTeam.pm_id,
    [user, currentTeam],
  )

  // Fetch main item detail
  const { data: mainItem, isLoading: loadingMain } = useQuery({
    queryKey: ['mainItem', currentTeamId, itemId],
    queryFn: () => getMainItemApi(currentTeamId!, itemId),
    enabled: !!currentTeamId && !!itemId,
  })

  // Fetch members for assignee select
  const { data: membersData } = useQuery({
    queryKey: ['teamMembers', currentTeamId],
    queryFn: () => listMembersApi(currentTeamId!),
    enabled: !!currentTeamId,
  })
  const memberOptions = useMemo(
    () => (membersData ?? []).map((m: TeamMemberResp) => ({ value: m.userId, label: m.displayName })),
    [membersData],
  )
  const memberNameMap = useMemo(() => {
    const map = new Map<number, string>()
    for (const m of membersData ?? []) {
      map.set(m.userId, m.displayName)
    }
    return map
  }, [membersData])

  // Fetch sub-items with filters
  const { data: subData, isLoading: loadingSubs } = useQuery({
    queryKey: ['subItems', currentTeamId, itemId, subFilters],
    queryFn: () => listSubItemsApi(currentTeamId!, itemId, subFilters),
    enabled: !!currentTeamId && !!itemId,
  })

  const subItems = subData?.items ?? []

  const overdue = mainItem ? isOverdue(mainItem) : false

  // --- Handlers ---

  const handleFilterChange = useCallback(
    (key: keyof SubItemFilter, value: string | number | undefined) => {
      setSubFilters((prev) => ({ ...prev, [key]: value }))
    },
    [],
  )

  const handleResetFilters = useCallback(() => {
    setSubFilters({})
  }, [])

  const handleCreateSubItem = useCallback(() => {
    setEditingSub(null)
    form.resetFields()
    setModalOpen(true)
  }, [form])

  const handleEditSubItem = useCallback(
    (sub: SubItem) => {
      setEditingSub(sub)
      form.setFieldsValue({
        title: sub.title,
        description: sub.description || undefined,
        priority: sub.priority,
        assigneeId: sub.assignee_id ?? undefined,
        startDate: sub.start_date ? dayjs(sub.start_date) : undefined,
        expectedEndDate: sub.expected_end_date ? dayjs(sub.expected_end_date) : undefined,
      })
      setModalOpen(true)
    },
    [form],
  )

  const handleModalSubmit = useCallback(async () => {
    try {
      const values = await form.validateFields()
      const req = {
        title: values.title,
        description: values.description,
        priority: values.priority,
        assigneeId: values.assigneeId,
        startDate: values.startDate?.format('YYYY-MM-DD'),
        expectedEndDate: values.expectedEndDate?.format('YYYY-MM-DD'),
      }

      if (editingSub) {
        await updateSubItemApi(currentTeamId!, editingSub.id, {
          title: req.title,
          description: req.description,
          priority: req.priority,
          expectedEndDate: req.expectedEndDate,
        })
        message.success('更新成功')
      } else {
        await createSubItemApi(currentTeamId!, itemId, req)
        message.success('创建成功')
      }

      setModalOpen(false)
      queryClient.invalidateQueries({ queryKey: ['subItems', currentTeamId, itemId] })
      queryClient.invalidateQueries({ queryKey: ['mainItem', currentTeamId, itemId] })
    } catch {
      // validation errors handled by form
    }
  }, [form, editingSub, currentTeamId, itemId, queryClient])

  const handleStatusChange = useCallback(
    (sub: SubItem) => {
      setStatusTarget(sub)
      statusForm.resetFields()
      setStatusModalOpen(true)
    },
    [statusForm],
  )

  const handleStatusSubmit = useCallback(async () => {
    if (!statusTarget) return
    try {
      const values = await statusForm.validateFields()
      await changeSubItemStatusApi(currentTeamId!, statusTarget.id, { status: values.status })
      message.success('状态变更成功')
      setStatusModalOpen(false)
      queryClient.invalidateQueries({ queryKey: ['subItems', currentTeamId, itemId] })
      queryClient.invalidateQueries({ queryKey: ['mainItem', currentTeamId, itemId] })
    } catch {
      // validation errors handled by form
    }
  }, [statusTarget, statusForm, currentTeamId, itemId, queryClient])

  const handleAssign = useCallback(
    (sub: SubItem) => {
      setAssignTarget(sub)
      assignForm.setFieldsValue({ assigneeId: sub.assignee_id ?? undefined })
      setAssignModalOpen(true)
    },
    [assignForm],
  )

  const handleAssignSubmit = useCallback(async () => {
    if (!assignTarget) return
    try {
      const values = await assignForm.validateFields()
      await assignSubItemApi(currentTeamId!, assignTarget.id, { assigneeId: values.assigneeId })
      message.success('分配成功')
      setAssignModalOpen(false)
      queryClient.invalidateQueries({ queryKey: ['subItems', currentTeamId, itemId] })
    } catch {
      // validation errors handled by form
    }
  }, [assignTarget, assignForm, currentTeamId, itemId, queryClient])

  // --- Render ---

  if (loadingMain) {
    return (
      <div data-testid="main-item-detail-page">
        <div data-testid="detail-skeleton">
          <Skeleton active paragraph={{ rows: 6 }} />
        </div>
      </div>
    )
  }

  if (!mainItem) return null

  return (
    <div data-testid="main-item-detail-page">
      {/* Breadcrumb */}
      <Breadcrumb
        data-testid="breadcrumb"
        items={[
          { title: <Link to="/items">事项视图</Link> },
          { title: mainItem.title },
        ]}
        style={{ marginBottom: 16 }}
      />

      {/* Detail Header Card */}
      <Card data-testid="detail-header" style={{ marginBottom: 24 }}>
        <Descriptions column={3} size="middle">
          <Descriptions.Item label="编码">
            <Tag style={{ fontFamily: 'monospace' }}>{mainItem.code}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="标题">
            <span style={{ fontWeight: 500 }}>{mainItem.title}</span>
          </Descriptions.Item>
          <Descriptions.Item label="优先级">
            <Tag color={PRIORITY_COLOR_MAP[mainItem.priority]}>{mainItem.priority}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="提出人">
            {memberNameMap.get(mainItem.proposer_id) || '-'}
          </Descriptions.Item>
          <Descriptions.Item label="负责人">
            {mainItem.assignee_id ? (memberNameMap.get(mainItem.assignee_id) || '-') : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="状态">
            <Tag color={STATUS_TAG_COLOR_MAP[mainItem.status]}>{mainItem.status}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="开始时间">
            {mainItem.start_date ? dayjs(mainItem.start_date).format('YYYY-MM-DD') : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="预期完成时间">
            {mainItem.expected_end_date ? (
              overdue ? (
                <Tooltip title={`已超期 ${overdueDays(mainItem)} 天`}>
                  <span data-testid="overdue-date" style={{ color: '#ff4d4f' }}>
                    {dayjs(mainItem.expected_end_date).format('YYYY-MM-DD')}
                  </span>
                </Tooltip>
              ) : (
                <span>{dayjs(mainItem.expected_end_date).format('YYYY-MM-DD')}</span>
              )
            ) : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="完成度">
            <div data-testid="header-progress">
              <Progress type="line" percent={mainItem.completion} size="small" style={{ width: 150 }} />
            </div>
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {/* Sub-item Section */}
      <Card
        title="子事项"
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            data-testid="create-sub-item-btn"
            onClick={handleCreateSubItem}
          >
            新建子事项
          </Button>
        }
      >
        {/* Sub-item Filter Bar */}
        <div style={{ display: 'flex', gap: 12, marginBottom: 16 }}>
          <Select
            data-testid="sub-filter-priority"
            placeholder="优先级"
            allowClear
            style={{ width: 120 }}
            value={subFilters.priority}
            onChange={(v) => handleFilterChange('priority', v)}
            options={[
              { value: 'P1', label: 'P1' },
              { value: 'P2', label: 'P2' },
              { value: 'P3', label: 'P3' },
            ]}
          />
          <Select
            data-testid="sub-filter-status"
            placeholder="状态"
            allowClear
            style={{ width: 140 }}
            value={subFilters.status}
            onChange={(v) => handleFilterChange('status', v)}
            options={STATUS_OPTIONS.map((s) => ({ value: s, label: s }))}
          />
          <Select
            data-testid="sub-filter-assignee"
            placeholder="负责人"
            allowClear
            showSearch
            style={{ width: 140 }}
            value={subFilters.assigneeId}
            onChange={(v) => handleFilterChange('assigneeId', v)}
            options={memberOptions}
          />
          <Button type="link" onClick={handleResetFilters}>
            重置
          </Button>
        </div>

        {/* Sub-item Table */}
        {loadingSubs ? (
          <Skeleton active paragraph={{ rows: 3 }} />
        ) : subItems.length === 0 ? (
          <div data-testid="sub-items-empty">
            <Empty description="暂无子事项" image={Empty.PRESENTED_IMAGE_SIMPLE} />
          </div>
        ) : (
          <Table
            dataSource={subItems}
            rowKey="id"
            pagination={false}
            size="middle"
            columns={[
              {
                title: '标题',
                dataIndex: 'title',
                key: 'title',
                render: (title: string, record: SubItem) => (
                  <a data-testid={`sub-item-link-${record.id}`} href={`/items/${record.main_item_id}/sub/${record.id}`}>
                    {title}
                  </a>
                ),
              },
              {
                title: '优先级',
                dataIndex: 'priority',
                key: 'priority',
                width: 80,
                render: (p: string) => <Tag color={PRIORITY_COLOR_MAP[p]}>{p}</Tag>,
              },
              {
                title: '负责人',
                dataIndex: 'assignee_id',
                key: 'assignee',
                width: 100,
                render: (id: number | null) => (id ? memberNameMap.get(id) || '-' : '-'),
              },
              {
                title: '完成度',
                dataIndex: 'completion',
                key: 'completion',
                width: 120,
                render: (c: number) => <Progress type="line" percent={c} size="small" />,
              },
              {
                title: '状态',
                dataIndex: 'status',
                key: 'status',
                width: 100,
                render: (s: string) => <Tag color={STATUS_TAG_COLOR_MAP[s]}>{s}</Tag>,
              },
              {
                title: '预期完成时间',
                dataIndex: 'expected_end_date',
                key: 'expected_end_date',
                width: 130,
                render: (d: string | null) => (d ? dayjs(d).format('YYYY-MM-DD') : '-'),
              },
              {
                title: '操作',
                key: 'actions',
                width: 200,
                render: (_: unknown, record: SubItem) => {
                  const canEdit = isPM || (user?.id === record.assignee_id)
                  return (
                    <div style={{ display: 'flex', gap: 8 }}>
                      {canEdit && (
                        <Button
                          type="link"
                          size="small"
                          data-testid={`sub-item-edit-${record.id}`}
                          onClick={() => handleEditSubItem(record)}
                        >
                          编辑
                        </Button>
                      )}
                      {canEdit && (
                        <Button
                          type="link"
                          size="small"
                          data-testid={`sub-item-status-change-${record.id}`}
                          onClick={() => handleStatusChange(record)}
                        >
                          变更状态
                        </Button>
                      )}
                      {isPM && (
                        <Button
                          type="link"
                          size="small"
                          data-testid={`sub-item-assign-${record.id}`}
                          onClick={() => handleAssign(record)}
                        >
                          分配负责人
                        </Button>
                      )}
                    </div>
                  )
                },
              },
            ]}
          />
        )}
      </Card>

      {/* Create/Edit Sub-item Modal */}
      <Modal
        title={editingSub ? '编辑子事项' : '新建子事项'}
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        onOk={handleModalSubmit}
        width={520}
        data-testid="sub-item-modal"
        okText="确认"
        cancelText="取消"
        okButtonProps={{ 'data-testid': 'sub-modal-submit-btn' }}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="title"
            label="标题"
            rules={[{ required: true, message: '请输入标题' }]}
          >
            <Input data-testid="sub-form-title" maxLength={100} showCount />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <TextArea data-testid="sub-form-description" rows={3} />
          </Form.Item>
          <Form.Item
            name="priority"
            label="优先级"
            rules={[{ required: true, message: '请选择优先级' }]}
          >
            <Select
              data-testid="sub-form-priority"
              options={[
                { value: 'P1', label: 'P1' },
                { value: 'P2', label: 'P2' },
                { value: 'P3', label: 'P3' },
              ]}
            />
          </Form.Item>
          <Form.Item
            name="assigneeId"
            label="负责人"
            rules={[{ required: true, message: '请选择负责人' }]}
          >
            <Select
              data-testid="sub-form-assignee"
              showSearch
              options={memberOptions}
            />
          </Form.Item>
          <Form.Item name="startDate" label="开始时间">
            <DatePicker data-testid="sub-form-start-date" style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="expectedEndDate" label="预期完成时间">
            <DatePicker data-testid="sub-form-expected-end-date" style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>

      {/* Status Change Modal */}
      <Modal
        title="变更状态"
        open={statusModalOpen}
        onCancel={() => setStatusModalOpen(false)}
        onOk={handleStatusSubmit}
        width={400}
        data-testid="status-change-modal"
        okText="确认"
        cancelText="取消"
      >
        <Form form={statusForm} layout="vertical">
          <Form.Item
            name="status"
            label="新状态"
            rules={[{ required: true, message: '请选择状态' }]}
          >
            <Select
              data-testid="status-change-select"
              options={STATUS_OPTIONS.map((s) => ({ value: s, label: s }))}
            />
          </Form.Item>
        </Form>
      </Modal>

      {/* Assignee Modal */}
      <Modal
        title="分配负责人"
        open={assignModalOpen}
        onCancel={() => setAssignModalOpen(false)}
        onOk={handleAssignSubmit}
        width={400}
        data-testid="assign-modal"
        okText="确认"
        cancelText="取消"
      >
        <Form form={assignForm} layout="vertical">
          <Form.Item
            name="assigneeId"
            label="负责人"
            rules={[{ required: true, message: '请选择负责人' }]}
          >
            <Select
              data-testid="assign-select"
              showSearch
              options={memberOptions}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
