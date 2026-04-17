import { useState, useMemo, useCallback } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Collapse,
  Select,
  Button,
  Tag,
  Progress,
  Avatar,
  Badge,
  Tooltip,
  Dropdown,
  Empty,
  Skeleton,
  Modal,
  Form,
  Input,
  DatePicker,
  message,
} from 'antd'
import { PlusOutlined, EllipsisOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import { useAuthStore } from '@/store/auth'
import { useTeamStore } from '@/store/team'
import { listMainItemsApi, createMainItemApi, updateMainItemApi, archiveMainItemApi } from '@/api/mainItems'
import { listSubItemsApi } from '@/api/subItems'
import { listMembersApi } from '@/api/teams'
import type { MainItem, SubItem, MainItemFilter } from '@/types'

const { TextArea } = Input

const STATUS_OPTIONS = ['未开始', '进行中', '待评审', '已完成', '已关闭', '阻塞中', '延期', '归档']

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

function isOverdue(item: MainItem): boolean {
  if (!item.expected_end_date) return false
  if (item.status === '已完成' || item.status === '已关闭') return false
  return dayjs().isAfter(dayjs(item.expected_end_date))
}

function overdueDays(item: MainItem): number {
  if (!item.expected_end_date) return 0
  return dayjs().diff(dayjs(item.expected_end_date), 'day')
}

function canArchive(item: MainItem): boolean {
  return item.status === '已完成' || item.status === '已关闭'
}

interface SubItemRowProps {
  sub: SubItem
  teamId: number
}

function SubItemRow({ sub, teamId }: SubItemRowProps) {
  return (
    <div
      data-testid={`sub-item-row-${sub.id}`}
      style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '8px 0' }}
    >
      <span style={{ flex: 1 }}>
        <a data-testid={`sub-item-link-${sub.id}`} href={`/items/${sub.main_item_id}/sub/${sub.id}`}>
          {sub.title}
        </a>
      </span>
      <Progress type="line" percent={sub.completion} size="small" style={{ width: 120 }} />
      <Tag color={STATUS_TAG_COLOR_MAP[sub.status] || 'default'}>{sub.status}</Tag>
      {sub.expected_end_date && (
        <span>{dayjs(sub.expected_end_date).format('YYYY-MM-DD')}</span>
      )}
    </div>
  )
}

interface MainItemPanelHeaderProps {
  item: MainItem
  isPM: boolean
  members: { value: number; label: string }[]
  onEdit: (item: MainItem) => void
  onArchive: (item: MainItem) => void
}

function MainItemPanelHeader({ item, isPM, members, onEdit, onArchive }: MainItemPanelHeaderProps) {
  const overdue = isOverdue(item)
  const isP1 = item.priority === 'P1'

  const menuItems = isPM ? [
    { key: 'edit', label: '编辑' },
    { key: 'archive', label: '归档' },
  ] : []

  const handleMenuClick = ({ key }: { key: string }) => {
    if (key === 'edit') {
      onEdit(item)
    } else if (key === 'archive') {
      onArchive(item)
    }
  }

  const assigneeName = useMemo(() => {
    if (!item.assignee_id) return ''
    const m = members.find((m) => m.value === item.assignee_id)
    return m?.label ?? ''
  }, [item.assignee_id, members])

  return (
    <div
      data-testid={`panel-header-${item.id}`}
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        width: '100%',
        borderLeft: overdue ? '3px solid #ff4d4f' : 'none',
        paddingLeft: overdue ? 8 : 0,
      }}
    >
      <Tag style={{ fontFamily: 'monospace', width: 80, textAlign: 'center' }}>{item.code}</Tag>
      {isP1 && (
        <span data-testid={`p1-badge-${item.id}`}>
          <Badge color="#fa8c16" />
        </span>
      )}
      <span style={{ flex: 1, fontWeight: 500 }}>{item.title}</span>
      <Tag color={PRIORITY_COLOR_MAP[item.priority]}>{item.priority}</Tag>
      {assigneeName && (
        <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
          <Avatar size={24}>{assigneeName.charAt(0)}</Avatar>
          <span>{assigneeName}</span>
        </span>
      )}
      <Progress type="line" percent={item.completion} size="small" style={{ width: 120 }} />
      <Tag color={STATUS_TAG_COLOR_MAP[item.status]}>{item.status}</Tag>
      {item.expected_end_date && (
        overdue ? (
          <Tooltip title={`已超期 ${overdueDays(item)} 天`}>
            <span data-testid={`overdue-date-${item.id}`} style={{ color: '#ff4d4f' }}>
              {dayjs(item.expected_end_date).format('YYYY-MM-DD')}
            </span>
          </Tooltip>
        ) : (
          <span>{dayjs(item.expected_end_date).format('YYYY-MM-DD')}</span>
        )
      )}
      {isPM && (
        <Dropdown
          menu={{ items: menuItems, onClick: handleMenuClick }}
          trigger={['click']}
        >
          <Button
            type="text"
            size="small"
            icon={<EllipsisOutlined />}
            data-testid={`actions-dropdown-${item.id}`}
            onClick={(e) => e.stopPropagation()}
          />
        </Dropdown>
      )}
    </div>
  )
}

