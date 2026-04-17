import { useState, useMemo, useCallback } from 'react'
import { useQuery, useQueryClient, useMutation } from '@tanstack/react-query'
import {
  Radio,
  Card,
  Tag,
  Button,
  Modal,
  Form,
  Input,
  Select,
  Empty,
  Skeleton,
  Typography,
  message,
} from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'
import { useAuthStore } from '@/store/auth'
import { useTeamStore } from '@/store/team'
import {
  listItemPoolApi,
  submitItemPoolApi,
  assignItemPoolApi,
  rejectItemPoolApi,
} from '@/api/itemPool'
import { listMembersApi } from '@/api/teams'
import { listMainItemsApi } from '@/api/mainItems'
import type { ItemPool, ItemPoolFilter, SubmitItemPoolReq, AssignItemPoolReq, RejectItemPoolReq } from '@/types'

dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

const { TextArea } = Input
const { Paragraph } = Typography

const POOL_STATUS_TAG_COLOR: Record<string, string> = {
  '待分配': 'blue',
  '已分配': 'green',
  '已拒绝': 'red',
}

interface SubmitFormValues {
  title: string
  background?: string
  expectedOutput?: string
}

interface AssignFormValues {
  mainItemId: number
  assigneeId: number
}

interface RejectFormValues {
  reason: string
}

interface PoolItemCardProps {
  item: ItemPool
  isPM: boolean
  memberNames: Record<number, string>
  mainItemOptions: { value: number; label: string; title: string }[]
  onAssign: (item: ItemPool) => void
  onReject: (item: ItemPool) => void
}

