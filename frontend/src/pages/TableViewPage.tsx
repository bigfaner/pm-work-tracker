import { useState, useMemo, useCallback } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Table,
  Tag,
  Progress,
  Select,
  Button,
  Empty,
  message,
} from 'antd'
import type { TablePaginationConfig } from 'antd'
import type { SorterResult } from 'antd/es/table/interface'
import { DownloadOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import { useNavigate } from 'react-router-dom'
import { useTeamStore } from '@/store/team'
import { getTableViewApi, exportTableCsvApi } from '@/api/views'
import { listMembersApi } from '@/api/teams'
import type { TableRow, TableFilter } from '@/types'

const PRIORITY_TAG_COLOR: Record<string, string> = {
  P1: 'orange',
  P2: 'blue',
  P3: 'default',
}

const STATUS_TAG_COLOR: Record<string, string> = {
  '未开始': 'default',
  '进行中': 'processing',
  '待评审': 'warning',
  '已完成': 'success',
  '已关闭': 'default',
  '阻塞中': 'error',
  '延期': 'orange',
  '归档': 'default',
}

const TYPE_OPTIONS = [
  { value: '', label: '全部' },
  { value: 'main', label: '主事项' },
  { value: 'sub', label: '子事项' },
]

const PRIORITY_OPTIONS = [
  { value: '', label: '全部' },
  { value: 'P1', label: 'P1' },
  { value: 'P2', label: 'P2' },
  { value: 'P3', label: 'P3' },
]

const STATUS_OPTIONS = [
  { value: '', label: '全部' },
  { value: '未开始', label: '未开始' },
  { value: '进行中', label: '进行中' },
  { value: '待评审', label: '待评审' },
  { value: '已完成', label: '已完成' },
  { value: '已关闭', label: '已关闭' },
  { value: '阻塞中', label: '阻塞中' },
  { value: '延期', label: '延期' },
  { value: '归档', label: '归档' },
]

const COMPLETED_STATUSES = new Set(['已完成', '已关闭'])

interface TableRowWithNav extends TableRow {
  navigatePath: string
}

export default function TableViewPage() {
  const navigate = useNavigate()
  const { currentTeamId } = useTeamStore()

  const [typeFilter, setTypeFilter] = useState<string | undefined>(undefined)
  const [priorityFilter, setPriorityFilter] = useState<string | undefined>(undefined)
  const [statusFilter, setStatusFilter] = useState<string | undefined>(undefined)
  const [assigneeFilter, setAssigneeFilter] = useState<number | undefined>(undefined)
  const [sortBy, setSortBy] = useState<string>('priority')
  const [sortOrder, setSortOrder] = useState<string>('desc')
  const [page, setPage] = useState(1)
  const [exporting, setExporting] = useState(false)

  // Fetch members for assignee filter
  const { data: membersData } = useQuery({
    queryKey: ['teamMembers', currentTeamId],
    queryFn: () => listMembersApi(currentTeamId!),
    enabled: !!currentTeamId,
  })

  const assigneeOptions = useMemo(
    () => [
      { value: 0, label: '全部' },
      ...(membersData ?? []).map((m) => ({ value: m.userId, label: m.displayName })),
    ],
    [membersData],
  )

  // Build filter params
  const filter: TableFilter = useMemo(() => {
    const f: TableFilter = {
      sortBy,
      sortOrder,
      page,
      pageSize: 50,
    }
    if (typeFilter) f.type = typeFilter
    if (priorityFilter) f.priority = priorityFilter
    if (statusFilter) f.status = statusFilter
    if (assigneeFilter) f.assigneeId = assigneeFilter
    return f
  }, [typeFilter, priorityFilter, statusFilter, assigneeFilter, sortBy, sortOrder, page])

  // Fetch table data
  const { data, isLoading } = useQuery({
    queryKey: ['tableView', currentTeamId, filter],
    queryFn: () => getTableViewApi(currentTeamId!, filter),
    enabled: !!currentTeamId,
  })

  const items = data?.items ?? []
  const total = data?.total ?? 0

  // Reset filters
  const handleReset = useCallback(() => {
    setTypeFilter(undefined)
    setPriorityFilter(undefined)
    setStatusFilter(undefined)
    setAssigneeFilter(undefined)
    setSortBy('priority')
    setSortOrder('desc')
    setPage(1)
  }, [])

  // CSV export
  const handleExport = useCallback(async () => {
    if (!currentTeamId) return
    setExporting(true)
    try {
      const exportFilter: TableFilter = { ...filter }
      delete exportFilter.page
      delete exportFilter.pageSize
      const blob = await exportTableCsvApi(currentTeamId, exportFilter)
      const url = URL.createObjectURL(new Blob([blob], { type: 'text/csv' }))
      const a = document.createElement('a')
      a.href = url
      a.download = 'items-export.csv'
      a.click()
      URL.revokeObjectURL(url)
      message.success('CSV 已导出')
    } catch (err: unknown) {
      const error = err as { response?: { data?: { code?: string } } }
      if (error?.response?.data?.code === 'NO_DATA') {
        message.error('当前筛选条件下无数据可导出')
      } else {
        message.error('导出失败')
      }
    } finally {
      setExporting(false)
    }
  }, [currentTeamId, filter])

  // Table change handler for sort and pagination
  const handleTableChange = useCallback(
    (_pagination: TablePaginationConfig, _filters: Record<string, unknown>, sorter: SorterResult<TableRow> | SorterResult<TableRow>[]) => {
      const s = Array.isArray(sorter) ? sorter[0] : sorter
      if (s.field && s.order) {
        setSortBy(String(s.field))
        setSortOrder(s.order === 'ascend' ? 'asc' : 'desc')
      } else {
        // Reset to default sort
        setSortBy('priority')
        setSortOrder('desc')
      }
      if (_pagination.current) {
        setPage(_pagination.current)
      }
    },
    [],
  )

  // Row click navigation
  const handleRowClick = useCallback(
    (record: TableRowWithNav) => {
      navigate(record.navigatePath)
    },
    [navigate],
  )

  // Enrich rows with navigation path
  const enrichedItems: TableRowWithNav[] = useMemo(
    () =>
      items.map((item) => ({
        ...item,
        navigatePath: item.type === 'main'
          ? `/items/${item.id}`
          : `/items/${item.id}/sub/${item.id}`,
      })),
    [items],
  )

  // Row class for overdue
  const getRowClassName = useCallback((record: TableRow) => {
    if (
      record.expectedEndDate &&
      !COMPLETED_STATUSES.has(record.status) &&
      dayjs(record.expectedEndDate).isBefore(dayjs(), 'day')
    ) {
      return 'row-overdue'
    }
    return ''
  }, [])

  const columns = [
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      width: 80,
      render: (type: string) => (
        <Tag color={type === 'main' ? 'blue' : 'default'}>
          {type === 'main' ? '主事项' : '子事项'}
        </Tag>
      ),
    },
    {
      title: '编号',
      dataIndex: 'code',
      key: 'code',
      width: 100,
      sorter: true,
      render: (code: string) => <span style={{ fontFamily: 'monospace' }}>{code}</span>,
    },
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      ellipsis: true,
      render: (title: string, record: TableRowWithNav) => (
        <a onClick={() => handleRowClick(record)}>{title}</a>
      ),
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      width: 80,
      sorter: true,
      render: (priority: string) => (
        <Tag color={PRIORITY_TAG_COLOR[priority]}>{priority}</Tag>
      ),
    },
    {
      title: '负责人',
      dataIndex: 'assigneeName',
      key: 'assigneeName',
      width: 120,
    },
    {
      title: '完成度',
      dataIndex: 'completion',
      key: 'completion',
      width: 140,
      sorter: true,
      render: (completion: number) => (
        <Progress percent={completion} size="small" style={{ width: 100 }} />
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={STATUS_TAG_COLOR[status]}>{status}</Tag>
      ),
    },
    {
      title: '预期完成时间',
      dataIndex: 'expectedEndDate',
      key: 'expectedEndDate',
      width: 130,
      sorter: true,
      render: (date: string | null, record: TableRow) => {
        if (!date) return '-'
        const isOverdue =
          !COMPLETED_STATUSES.has(record.status) && dayjs(date).isBefore(dayjs(), 'day')
        return (
          <span style={{ color: isOverdue ? '#ff4d4f' : undefined }}>
            {dayjs(date).format('YYYY-MM-DD')}
          </span>
        )
      },
    },
    {
      title: '实际完成时间',
      dataIndex: 'actualEndDate',
      key: 'actualEndDate',
      width: 130,
      sorter: true,
      render: (date: string | null) => (date ? dayjs(date).format('YYYY-MM-DD') : '-'),
    },
  ]

  return (
    <div data-testid="table-view-page">
      {/* Page Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <h2 style={{ margin: 0 }}>表格视图</h2>
        <Button
          icon={<DownloadOutlined />}
          data-testid="export-csv-btn"
          onClick={handleExport}
          loading={exporting}
          disabled={exporting}
        >
          导出 CSV
        </Button>
      </div>

      {/* Filter Bar */}
      <div style={{ display: 'flex', gap: 12, marginBottom: 16 }}>
        <Select
          data-testid="filter-type"
          placeholder="类型"
          style={{ width: 120 }}
          allowClear
          options={TYPE_OPTIONS}
          value={typeFilter || undefined}
          onChange={(val: string | undefined) => {
            setTypeFilter(val || undefined)
            setPage(1)
          }}
        />
        <Select
          data-testid="filter-priority"
          placeholder="优先级"
          style={{ width: 120 }}
          allowClear
          options={PRIORITY_OPTIONS}
          value={priorityFilter || undefined}
          onChange={(val: string | undefined) => {
            setPriorityFilter(val || undefined)
            setPage(1)
          }}
        />
        <Select
          data-testid="filter-status"
          placeholder="状态"
          style={{ width: 140 }}
          allowClear
          options={STATUS_OPTIONS}
          value={statusFilter || undefined}
          onChange={(val: string | undefined) => {
            setStatusFilter(val || undefined)
            setPage(1)
          }}
        />
        <Select
          data-testid="filter-assignee"
          placeholder="负责人"
          style={{ width: 140 }}
          allowClear
          options={assigneeOptions}
          fieldNames={{ value: 'value', label: 'label' }}
          value={assigneeFilter || undefined}
          onChange={(val: number | undefined) => {
            setAssigneeFilter(val || undefined)
            setPage(1)
          }}
        />
        <Button
          type="link"
          data-testid="filter-reset"
          onClick={handleReset}
        >
          重置
        </Button>
      </div>

      {/* Table */}
      <Table<TableRowWithNav>
        rowKey="id"
        columns={columns}
        dataSource={enrichedItems}
        loading={isLoading}
        onChange={handleTableChange as never}
        onRow={(record) => ({
          onClick: () => handleRowClick(record),
          style: { cursor: 'pointer' },
        })}
        rowClassName={getRowClassName}
        pagination={{
          pageSize: 50,
          current: page,
          total,
          showTotal: (t) => `共 ${t} 条`,
          onChange: (p) => setPage(p),
        }}
        locale={{
          emptyText: (
            <div data-testid="table-empty">
              <Empty description="暂无事项" />
            </div>
          ),
        }}
      />
    </div>
  )
}