interface SubItemListProps {
  teamId: number
  mainItemId: number
}

export function SubItemList({ teamId, mainItemId }: SubItemListProps) {
  const { data, isLoading } = useQuery({
    queryKey: ['subItems', teamId, mainItemId],
    queryFn: () => listSubItemsApi(teamId, mainItemId),
  })

  if (isLoading) {
    return <Skeleton active paragraph={{ rows: 2 }} />
  }

  const subs = data?.items ?? []
  if (subs.length === 0) {
    return <Empty description="暂无子事项" image={Empty.PRESENTED_IMAGE_SIMPLE} />
  }

  return (
    <div data-testid={`sub-item-list-${mainItemId}`}>
      {subs.map((sub) => (
        <SubItemRow key={sub.id} sub={sub} teamId={teamId} />
      ))}
    </div>
  )
}

interface ModalFormValues {
  title: string
  priority: string
  assigneeId?: number
  expectedEndDate?: dayjs.Dayjs
  description?: string
}

export default function ItemViewPage() {
  const { user } = useAuthStore()
  const { currentTeamId, teams } = useTeamStore()
  const queryClient = useQueryClient()

  const [filters, setFilters] = useState<MainItemFilter>({ archived: false })
  const [modalOpen, setModalOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<MainItem | null>(null)
  const [expandedKeys, setExpandedKeys] = useState<string[]>([])
  const [form] = Form.useForm<ModalFormValues>()

  const currentTeam = useMemo(
    () => teams.find((t) => t.id === currentTeamId),
    [teams, currentTeamId],
  )
  const isPM = useMemo(
    () => !!user && !!currentTeam && user.id === currentTeam.pm_id,
    [user, currentTeam],
  )

  // Fetch members for assignee select
  const { data: membersData } = useQuery({
    queryKey: ['teamMembers', currentTeamId],
    queryFn: () => listMembersApi(currentTeamId!),
    enabled: !!currentTeamId,
  })
  const memberOptions = useMemo(
    () => (membersData ?? []).map((m) => ({ value: m.userId, label: m.displayName })),
    [membersData],
  )

  // Fetch main items
  const { data, isLoading } = useQuery({
    queryKey: ['mainItems', currentTeamId, filters],
    queryFn: () => listMainItemsApi(currentTeamId!, filters),
    enabled: !!currentTeamId,
  })

  const items = data?.items ?? []
  const hasActiveFilters = !!(filters.priority || filters.status || filters.assigneeId)

  const handleFilterChange = useCallback(
    (key: keyof MainItemFilter, value: string | number | undefined) => {
      setFilters((prev) => ({ ...prev, [key]: value }))
    },
    [],
  )

  const handleReset = useCallback(() => {
    setFilters({ archived: false })
  }, [])

  const handleCreate = useCallback(() => {
    setEditingItem(null)
    form.resetFields()
    setModalOpen(true)
  }, [form])

  const handleEdit = useCallback(
    (item: MainItem) => {
      setEditingItem(item)
      form.setFieldsValue({
        title: item.title,
        priority: item.priority,
        assigneeId: item.assignee_id ?? undefined,
        expectedEndDate: item.expected_end_date ? dayjs(item.expected_end_date) : undefined,
      })
      setModalOpen(true)
    },
    [form],
  )

  const handleArchive = useCallback(
    (item: MainItem) => {
      if (!canArchive(item)) {
        message.warning('请先完成或关闭事项再归档')
        return
      }
      Modal.confirm({
        title: '确认归档？',
        onOk: async () => {
          await archiveMainItemApi(currentTeamId!, item.id)
          message.success('已归档')
          queryClient.invalidateQueries({ queryKey: ['mainItems', currentTeamId] })
        },
      })
    },
    [currentTeamId, queryClient],
  )

  const handleModalSubmit = useCallback(async () => {
    try {
      const values = await form.validateFields()
      const req = {
        title: values.title,
        priority: values.priority,
        assigneeId: values.assigneeId,
        expectedEndDate: values.expectedEndDate?.format('YYYY-MM-DD'),
      }

      if (editingItem) {
        await updateMainItemApi(currentTeamId!, editingItem.id, {
          title: req.title,
          priority: req.priority,
          assigneeId: req.assigneeId,
          expectedEndDate: req.expectedEndDate,
        })
        message.success('更新成功')
      } else {
        await createMainItemApi(currentTeamId!, req)
        message.success('创建成功')
      }

      setModalOpen(false)
      queryClient.invalidateQueries({ queryKey: ['mainItems', currentTeamId] })
    } catch {
      // validation errors handled by form
    }
  }, [form, editingItem, currentTeamId, queryClient])

  return (
    <div data-testid="item-view-page">
      {/* Page Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <h2 style={{ margin: 0 }}>事项视图</h2>
        {isPM && (
          <Button
            type="primary"
            icon={<PlusOutlined />}
            data-testid="create-main-item-btn"
            onClick={handleCreate}
          >
            新建主事项
          </Button>
        )}
      </div>

      {/* Filter Bar */}
      <div style={{ display: 'flex', gap: 12, marginBottom: 16 }}>
        <Select
          data-testid="filter-priority"
          placeholder="优先级"
          allowClear
          style={{ width: 120 }}
          value={filters.priority}
          onChange={(v) => handleFilterChange('priority', v)}
          options={[
            { value: 'P1', label: 'P1' },
            { value: 'P2', label: 'P2' },
            { value: 'P3', label: 'P3' },
          ]}
        />
        <Select
          data-testid="filter-status"
          placeholder="状态"
          allowClear
          style={{ width: 140 }}
          value={filters.status}
          onChange={(v) => handleFilterChange('status', v)}
          options={STATUS_OPTIONS.map((s) => ({ value: s, label: s }))}
        />
        <Select
          data-testid="filter-assignee"
          placeholder="负责人"
          allowClear
          showSearch
          style={{ width: 140 }}
          value={filters.assigneeId}
          onChange={(v) => handleFilterChange('assigneeId', v)}
          options={memberOptions}
        />
        <Button
          type="link"
          data-testid="filter-reset"
          onClick={handleReset}
        >
          重置
        </Button>
      </div>

      {/* Content */}
      {isLoading ? (
        <div data-testid="main-items-skeleton">
          <Skeleton active paragraph={{ rows: 3 }} />
        </div>
      ) : items.length === 0 && !hasActiveFilters ? (
        <div data-testid="empty-state">
          <Empty description="暂无事项" />
          {isPM && (
            <Button
              type="primary"
              icon={<PlusOutlined />}
              data-testid="empty-create-btn"
              onClick={handleCreate}
            >
              新建主事项
            </Button>
          )}
        </div>
      ) : items.length === 0 && hasActiveFilters ? (
        <div data-testid="filter-empty-state">
          <Empty description="没有符合条件的事项" />
          <Button
            type="primary"
            data-testid="filter-reset-btn"
            onClick={handleReset}
          >
            重置筛选
          </Button>
        </div>
      ) : (
        <Collapse
          activeKey={expandedKeys}
          onChange={(keys) => setExpandedKeys(keys as string[])}
          items={items.map((item) => ({
            key: String(item.id),
            label: (
              <MainItemPanelHeader
                item={item}
                isPM={isPM}
                members={memberOptions}
                onEdit={handleEdit}
                onArchive={handleArchive}
              />
            ),
            children: expandedKeys.includes(String(item.id)) ? (
              <SubItemList teamId={currentTeamId!} mainItemId={item.id} />
            ) : null,
          }))}
        />
      )}

      {/* Create/Edit Modal */}
      <Modal
        title={editingItem ? '编辑主事项' : '新建主事项'}
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        onOk={handleModalSubmit}
        width={520}
        data-testid="main-item-modal"
        okText="确认"
        cancelText="取消"
        okButtonProps={{ 'data-testid': 'modal-submit-btn' }}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="title"
            label="标题"
            rules={[{ required: true, message: '请输入标题' }]}
          >
            <Input data-testid="form-title" maxLength={100} showCount />
          </Form.Item>
          <Form.Item
            name="priority"
            label="优先级"
            rules={[{ required: true, message: '请选择优先级' }]}
          >
            <Select
              data-testid="form-priority"
              options={[
                { value: 'P1', label: 'P1' },
                { value: 'P2', label: 'P2' },
                { value: 'P3', label: 'P3' },
              ]}
            />
          </Form.Item>
          <Form.Item name="assigneeId" label="负责人">
            <Select
              data-testid="form-assignee"
              showSearch
              options={memberOptions}
            />
          </Form.Item>
          <Form.Item name="expectedEndDate" label="预期完成时间">
            <DatePicker data-testid="form-expected-end-date" style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <TextArea data-testid="form-description" rows={3} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