function PoolItemCard({ item, isPM, memberNames, mainItemOptions, onAssign, onReject }: PoolItemCardProps) {
  const isPending = item.status === '待分配'
  const isAssigned = item.status === '已分配'
  const isRejected = item.status === '已拒绝'
  const grayed = isAssigned || isRejected

  const assignedMainTitle = useMemo(() => {
    if (!item.assigned_main_id) return ''
    const m = mainItemOptions.find((o) => o.value === item.assigned_main_id)
    return m?.title ?? ''
  }, [item.assigned_main_id, mainItemOptions])

  return (
    <Card
      data-testid={`pool-card-${item.id}`}
      hoverable
      size="small"
      style={{
        marginBottom: 12,
        opacity: grayed ? 0.7 : 1,
        borderLeft: isPending ? '3px solid #1677ff' : 'none',
      }}
    >
      {/* Header: title + status tag + submitter + time */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
        <span style={{ fontWeight: 500, flex: 1 }}>{item.title}</span>
        <Tag color={POOL_STATUS_TAG_COLOR[item.status]}>{item.status}</Tag>
        <span style={{ color: '#8c8c8c', fontSize: 12 }}>
          {memberNames[item.submitter_id] ?? ''} {dayjs(item.created_at).fromNow()}
        </span>
      </div>

      {/* Body: background + expected output */}
      {(item.background || item.expected_output) && (
        <div style={{ marginBottom: 8 }}>
          {item.background && (
            <div style={{ marginBottom: 4 }}>
              <span style={{ color: '#8c8c8c', fontSize: 12 }}>背景: </span>
              <Paragraph
                ellipsis={{ rows: 2, expandable: true, symbol: '展开' }}
                style={{ display: 'inline', marginBottom: 0 }}
              >
                {item.background}
              </Paragraph>
            </div>
          )}
          {item.expected_output && (
            <div>
              <span style={{ color: '#8c8c8c', fontSize: 12 }}>预期产出: </span>
              <Paragraph
                ellipsis={{ rows: 2, expandable: true, symbol: '展开' }}
                style={{ display: 'inline', marginBottom: 0 }}
              >
                {item.expected_output}
              </Paragraph>
            </div>
          )}
        </div>
      )}

      {/* Footer by status */}
      {isPending && isPM && (
        <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
          <Button
            type="primary"
            size="small"
            data-testid={`assign-btn-${item.id}`}
            onClick={() => onAssign(item)}
          >
            分配
          </Button>
          <Button
            danger
            size="small"
            data-testid={`reject-btn-${item.id}`}
            onClick={() => onReject(item)}
          >
            拒绝
          </Button>
        </div>
      )}
      {isAssigned && (
        <div style={{ color: '#8c8c8c', fontSize: 12 }}>
          已挂载至:
          {item.assigned_main_id ? (
            <a href={`/items/${item.assigned_main_id}`} style={{ marginLeft: 4 }}>
              {assignedMainTitle}
            </a>
          ) : (
            <span style={{ marginLeft: 4 }}>-</span>
          )}
          {item.assignee_id ? (
            <span style={{ marginLeft: 8 }}>负责人: {memberNames[item.assignee_id] ?? ''}</span>
          ) : null}
        </div>
      )}
      {isRejected && item.reject_reason && (
        <div style={{ color: '#8c8c8c', fontSize: 12 }}>
          拒绝原因: {item.reject_reason}
        </div>
      )}
    </Card>
  )
}

export default function ItemPoolPage() {
  const { user } = useAuthStore()
  const { currentTeamId, teams } = useTeamStore()
  const queryClient = useQueryClient()

  const [statusFilter, setStatusFilter] = useState<string | undefined>(undefined)
  const [submitModalOpen, setSubmitModalOpen] = useState(false)
  const [assignModalOpen, setAssignModalOpen] = useState(false)
  const [rejectModalOpen, setRejectModalOpen] = useState(false)
  const [selectedItem, setSelectedItem] = useState<ItemPool | null>(null)

  const [submitForm] = Form.useForm<SubmitFormValues>()
  const [assignForm] = Form.useForm<AssignFormValues>()
  const [rejectForm] = Form.useForm<RejectFormValues>()

  const currentTeam = useMemo(
    () => teams.find((t) => t.id === currentTeamId),
    [teams, currentTeamId],
  )
  const isPM = useMemo(
    () => !!user && !!currentTeam && user.id === currentTeam.pm_id,
    [user, currentTeam],
  )

  // Fetch members
  const { data: membersData } = useQuery({
    queryKey: ['teamMembers', currentTeamId],
    queryFn: () => listMembersApi(currentTeamId!),
    enabled: !!currentTeamId,
  })
  const memberNames = useMemo(() => {
    const map: Record<number, string> = {}
    for (const m of membersData ?? []) {
      map[m.userId] = m.displayName
    }
    return map
  }, [membersData])

  const memberOptions = useMemo(
    () => (membersData ?? []).map((m) => ({ value: m.userId, label: m.displayName })),
    [membersData],
  )

  // Fetch pool items
  const filter: ItemPoolFilter = useMemo(() => {
    const f: ItemPoolFilter = {}
    if (statusFilter) f.status = statusFilter
    return f
  }, [statusFilter])

  const { data, isLoading } = useQuery({
    queryKey: ['itemPool', currentTeamId, filter],
    queryFn: () => listItemPoolApi(currentTeamId!, filter),
    enabled: !!currentTeamId,
  })

  // Fetch main items for assign modal
  const { data: mainItemsData } = useQuery({
    queryKey: ['mainItems', currentTeamId, { archived: false }],
    queryFn: () => listMainItemsApi(currentTeamId!, { archived: false }),
    enabled: !!currentTeamId,
  })
  const mainItemOptions = useMemo(
    () => (mainItemsData?.items ?? []).map((m) => ({
      value: m.id,
      label: `${m.code} ${m.title}`,
      title: m.title,
    })),
    [mainItemsData],
  )

  const items = data?.items ?? []

  // Mutations
  const submitMutation = useMutation({
    mutationFn: (req: SubmitItemPoolReq) => submitItemPoolApi(currentTeamId!, req),
    onSuccess: () => {
      message.success('已提交到事项池')
      setSubmitModalOpen(false)
      submitForm.resetFields()
      queryClient.invalidateQueries({ queryKey: ['itemPool', currentTeamId] })
    },
  })

  const assignMutation = useMutation({
    mutationFn: ({ poolId, req }: { poolId: number; req: AssignItemPoolReq }) =>
      assignItemPoolApi(currentTeamId!, poolId, req),
    onSuccess: () => {
      message.success('已分配')
      setAssignModalOpen(false)
      assignForm.resetFields()
      setSelectedItem(null)
      queryClient.invalidateQueries({ queryKey: ['itemPool', currentTeamId] })
    },
  })

  const rejectMutation = useMutation({
    mutationFn: ({ poolId, req }: { poolId: number; req: RejectItemPoolReq }) =>
      rejectItemPoolApi(currentTeamId!, poolId, req),
    onSuccess: () => {
      message.success('已拒绝')
      setRejectModalOpen(false)
      rejectForm.resetFields()
      setSelectedItem(null)
      queryClient.invalidateQueries({ queryKey: ['itemPool', currentTeamId] })
    },
  })

  // Handlers
  const handleOpenAssign = useCallback((item: ItemPool) => {
    setSelectedItem(item)
    assignForm.resetFields()
    setAssignModalOpen(true)
  }, [assignForm])

  const handleOpenReject = useCallback((item: ItemPool) => {
    setSelectedItem(item)
    rejectForm.resetFields()
    setRejectModalOpen(true)
  }, [rejectForm])

  const handleSubmitOk = useCallback(async () => {
    try {
      const values = await submitForm.validateFields()
      submitMutation.mutate({
        title: values.title,
        background: values.background,
        expectedOutput: values.expectedOutput,
      })
    } catch {
      // validation errors
    }
  }, [submitForm, submitMutation])

  const handleAssignOk = useCallback(async () => {
    try {
      const values = await assignForm.validateFields()
      assignMutation.mutate({
        poolId: selectedItem!.id,
        req: { mainItemId: values.mainItemId, assigneeId: values.assigneeId },
      })
    } catch {
      // validation errors
    }
  }, [assignForm, assignMutation, selectedItem])

  const handleRejectOk = useCallback(async () => {
    try {
      const values = await rejectForm.validateFields()
      rejectMutation.mutate({
        poolId: selectedItem!.id,
        req: { reason: values.reason },
      })
    } catch {
      // validation errors
    }
  }, [rejectForm, rejectMutation, selectedItem])

  return (
    <div data-testid="item-pool-page">
      {/* Page Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <h2 style={{ margin: 0 }}>事项池</h2>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          data-testid="submit-pool-btn"
          onClick={() => {
            submitForm.resetFields()
            setSubmitModalOpen(true)
          }}
        >
          提交到事项池
        </Button>
      </div>

      {/* Filter Bar */}
      <div style={{ marginBottom: 16 }}>
        <Radio.Group
          data-testid="status-filter"
          buttonStyle="solid"
          value={statusFilter ?? '全部'}
          onChange={(e) => {
            setStatusFilter(e.target.value === '全部' ? undefined : e.target.value)
          }}
        >
          <Radio.Button value="全部">全部</Radio.Button>
          <Radio.Button value="待分配">待分配</Radio.Button>
          <Radio.Button value="已分配">已分配</Radio.Button>
          <Radio.Button value="已拒绝">已拒绝</Radio.Button>
        </Radio.Group>
      </div>

      {/* Content */}
      {isLoading ? (
        <div data-testid="pool-skeleton">
          {[1, 2, 3].map((i) => (
            <Card key={i} size="small" style={{ marginBottom: 12 }}>
              <Skeleton active paragraph={{ rows: 2 }} />
            </Card>
          ))}
        </div>
      ) : items.length === 0 && !statusFilter ? (
        <div data-testid="pool-empty">
          <Empty description="事项池暂无内容" />
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => {
              submitForm.resetFields()
              setSubmitModalOpen(true)
            }}
          >
            提交到事项池
          </Button>
        </div>
      ) : items.length === 0 && statusFilter ? (
        <div data-testid="pool-filter-empty">
          <Empty description="该状态下暂无事项" />
        </div>
      ) : (
        items.map((item) => (
          <PoolItemCard
            key={item.id}
            item={item}
            isPM={isPM}
            memberNames={memberNames}
            mainItemOptions={mainItemOptions}
            onAssign={handleOpenAssign}
            onReject={handleOpenReject}
          />
        ))
      )}

      {/* Submit Modal */}
      <Modal
        title="提交到事项池"
        open={submitModalOpen}
        onCancel={() => setSubmitModalOpen(false)}
        onOk={handleSubmitOk}
        width={520}
        data-testid="submit-pool-modal"
        okText="确认"
        cancelText="取消"
        okButtonProps={{ 'data-testid': 'submit-pool-ok', loading: submitMutation.isPending }}
      >
        <Form form={submitForm} layout="vertical">
          <Form.Item
            name="title"
            label="标题"
            rules={[{ required: true, message: '请填写标题' }]}
          >
            <Input data-testid="form-pool-title" maxLength={100} showCount />
          </Form.Item>
          <Form.Item name="background" label="背景">
            <TextArea data-testid="form-pool-background" rows={3} />
          </Form.Item>
          <Form.Item name="expectedOutput" label="预期产出">
            <TextArea data-testid="form-pool-expected-output" rows={3} />
          </Form.Item>
        </Form>
      </Modal>

      {/* Assign Modal */}
      <Modal
        title="分配事项"
        open={assignModalOpen}
        onCancel={() => { setAssignModalOpen(false); setSelectedItem(null) }}
        onOk={handleAssignOk}
        width={480}
        data-testid="assign-modal"
        okText="确认"
        cancelText="取消"
        okButtonProps={{ 'data-testid': 'assign-ok', loading: assignMutation.isPending }}
      >
        <Form form={assignForm} layout="vertical">
          <Form.Item
            name="mainItemId"
            label="挂载主事项"
            rules={[{ required: true, message: '请选择挂载的主事项' }]}
          >
            <Select
              data-testid="form-assign-main-item"
              showSearch
              placeholder="请选择主事项"
              options={mainItemOptions}
            />
          </Form.Item>
          <Form.Item
            name="assigneeId"
            label="负责人"
            rules={[{ required: true, message: '请选择负责人' }]}
          >
            <Select
              data-testid="form-assign-assignee"
              showSearch
              placeholder="请选择负责人"
              options={memberOptions}
            />
          </Form.Item>
        </Form>
      </Modal>

      {/* Reject Modal */}
      <Modal
        title="拒绝事项"
        open={rejectModalOpen}
        onCancel={() => { setRejectModalOpen(false); setSelectedItem(null) }}
        onOk={handleRejectOk}
        width={400}
        data-testid="reject-modal"
        okText="确认"
        cancelText="取消"
        okButtonProps={{ 'data-testid': 'reject-ok', loading: rejectMutation.isPending }}
      >
        <Form form={rejectForm} layout="vertical">
          <Form.Item
            name="reason"
            label="拒绝原因"
            rules={[{ required: true, message: '请填写拒绝原因' }]}
          >
            <TextArea data-testid="form-reject-reason" rows={3} maxLength={200} showCount />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
